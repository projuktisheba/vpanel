package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DeployLaravelSite deploys a Laravel site with a domain-specific FPM pool, nginx vhost,
// composer install (with fallback), artisan key:generate + optimize, and permission fixes.
// Call: DeployLaravelSite("example.com", "/home/samiul/projuktisheba/bin/PHP/example.com", "samiul")
func DeployLaravelSite(domain, projectPath, sysUser string) error {
	if domain == "" || projectPath == "" || sysUser == "" {
		return fmt.Errorf("domain, projectPath and sysUser are required")
	}

	// 1) Detect PHP version (robust, safe default)
	phpVersion := detectPHPVersionSafe(projectPath) // e.g. "8.2"

	// explicit binary paths to avoid PATH / sudo issues
	phpBin := fmt.Sprintf("/usr/bin/php%s", phpVersion)
	// domain-specific socket and pool
	socketPath := fmt.Sprintf("/run/php/php%s-%s-fpm.sock", phpVersion, domain)
	poolPath := fmt.Sprintf("/etc/php/%s/fpm/pool.d/%s.conf", phpVersion, domain)
	logPath := fmt.Sprintf("/var/log/php%s-%s-error.log", phpVersion, domain)

	// 2) Create FPM pool file (owned by root via sudo tee)
	poolContent := fmt.Sprintf(`[%s]
user = %s
group = %s
listen = %s
listen.owner = www-data
listen.group = www-data
listen.mode = 0660

pm = dynamic
pm.max_children = 10
pm.start_servers = 3
pm.min_spare_servers = 2
pm.max_spare_servers = 6

catch_workers_output = yes
php_admin_value[error_log] = %s
php_admin_flag[log_errors] = on
chdir = /
`, domain, sysUser, sysUser, socketPath, logPath)

	if err := writeWithSudo(poolPath, []byte(poolContent)); err != nil {
		return fmt.Errorf("write fpm pool: %w", err)
	}

	// ensure log file exists and has sane ownership
	if err := runCmdSudo("touch", logPath); err != nil {
		return fmt.Errorf("touch log: %w", err)
	}
	_ = runCmdSudo("chown", fmt.Sprintf("%s:%s", sysUser, sysUser), logPath)
	_ = runCmdSudo("chmod", "644", logPath)

	// 3) Restart php-fpm service for that version
	if err := runCmdSudo("systemctl", "restart", fmt.Sprintf("php%s-fpm", phpVersion)); err != nil {
		return fmt.Errorf("restart php%s-fpm: %w", phpVersion, err)
	}

	// 4) Detect public directory (Laravel)
	publicDir := detectPublicDir(projectPath)

	// 5) Create nginx config (use $document_root not $realpath_root)
	nginxConfPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", domain)
	nginxConf := fmt.Sprintf(`server {
    listen 80;
    server_name %s www.%s;
    root %s;
    index index.php index.html;

    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";

    charset utf-8;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location = /favicon.ico { access_log off; log_not_found off; }
    location = /robots.txt  { access_log off; log_not_found off; }

    error_page 404 /index.php;

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:%s;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_buffers 16 16k;
        fastcgi_buffer_size 32k;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }
}
`, domain, domain, publicDir, socketPath)

	if err := writeWithSudo(nginxConfPath, []byte(nginxConf)); err != nil {
		return fmt.Errorf("write nginx conf: %w", err)
	}

	// enable site & test nginx
	available := nginxConfPath
	enabled := fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", domain)
	if err := runCmdSudo("ln", "-sf", available, enabled); err != nil {
		return fmt.Errorf("enable nginx site: %w", err)
	}
	if err := runCmdSudo("nginx", "-t"); err != nil {
		return fmt.Errorf("nginx -t failed: %w", err)
	}
	if err := runCmdSudo("systemctl", "reload", "nginx"); err != nil {
		return fmt.Errorf("nginx reload failed: %w", err)
	}

	// 6) Fix ownership & permissions BEFORE composer (avoid permission issues)
	// chown project to sysUser (you may prefer www-data or sysUser depending on your setup)
	if err := runCmdSudo("chown", "-R", fmt.Sprintf("%s:%s", sysUser, sysUser), projectPath); err != nil {
		// non-fatal but warn
		return fmt.Errorf("chown project failed: %w", err)
	}
	// ensure storage + bootstrap/cache exist, owned and writable
	storage := filepath.Join(projectPath, "storage")
	bootstrapCache := filepath.Join(projectPath, "bootstrap", "cache")
	_ = runCmdSudo("mkdir", "-p", storage)
	_ = runCmdSudo("mkdir", "-p", bootstrapCache)
	if err := runCmdSudo("chown", "-R", fmt.Sprintf("%s:%s", sysUser, sysUser), storage); err != nil {
		return fmt.Errorf("chown storage: %w", err)
	}
	if err := runCmdSudo("chown", "-R", fmt.Sprintf("%s:%s", sysUser, sysUser), bootstrapCache); err != nil {
		return fmt.Errorf("chown bootstrap cache: %w", err)
	}
	if err := runCmdSudo("chmod", "-R", "775", storage); err != nil {
		return fmt.Errorf("chmod storage: %w", err)
	}
	if err := runCmdSudo("chmod", "-R", "775", bootstrapCache); err != nil {
		return fmt.Errorf("chmod bootstrap cache: %w", err)
	}

	// 7) Run composer install (install -> fallback update). Run as sysUser via sudo -u for safety.
	if err := runComposerAsUser(projectPath, phpBin, sysUser); err != nil {
		return fmt.Errorf("composer install/update failed: %w", err)
	}

	// 8) Run artisan key:generate + optimize (as sysUser)
	if err := runArtisanAsUser(projectPath, phpBin, sysUser); err != nil {
		return fmt.Errorf("artisan commands failed: %w", err)
	}

	// 9) Final permission fix (ensure www-data or sysUser ownership per your policy)
	// Keep as sysUser here
	if err := runCmdSudo("chown", "-R", fmt.Sprintf("%s:%s", sysUser, sysUser), projectPath); err != nil {
		return fmt.Errorf("final chown failed: %w", err)
	}
	_ = runCmdSudo("chmod", "-R", "755", projectPath)

	return nil
}

