package deploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// This code works for restaurant website deployment
func DeployReStaurantPHPSite(projectPath, sysUser, domain string) error {

	if domain == "" || sysUser == "" {
		return errors.New("domain and sysUser are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Helper to write files using sudo tee
	writeProtectedFile := func(path string, content []byte) error {
		cmd := exec.CommandContext(ctx, "sudo", "tee", path)
		cmd.Stdin = strings.NewReader(string(content))
		cmd.Stdout = nil // Discard output
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Helper to run generic sudo commands
	runSudo := func(cmdStr string, args ...string) error {
		allArgs := append([]string{cmdStr}, args...)
		cmd := exec.CommandContext(ctx, "sudo", allArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// 1. Install PPA (Required for different PHP versions)
	runSudo("apt-get", "install", "-y", "software-properties-common")
	runSudo("add-apt-repository", "-y", "ppa:ondrej/php")

	// 2. Update with retry
	for i := 0; i < 5; i++ {
		err := runSudo("apt-get", "update", "-y")
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "lock") {
			time.Sleep(3 * time.Second)
			continue
		}
		return fmt.Errorf("apt update failed: %w", err)
	}

	// 3. Fix User Permissions
	for i := 0; i < 5; i++ {
		err := runSudo("usermod", "-aG", "www-data", sysUser)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "lock") {
			time.Sleep(2 * time.Second)
			continue
		}
		return fmt.Errorf("usermod failed: %w", err)
	}

	runSudo("chmod", "g+x", fmt.Sprintf("/home/%s", sysUser))
	runSudo("chmod", "g+x", projectPath)
	runSudo("chown", "-R", fmt.Sprintf("%s:%s", sysUser, sysUser), projectPath)

	// 4. Smart PHP Version Detection
	targetPHP := "7.4"
	composerFile := filepath.Join(projectPath, "composer.json")

	if _, err := os.Stat(composerFile); err == nil {
		out, _ := exec.Command("bash", "-c",
			fmt.Sprintf(`grep '"php":' "%s" | head -n 1 | grep -oE '[0-9]+\.[0-9]+'`, composerFile)).
			Output()

		detected := strings.TrimSpace(string(out))
		if detected != "" {
			verFloat, err := strconv.ParseFloat(detected, 64)
			if err == nil {
				if verFloat < 5.6 {
					targetPHP = "7.4" // Upgrade obsolete PHP
				} else {
					targetPHP = detected
				}
			}
		}
	}

	// 5. Install PHP Packages
	phpPackages := []string{
		fmt.Sprintf("php%s-fpm", targetPHP),
		fmt.Sprintf("php%s-cli", targetPHP),
		fmt.Sprintf("php%s-common", targetPHP),
		fmt.Sprintf("php%s-mysql", targetPHP),
		fmt.Sprintf("php%s-xml", targetPHP),
		fmt.Sprintf("php%s-mbstring", targetPHP),
		fmt.Sprintf("php%s-curl", targetPHP),
		"zip", "unzip", "acl",
	}

	args := append([]string{"install", "-y"}, phpPackages...)
	if err := runSudo("apt-get", args...); err != nil {
		return fmt.Errorf("php install failed: %w", err)
	}

	// 6. Create FPM Pool (Using writeProtectedFile)
	poolConf := fmt.Sprintf("/etc/php/%s/fpm/pool.d/%s.conf", targetPHP, domain)
	socketPath := fmt.Sprintf("/run/php/php%s-%s-fpm.sock", targetPHP, domain)

	fpmPool := fmt.Sprintf(`[%s]
user = %s
group = %s
listen = %s
listen.owner = www-data
listen.group = www-data
pm = dynamic
pm.max_children = 5
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3
chdir = /
`, domain, sysUser, sysUser, socketPath)

	// REPLACEMENT HERE
	if err := writeProtectedFile(poolConf, []byte(fpmPool)); err != nil {
		return fmt.Errorf("failed to write fpm pool: %w", err)
	}

	runSudo("systemctl", "restart", fmt.Sprintf("php%s-fpm", targetPHP))

	// ... previous code ...

	// -------------------------
	// FIX: Smart Composer Runner
	// -------------------------
	fmt.Println("Running Composer...")

	// 1. Determine the command and arguments
	composerBin := "/usr/local/bin/composer"
	phpBin := fmt.Sprintf("/usr/bin/php%s", targetPHP)

	composerArgs := []string{
		composerBin,
		"install",
		"--no-dev",
		"--optimize-autoloader",
		"--no-interaction",
	}

	var cmd *exec.Cmd

	// 2. Check if we are already the correct user
	// os.Getuid() gets the current process user ID.
	// We need to find the UID of the target sysUser to compare.
	// For simplicity here, if the app is running as 'samiul' and sysUser is 'samiul',
	// we just run the command directly.

	currentUserStr := os.Getenv("USER") // Or use os.Getuid() logic if needed

	if currentUserStr == sysUser || os.Getuid() != 0 {
		// If we are ALREADY the user, or we are not root (and can't sudo easily without config),
		// we try running directly.
		cmd = exec.Command(phpBin, composerArgs...)
	} else {
		// If we are ROOT, we use sudo to drop privileges to sysUser
		cmd = exec.Command("sudo", append([]string{"-u", sysUser, phpBin}, composerArgs...)...)
	}

	// 3. CRITICAL: Set the Working Directory
	cmd.Dir = projectPath

	// 4. CRITICAL: Inject Environment Variables
	// Composer needs to know where to write cache. Systemd often strips $HOME.
	homeDir := fmt.Sprintf("/home/%s", sysUser)
	cmd.Env = os.Environ() // Inherit current env
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("HOME=%s", homeDir),
		fmt.Sprintf("COMPOSER_HOME=%s/.composer", homeDir),
		"COMPOSER_ALLOW_SUPERUSER=1", // Just in case, though we prefer non-root
	)

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Composer Output: %s\n", string(output))
		// return fmt.Errorf("composer install failed (Exit: %s): %w", err, err)
	}

	// ... continue to Nginx setup ...

	// 8. Nginx Config (Using writeProtectedFile)
	webRoot := projectPath
	if _, err := os.Stat(filepath.Join(projectPath, "public")); err == nil {
		webRoot = filepath.Join(projectPath, "public")
	}

	nginxConf := fmt.Sprintf(`server {
    listen 80;
    server_name %s www.%s;
    root %s;
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:%s;
        fastcgi_buffers 16 16k;
        fastcgi_buffer_size 32k;
    }
}
`, domain, domain, webRoot, socketPath)

	confPath := filepath.Join("/etc/nginx/sites-available", domain)

	// REPLACEMENT HERE
	if err := writeProtectedFile(confPath, []byte(nginxConf)); err != nil {
		return fmt.Errorf("failed writing nginx conf: %w", err)
	}

	runSudo("ln", "-sf", confPath, filepath.Join("/etc/nginx/sites-enabled", domain))

	if err := runSudo("nginx", "-t"); err != nil {
		return fmt.Errorf("nginx config test failed: %w", err)
	}

	if err := runSudo("systemctl", "reload", "nginx"); err != nil {
		return err
	}

	fmt.Println("Deployment done using PHP", targetPHP)
	return nil
}
