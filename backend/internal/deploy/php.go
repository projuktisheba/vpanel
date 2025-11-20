package deploy

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	aptTimeout     = 10 * time.Minute
	cmdTimeout     = 2 * time.Minute
	phpFpmProcWait = 2 * time.Second
)

var (
	phpPoolTpl = `[{{.Domain}}]
user = {{.SysUser}}
group = {{.SysUser}}
listen = {{.Socket}}
listen.owner = www-data
listen.group = www-data
pm = dynamic
pm.max_children = 10
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 5
chdir = {{.ProjectPath}}
`
	nginxTpl = `server {
    listen 80;
    server_name {{.Domain}} www.{{.Domain}};

    root {{.WebRoot}};
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:{{.Socket}};
        fastcgi_buffers 16 16k;
        fastcgi_buffer_size 32k;
    }

    location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
        expires 6M;
        add_header Cache-Control "public";
    }

    location ~ /\.ht {
        deny all;
    }
}
`
)

type deployParams struct {
	Domain      string
	SysUser     string
	BasePath    string
	ProjectPath string
	TargetPHP   string // "7.4", "8.1" etc
	Socket      string
	WebRoot     string
	Framework   string
}

// run runs a command with timeout, returns combined output on success or error with output.
func run(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

// runInteractive returns combined output too, but allows longer timeout.
func runInteractive(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func detectFramework(projectPath string) (string, error) {
	// detect Laravel, WordPress, CodeIgniter3, fallback to php
	if fileExists(filepath.Join(projectPath, "artisan")) {
		return "laravel", nil
	}
	if fileExists(filepath.Join(projectPath, "wp-config.php")) {
		return "wordpress", nil
	}
	index := filepath.Join(projectPath, "index.php")
	if fileExists(index) {
		// check if CodeIgniter string present
		f, err := os.Open(index)
		if err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for i := 0; i < 100 && scanner.Scan(); i++ {
				line := scanner.Text()
				if strings.Contains(line, "CodeIgniter") {
					return "codeigniter", nil
				}
			}
		}
		return "php", nil
	}
	return "unknown", nil
}

func detectPHPVersionFromComposer(projectPath string) (string, error) {
	cf := filepath.Join(projectPath, "composer.json")
	if !fileExists(cf) {
		return "", nil
	}
	f, err := os.Open(cf)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "\"php\"") {
			// crude extraction: match first occurrence of x.y or x.y.z
			// e.g. "php": "^7.4|^8.0"
			toks := strings.Fields(line)
			for _, tok := range toks {
				// remove quotes and commas
				clean := strings.Trim(tok, `",`)
				// try to find pattern digit.digit
				parts := strings.Split(clean, ".")
				if len(parts) >= 2 {
					if _, err := strconv.Atoi(parts[0]); err == nil {
						if _, err2 := strconv.Atoi(strings.TrimRight(parts[1], `^~>=`)); err2 == nil {
							return fmt.Sprintf("%s.%s", parts[0], strings.TrimRight(parts[1], `^~>=`)), nil
						}
					}
				}
			}
		}
	}
	return "", nil
}

func latestInstalledPHP() (string, error) {
	// list /etc/php and return latest by sort
	dir := "/etc/php"
	f, err := os.Open(dir)
	if err != nil {
		return "", err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", nil
	}
	// sort descending lexicographically (works for "7.4" "8.1")
	// simple selection of max
	max := names[0]
	for _, n := range names {
		if n > max {
			max = n
		}
	}
	return max, nil
}

// writeFileAtomic writes file content with mode and ownership.
func writeFileAtomic(path string, content []byte, mode os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	return nil
}

