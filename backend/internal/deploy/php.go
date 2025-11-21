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
