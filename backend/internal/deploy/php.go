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

func DeployPHPSite(ctx context.Context, projectPath, sysUser, domain string) error {

	if domain == "" || sysUser == "" {
		return errors.New("domain and sysUser are required")
	}

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

		// NEW REQUIRED EXTENSIONS:
		fmt.Sprintf("php%s-intl", targetPHP),   // Required by escpos-php
		fmt.Sprintf("php%s-gd", targetPHP),     // Required by phpspreadsheet
		fmt.Sprintf("php%s-bcmath", targetPHP), // Recommended for spreadsheets/financial math
		fmt.Sprintf("php%s-zip", targetPHP),    // Critical for Composer speed
		fmt.Sprintf("php%s-soap", targetPHP),   // Good to have for APIs
		"zip", "unzip", "acl",
	}

	args := append([]string{"install", "-y"}, phpPackages...)
	if err := runSudo("apt-get", args...); err != nil {
		return fmt.Errorf("php install failed: %w", err)
	}

	// 6. Create log directory
	logFile := fmt.Sprintf("/var/log/php%s-%s-error.log", targetPHP, domain)

	// 6.1 Create the log file
	if err := runSudo("touch", logFile); err != nil {
		return err
	}

	// 6.2change ownership
	if err := runSudo("chown", fmt.Sprintf("%s:%s", sysUser, sysUser), logFile); err != nil {
		fmt.Println("chown error:", err)
	}

	// 6.3. Set permissions
	if err := runSudo("chmod", "644", logFile); err != nil {
		fmt.Println("chown error:", err)
	}

	// 7. Create FPM Pool (Using writeProtectedFile)
	poolConf := fmt.Sprintf("/etc/php/%s/fpm/pool.d/%s.conf", targetPHP, domain)
	socketPath := fmt.Sprintf("/run/php/php%s-%s-fpm.sock", targetPHP, domain)

	fpmPool := fmt.Sprintf(`[%s]
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

`, domain, sysUser, sysUser, socketPath, logFile)

	// REPLACEMENT HERE
	if err := writeProtectedFile(poolConf, []byte(fpmPool)); err != nil {
		return fmt.Errorf("failed to write fpm pool: %w", err)
	}

	runSudo("systemctl", "restart", fmt.Sprintf("php%s-fpm", targetPHP))

	// ... (previous code remains the same) ...

	// -------------------------
	// FIX: Smart Composer Runner (Install with Fallback)
	// -------------------------
	fmt.Println("Running Composer...")

	// Helper function to run composer commands
	runComposer := func(action string) error {
		composerBin := "/usr/local/bin/composer"
		phpBin := fmt.Sprintf("/usr/bin/php%s", targetPHP)

		composerArgs := []string{
			composerBin,
			action,
			"--no-dev",
			"--optimize-autoloader",
			"--no-interaction",
			// "--ignore-platform-reqs", // Uncomment if you have extension issues
		}

		var cmd *exec.Cmd

		// Determine if we need sudo
		// If we are root, we MUST sudo down to sysUser
		if os.Geteuid() == 0 {
			cmd = exec.Command("sudo", append([]string{"-u", sysUser, phpBin}, composerArgs...)...)
		} else {
			// If already running as the user (or similar), run directly
			cmd = exec.Command(phpBin, composerArgs...)
		}

		cmd.Dir = projectPath

		// Inject Environment Variables
		homeDir := fmt.Sprintf("/home/%s", sysUser)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("HOME=%s", homeDir),
			fmt.Sprintf("COMPOSER_HOME=%s/.composer", homeDir),
		)

		// Capture output
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Return the error AND the output so we can see why
			return fmt.Errorf("%s", string(output))
		}
		return nil
	}

	// 1. Try 'install' first (Fastest, respects lock file)
	fmt.Println("Attempting 'composer install'...")
	if err := runComposer("install"); err != nil {

		// 2. Check if failure was due to Lock File Mismatch
		// Composer usually mentions "lock file" or "constraints" in the output
		errStr := err.Error()
		if strings.Contains(errStr, "lock file") || strings.Contains(errStr, "constraint") {
			fmt.Println("Composer.lock out of sync. Falling back to 'composer update'...")

			// 3. Fallback to 'update' (Slower, generates new lock file)
			if errUpdate := runComposer("update"); errUpdate != nil {
				return fmt.Errorf("composer update failed: %w", errUpdate)
			}
			fmt.Println("Composer update successful.")
		} else {
			// It was some other error (like permission or network), fail.
			return fmt.Errorf("composer install failed: %w", err)
		}
	}

	// ... (continue to Nginx setup) ...

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


// DeletePHPSite removes all traces of a deployed PHP site
func DeletePHPSite(ctx context.Context, projectPath, sysUser, domain string) error {
	if domain == "" || sysUser == "" {
		return fmt.Errorf("domain and sysUser are required")
	}

	// Helper to run sudo commands
	runSudo := func(cmdStr string, args ...string) error {
		allArgs := append([]string{cmdStr}, args...)
		cmd := exec.CommandContext(ctx, "sudo", allArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Dynamically detect PHP version for this project 
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

	// 1. Stop FPM pool if exists
	runSudo("systemctl", "stop", fmt.Sprintf("php%s-fpm", targetPHP))

	// 2. Remove project files
	runSudo("rm", "-rf", projectPath)

	// 3. Remove FPM pool file
	poolConf := fmt.Sprintf("/etc/php/%s/fpm/pool.d/%s.conf", targetPHP, domain)
	runSudo("rm", "-f", poolConf)

	// 4. Remove log file
	logFile := fmt.Sprintf("/var/log/php%s-%s-error.log", targetPHP, domain)
	runSudo("rm", "-f", logFile)

	// 5. Remove Nginx config
	nginxConf := filepath.Join("/etc/nginx/sites-available", domain)
	runSudo("rm", "-f", nginxConf)

	nginxEnabled := filepath.Join("/etc/nginx/sites-enabled", domain)
	runSudo("rm", "-f", nginxEnabled)

	// 6. Reload PHP-FPM and Nginx
	runSudo("systemctl", "reload", fmt.Sprintf("php%s-fpm", targetPHP))
	runSudo("systemctl", "reload", "nginx")

	// 7. Reset user permissions (optional)
	runSudo("chmod", "g-x", fmt.Sprintf("/home/%s", sysUser))

	fmt.Println("Project deleted successfully:", domain)
	return nil
}