// DeploySite performs the deployment for a single project.
func DeployPHPSite(domain, sysUser, basePath string) error {
	ctx := context.Background()

	if os.Geteuid() != 0 {
		return errors.New("must run as root")
	}

	if domain == "" || sysUser == "" {
		return errors.New("domain and sysUser are required")
	}

	params := deployParams{
		Domain:   domain,
		SysUser:  sysUser,
		BasePath: basePath,
	}
	params.ProjectPath = filepath.Join(basePath, domain)

	if !fileExists(params.ProjectPath) {
		return fmt.Errorf("project path does not exist: %s", params.ProjectPath)
	}
	fmt.Printf("Deploying %s -> %s (user: %s)\n", params.Domain, params.ProjectPath, params.SysUser)

	// 1) Ensure www-data is in user's group
	out, err := run(ctx, "id", "-nG", "www-data")
	if err == nil {
		if !strings.Contains(out, params.SysUser) {
			if _, err := run(ctx, "usermod", "-aG", params.SysUser, "www-data"); err != nil {
				return fmt.Errorf("failed to add www-data to group %s: %w\noutput:%s", params.SysUser, err, out)
			}
			fmt.Println("Added www-data to user's group")
		}
	} else {
		// Not fatal: continue, but log
		fmt.Printf("Warning: couldn't query www-data groups: %v\n", err)
	}

	// 1b) Ensure traversal bits on /home/user and basePath
	homeDir := filepath.Join("/home", params.SysUser)
	if fileExists(homeDir) {
		if _, err := run(ctx, "chmod", "g+x", homeDir); err != nil {
			fmt.Printf("Warning chmod g+x %s: %v\n", homeDir, err)
		}
	}
	if fileExists(basePath) {
		if _, err := run(ctx, "chmod", "g+x", basePath); err != nil {
			fmt.Printf("Warning chmod g+x %s: %v\n", basePath, err)
		}
	}

	// 1c) chown project to client user (do not force root)
	if _, err := run(ctx, "chown", "-R", fmt.Sprintf("%s:%s", params.SysUser, params.SysUser), params.ProjectPath); err != nil {
		return fmt.Errorf("chown project path failed: %w", err)
	}

	// 2) Install basic deps (software-properties-common) and apt helpers
	fmt.Println("Ensuring apt helpers & basic packages ...")
	if _, err := runInteractive(ctx, aptTimeout, "apt-get", "update", "-qq"); err != nil {
		return fmt.Errorf("apt-get update failed: %w", err)
	}
	if _, err := runInteractive(ctx, aptTimeout, "apt-get", "install", "-y", "-qq", "software-properties-common", "apt-transport-https", "ca-certificates", "gnupg", "lsb-release"); err != nil {
		return fmt.Errorf("install apt helpers failed: %w", err)
	}

	// Add ondrej/php PPA if not present
	ppaListOut, _ := run(ctx, "bash", "-lc", "grep -R \"ondrej/php\" /etc/apt/sources.list /etc/apt/sources.list.d || true")
	if strings.TrimSpace(ppaListOut) == "" {
		fmt.Println("Adding ondrej/php PPA ...")
		if _, err := runInteractive(ctx, aptTimeout, "add-apt-repository", "-y", "ppa:ondrej/php"); err != nil {
			// Not fatal; continue but warn
			fmt.Printf("Warning: add-apt-repository failed: %v\n", err)
		}
	}
	if _, err := runInteractive(ctx, aptTimeout, "apt-get", "update", "-qq"); err != nil {
		return fmt.Errorf("apt-get update after ppa failed: %w", err)
	}
	if _, err := runInteractive(ctx, aptTimeout, "apt-get", "install", "-y", "-qq", "acl", "zip", "unzip"); err != nil {
		return fmt.Errorf("apt-get install acl/zip failed: %w", err)
	}

	// 3) Determine PHP version
	raw, _ := detectPHPVersionWrapper(params.ProjectPath)
	target := raw
	if target == "" {
		installed, _ := latestInstalledPHP()
		if installed != "" {
			target = installed
		} else {
			target = "7.4"
		}
	}
	// normalize to X.Y
	parts := strings.Split(target, ".")
	if len(parts) >= 2 {
		target = parts[0] + "." + parts[1]
	}
	params.TargetPHP = target
	fmt.Printf("Target PHP version: %s\n", params.TargetPHP)

	// Install php if missing
	if !fileExists(filepath.Join("/etc/php", params.TargetPHP, "fpm")) {
		fmt.Printf("Installing php%s and common extensions...\n", params.TargetPHP)
		pkgs := []string{
			fmt.Sprintf("php%s-fpm", params.TargetPHP),
			fmt.Sprintf("php%s-cli", params.TargetPHP),
			fmt.Sprintf("php%s-mysql", params.TargetPHP),
			fmt.Sprintf("php%s-xml", params.TargetPHP),
			fmt.Sprintf("php%s-mbstring", params.TargetPHP),
			fmt.Sprintf("php%s-curl", params.TargetPHP),
			fmt.Sprintf("php%s-gd", params.TargetPHP),
			fmt.Sprintf("php%s-zip", params.TargetPHP),
			fmt.Sprintf("php%s-json", params.TargetPHP),
		}
		args := append([]string{"install", "-y", "-qq"}, pkgs...)
		if _, err := runInteractive(ctx, aptTimeout, "apt-get", args...); err != nil {
			return fmt.Errorf("failed to install php packages: %w", err)
		}
	}

	// 4) PHP-FPM pool creation
	socket := fmt.Sprintf("/run/php/php%s-%s-fpm.sock", params.TargetPHP, params.Domain)
	params.Socket = socket
	poolPath := filepath.Join("/etc/php", params.TargetPHP, "fpm", "pool.d", params.Domain+".conf")
	tpl := template.Must(template.New("pool").Parse(phpPoolTpl))
	var poolBuf bytes.Buffer
	_ = tpl.Execute(&poolBuf, map[string]string{
		"Domain":      params.Domain,
		"SysUser":     params.SysUser,
		"Socket":      socket,
		"ProjectPath": params.ProjectPath,
	})
	if err := writeFileAtomic(poolPath, poolBuf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing pool file failed: %w", err)
	}
	// restart php-fpm
	if _, err := run(ctx, "systemctl", "restart", fmt.Sprintf("php%s-fpm", params.TargetPHP)); err != nil {
		return fmt.Errorf("restart php-fpm failed: %w", err)
	}
	time.Sleep(phpFpmProcWait)

	// 5) Ensure framework-specific writable dirs
	framework, err := detectFramework(params.ProjectPath)
	if err != nil {
		return fmt.Errorf("framework detect failed: %w", err)
	}
	params.Framework = framework
	fmt.Printf("Framework: %s\n", framework)

	switch framework {
	case "laravel":
		ensureDir(filepath.Join(params.ProjectPath, "storage"))
		ensureDir(filepath.Join(params.ProjectPath, "bootstrap", "cache"))
		if _, err := run(ctx, "chown", "-R", fmt.Sprintf("%s:%s", params.SysUser, params.SysUser), filepath.Join(params.ProjectPath, "storage")); err != nil {
			return fmt.Errorf("chown storage failed: %w", err)
		}
		if _, err := run(ctx, "chown", "-R", fmt.Sprintf("%s:%s", params.SysUser, params.SysUser), filepath.Join(params.ProjectPath, "bootstrap", "cache")); err != nil {
			return fmt.Errorf("chown bootstrap cache failed: %w", err)
		}
		if _, err := run(ctx, "chmod", "-R", "775", filepath.Join(params.ProjectPath, "storage")); err != nil {
			return fmt.Errorf("chmod storage failed: %w", err)
		}
		if _, err := run(ctx, "chmod", "-R", "775", filepath.Join(params.ProjectPath, "bootstrap", "cache")); err != nil {
			return fmt.Errorf("chmod bootstrap cache failed: %w", err)
		}
	case "codeigniter":
		ensureDir(filepath.Join(params.ProjectPath, "application", "logs"))
		ensureDir(filepath.Join(params.ProjectPath, "application", "cache"))
		if _, err := run(ctx, "chown", "-R", fmt.Sprintf("%s:%s", params.SysUser, params.SysUser), filepath.Join(params.ProjectPath, "application", "logs")); err != nil {
			return fmt.Errorf("chown ci logs failed: %w", err)
		}
		if _, err := run(ctx, "chown", "-R", fmt.Sprintf("%s:%s", params.SysUser, params.SysUser), filepath.Join(params.ProjectPath, "application", "cache")); err != nil {
			return fmt.Errorf("chown ci cache failed: %w", err)
		}
		if _, err := run(ctx, "chmod", "-R", "775", filepath.Join(params.ProjectPath, "application", "logs")); err != nil {
			return fmt.Errorf("chmod ci logs failed: %w", err)
		}
		if _, err := run(ctx, "chmod", "-R", "775", filepath.Join(params.ProjectPath, "application", "cache")); err != nil {
			return fmt.Errorf("chmod ci cache failed: %w", err)
		}
	case "wordpress":
		ensureDir(filepath.Join(params.ProjectPath, "wp-content"))
		if _, err := run(ctx, "chown", "-R", fmt.Sprintf("%s:%s", params.SysUser, params.SysUser), filepath.Join(params.ProjectPath, "wp-content")); err != nil {
			return fmt.Errorf("chown wp-content failed: %w", err)
		}
		if _, err := run(ctx, "chmod", "-R", "775", filepath.Join(params.ProjectPath, "wp-content")); err != nil {
			return fmt.Errorf("chmod wp-content failed: %w", err)
		}
	default:
		// nothing special
	}

	// ensure project ownership by client
	if _, err := run(ctx, "chown", "-R", fmt.Sprintf("%s:%s", params.SysUser, params.SysUser), params.ProjectPath); err != nil {
		return fmt.Errorf("chown project failed: %w", err)
	}

	// set base permissions files 644, dirs 755
	if _, err := run(ctx, "find", params.ProjectPath, "-type", "f", "-exec", "chmod", "644", "{}", ";"); err != nil {
		return fmt.Errorf("chmod files failed: %w", err)
	}
	if _, err := run(ctx, "find", params.ProjectPath, "-type", "d", "-exec", "chmod", "755", "{}", ";"); err != nil {
		return fmt.Errorf("chmod dirs failed: %w", err)
	}

	// 6) run composer as client user if composer.json exists
	composerPath := ""
	if p, err := exec.LookPath("composer"); err == nil {
		composerPath = p
	} else if fileExists("/usr/local/bin/composer") {
		composerPath = "/usr/local/bin/composer"
	}
	if fileExists(filepath.Join(params.ProjectPath, "composer.json")) && composerPath != "" {
		fmt.Println("Running composer install as site owner ...")
		phpBin := fmt.Sprintf("/usr/bin/php%s", params.TargetPHP)
		if _, err := run(ctx, "sudo", "-H", "-u", params.SysUser, "bash", "-lc", fmt.Sprintf("COMPOSER_HOME=$HOME/.composer %s %s install --no-dev --optimize-autoloader --no-interaction", phpBin, composerPath)); err != nil {
			// warn but continue
			fmt.Printf("Warning: composer install returned error: %v\n", err)
		}
	}

	// 7) Write nginx config
	webRoot := params.ProjectPath
	// if there's public and no index.php at root, use public
	if fileExists(filepath.Join(params.ProjectPath, "public")) && !fileExists(filepath.Join(params.ProjectPath, "index.php")) {
		webRoot = filepath.Join(params.ProjectPath, "public")
	}
	params.WebRoot = webRoot

	nginxPath := filepath.Join("/etc/nginx/sites-available", params.Domain)
	tplNginx := template.Must(template.New("nginx").Parse(nginxTpl))
	var nginxBuf bytes.Buffer
	_ = tplNginx.Execute(&nginxBuf, map[string]string{
		"Domain":  params.Domain,
		"WebRoot": params.WebRoot,
		"Socket":  params.Socket,
	})
	if err := writeFileAtomic(nginxPath, nginxBuf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing nginx config failed: %w", err)
	}
	// symlink
	if _, err := run(ctx, "ln", "-sf", nginxPath, filepath.Join("/etc/nginx/sites-enabled", params.Domain)); err != nil {
		return fmt.Errorf("creating nginx symlink failed: %w", err)
	}
	// test & reload
	if out, err := run(ctx, "nginx", "-t"); err != nil {
		return fmt.Errorf("nginx -t failed: %w\noutput:%s", err, out)
	}
	if _, err := run(ctx, "systemctl", "reload", "nginx"); err != nil {
		return fmt.Errorf("systemctl reload nginx failed: %w", err)
	}

	fmt.Printf("Deployment finished. PHP-FPM running as %s on socket %s\n", params.SysUser, params.Socket)
	return nil
}

// detectPHPVersionWrapper tries composer, else returns empty
func detectPHPVersionWrapper(projectPath string) (string, error) {
	v, err := detectPHPVersionFromComposer(projectPath)
	if err != nil {
		return "", err
	}
	if v != "" {
		// normalize to major.minor
		parts := strings.Split(v, ".")
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1], nil
		}
		return v, nil
	}
	// fallback to latest installed
	installed, err := latestInstalledPHP()
	if err == nil && installed != "" {
		return installed, nil
	}
	return "", nil
}

// func main() {
// 	if len(os.Args) < 3 {
// 		fmt.Println("Usage: sudo ./deploy_with_owner <domain> <sys_user> [base_path]")
// 		os.Exit(2)
// 	}
// 	domain := os.Args[1]
// 	sysUser := os.Args[2]
// 	basePath := "/home/" + sysUser + "/projuktisheba/bin"
// 	if len(os.Args) >= 4 {
// 		basePath = os.Args[3]
// 	}
// 	if err := DeploySite(domain, sysUser, basePath); err != nil {
// 		fmt.Fprintf(os.Stderr, "Deployment failed: %v\n", err)
// 		os.Exit(1)
// 	}
// }