// ---------------------- helpers ----------------------

func detectPHPVersionSafe(projectPath string) string {
	// default to 8.2 (modern and compatible)
	defaultVer := "8.2"
	composerPath := filepath.Join(projectPath, "composer.json")
	data, err := os.ReadFile(composerPath)
	if err != nil {
		return defaultVer
	}
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return defaultVer
	}
	require, ok := doc["require"].(map[string]interface{})
	if !ok {
		return defaultVer
	}
	raw, ok := require["php"].(string)
	if !ok {
		return defaultVer
	}
	raw = strings.ToLower(raw)
	// look for known major.minor tokens
	for _, v := range []string{"8.4", "8.3", "8.2", "8.1", "7.4"} {
		if strings.Contains(raw, v) {
			return v
		}
	}
	// fallback if constraint contains ">=8.2" or similar
	for _, v := range []string{"8.4", "8.3", "8.2", "8.1", "7.4"} {
		if strings.Contains(raw, strings.TrimPrefix(v, "8.")) {
			return v
		}
	}
	return defaultVer
}

func writeWithSudo(path string, content []byte) error {
	cmd := exec.Command("sudo", "tee", path)
	cmd.Stdin = strings.NewReader(string(content))
	// capture potential error output
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sudo tee failed: %w: %s", err, string(out))
	}
	return nil
}

func runCmdSudo(name string, args ...string) error {
	all := append([]string{name}, args...)
	cmd := exec.Command("sudo", all...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command sudo %s failed: %w: %s", strings.Join(all, " "), err, string(out))
	}
	return nil
}

func detectPublicDir(projectPath string) string {
	publicDir := filepath.Join(projectPath, "public")
	if _, err := os.Stat(publicDir); err == nil {
		return publicDir
	}
	return projectPath
}

func runComposerAsUser(projectPath, phpBin, sysUser string) error {
	// try install first
	composerBin := "/usr/local/bin/composer"
	if _, err := os.Stat(composerBin); err != nil {
		// fallback to composer in PATH
		composerBin = "composer"
	}
	installArgs := []string{phpBin, composerBin, "install", "--no-dev", "--optimize-autoloader", "-n", "--no-interaction"}
	cmd := exec.Command("sudo", append([]string{"-u", sysUser}, installArgs...)...)
	cmd.Dir = projectPath
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	outStr := string(out)
	// if composer complains about lock/constraint, try update
	if strings.Contains(strings.ToLower(outStr), "lock file") || strings.Contains(strings.ToLower(outStr), "constraint") || strings.Contains(strings.ToLower(outStr), "your requirements could not be resolved") {
		updateArgs := []string{phpBin, composerBin, "update", "--no-dev", "--optimize-autoloader", "-n", "--no-interaction"}
		cmd2 := exec.Command("sudo", append([]string{"-u", sysUser}, updateArgs...)...)
		cmd2.Dir = projectPath
		out2, err2 := cmd2.CombinedOutput()
		if err2 != nil {
			return fmt.Errorf("composer update failed: %w: %s", err2, string(out2))
		}
		return nil
	}
	return fmt.Errorf("composer install failed: %s", outStr)
}
func runArtisanAsUser(projectPath, phpBin, sysUser string) error {
	artisan := filepath.Join(projectPath, "artisan")

	// Check if artisan exists
	if _, err := os.Stat(artisan); err != nil {
		return nil // nothing to run
	}

	// List of artisan commands to run
	commands := [][]string{
		{"key:generate", "--force"},
		{"optimize"},
		{"storage:link"},
		{"route:clear"},
		{"route:cache"},
		{"config:clear"},
		{"config:cache"},
		{"view:clear"},
		{"view:cache"},
	}

	for _, args := range commands {
		cmdArgs := append([]string{"-u", sysUser, phpBin, artisan}, args...)
		cmd := exec.Command("sudo", cmdArgs...)
		cmd.Dir = projectPath

		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("artisan %s failed: %w: %s", strings.Join(args, " "), err, string(out))
		}
	}

	return nil
}
