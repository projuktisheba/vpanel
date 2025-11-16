package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func appNameGenerator(domain string)string{
	parts := strings.Split(domain, ".")
	return strings.Join(parts, "_")
}

// DeployLaravelHandler deploys a Laravel app to a custom folder
func DeployLaravelHandler(appFramework, domain, phpVersion, dbName, dbUser, dbPass string) error {

	fmt.Println("Starting Laravel deployment...")
	// Generating appName
	appName := appNameGenerator(domain)
	// Generating appDir
	// home, _:= 
	appDir := appNameGenerator(appFramework)
	// ---------------------------------------------------------
	// Helper to run commands through SUDO automatically
	// ---------------------------------------------------------
	run := func(cmd string, args ...string) error {
		full := append([]string{cmd}, args...) // prefix with real command
		fmt.Printf("Running: sudo %s %s\n", cmd, strings.Join(args, " "))
		c := exec.Command("sudo", full...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	// ---------------------------------------------------------
	// 1. Update system
	// ---------------------------------------------------------
	if err := run("apt", "update"); err != nil {
		return err
	}
	if err := run("apt", "upgrade", "-y"); err != nil {
		return err
	}

	// ---------------------------------------------------------
	// 2. Install required packages
	// ---------------------------------------------------------
	pkgs := []string{"nginx", "mysql-server", "unzip", "git", "curl", "software-properties-common"}
	args := append([]string{"install", "-y"}, pkgs...)
	if err := run("apt", args...); err != nil {
		return err
	}

	// ---------------------------------------------------------
	// 3. Add PHP PPA + install PHP packages
	// ---------------------------------------------------------
	if err := run("add-apt-repository", "ppa:ondrej/php", "-y"); err != nil {
		return err
	}
	if err := run("apt", "update"); err != nil {
		return err
	}

	// Auto detect PHP version
	detectedPHP, err := DetectPHPVersion()
	if err != nil {
		return fmt.Errorf("cannot detect PHP version: %v", err)
	}
	phpVersion = detectedPHP

	fmt.Println("Detected PHP version:", phpVersion)

	phpPkgs := []string{
		"php" + phpVersion,
		"php" + phpVersion + "-fpm",
		"php" + phpVersion + "-cli",
		"php" + phpVersion + "-mbstring",
		"php" + phpVersion + "-xml",
		"php" + phpVersion + "-bcmath",
		"php" + phpVersion + "-curl",
		"php" + phpVersion + "-zip",
		"php" + phpVersion + "-gd",
		"php" + phpVersion + "-intl",
		"php" + phpVersion + "-mysql",
		"php" + phpVersion + "-soap",
	}

	args = append([]string{"install", "-y"}, phpPkgs...)
	if err := run("apt", args...); err != nil {
		return err
	}

	// ---------------------------------------------------------
	// 4. Install Composer
	// ---------------------------------------------------------
	if err := run("php", "-r", "copy('https://getcomposer.org/installer','composer-setup.php');"); err != nil {
		return err
	}
	if err := run("php", "composer-setup.php", "--install-dir=/usr/local/bin", "--filename=composer"); err != nil {
		return err
	}
	if err := run("php", "-r", "unlink('composer-setup.php');"); err != nil {
		return err
	}

	// ---------------------------------------------------------
	// 5. Permissions
	// ---------------------------------------------------------
	if err := run("chown", "-R", "www-data:www-data", appDir); err != nil {
		return err
	}
	if err := run("find", appDir, "-type", "f", "-exec", "chmod", "644", "{}", ";"); err != nil {
		return err
	}
	if err := run("find", appDir, "-type", "d", "-exec", "chmod", "755", "{}", ";"); err != nil {
		return err
	}
	if err := run("chmod", "-R", "775", appDir+"/storage"); err != nil {
		return err
	}
	if err := run("chmod", "-R", "775", appDir+"/bootstrap/cache"); err != nil {
		return err
	}

	// ---------------------------------------------------------
	// 6. Composer install
	// ---------------------------------------------------------
	if err := run("composer", "install", "--optimize-autoloader", "--no-dev", "--working-dir="+appDir); err != nil {
		return err
	}

	// ---------------------------------------------------------
	// 7. Run Laravel Artisan Commands
	// ---------------------------------------------------------
	run("php", appDir+"/artisan", "key:generate")
	run("php", appDir+"/artisan", "migrate", "--force")
	run("php", appDir+"/artisan", "db:seed", "--force")
	run("php", appDir+"/artisan", "storage:link")
	run("php", appDir+"/artisan", "config:cache")
	run("php", appDir+"/artisan", "route:cache")
	run("php", appDir+"/artisan", "view:cache")

	// ---------------------------------------------------------
	// 8. Nginx Config
	// ---------------------------------------------------------
	nginxConfig := fmt.Sprintf("/etc/nginx/sites-available/%s", appName)

	configContent := fmt.Sprintf(`
		server {
			listen 80;
			server_name %s www.%s;

			root %s/public;
			index index.php index.html;

			add_header X-Frame-Options "SAMEORIGIN";
			add_header X-Content-Type-Options "nosniff";

			location / {
				try_files $uri $uri/ /index.php?$query_string;
			}

			location ~ \.php$ {
				include snippets/fastcgi-php.conf;
				fastcgi_pass unix:/run/php/php%s-fpm.sock;
				fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
				include fastcgi_params;
			}

			location ~ /\.ht {
				deny all;
			}
		}
		`, domain, domain, appDir, phpVersion)

	if err := os.WriteFile(nginxConfig, []byte(configContent), 0644); err != nil {
		return err
	}

	if err := run("ln", "-sf", nginxConfig, "/etc/nginx/sites-enabled/"); err != nil {
		return err
	}
	if err := run("nginx", "-t"); err != nil {
		return err
	}
	if err := run("systemctl", "reload", "nginx"); err != nil {
		return err
	}

	// ---------------------------------------------------------
	// 9. Install Certbot + Get SSL Certificate
	// ---------------------------------------------------------
	if err := run("apt", "install", "-y", "certbot", "python3-certbot-nginx"); err != nil {
		return err
	}

	if err := run("certbot", "--nginx",
		"-d", domain,
		"-d", "www."+domain,
		"--non-interactive",
		"--agree-tos",
		"-m", "admin@"+domain,
	); err != nil {
		return err
	}

	// ---------------------------------------------------------
	// 10. Enable Auto-Renew Timer
	// ---------------------------------------------------------
	if err := run("systemctl", "enable", "certbot.timer"); err != nil {
		return err
	}
	if err := run("systemctl", "start", "certbot.timer"); err != nil {
		return err
	}

	fmt.Printf("\nDeployment completed for %s at %s!\n", appName, appDir)
	return nil
}

// DeployCodeIgniterHandler deploys a CodeIgniter project to a custom folder with SSL
func DeployCodeIgniterHandler(appName, appDir, domain, phpVersion, gitRepo, gitBranch string) error {
	fmt.Println("Starting CodeIgniter deployment...")

	// Helper to run commands
	run := func(cmd string, args ...string) error {
		full := append([]string{cmd}, args...)
		fmt.Printf("Running: sudo %s %s\n", cmd, strings.Join(args, " "))
		c := exec.Command("sudo", full...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	// 1. System update
	run("apt", "update")
	run("apt", "upgrade", "-y")

	// 2. Install packages
	pkgs := []string{"nginx", "mysql-server", "unzip", "git", "curl", "software-properties-common"}
	run("apt", append([]string{"install", "-y"}, pkgs...)...)

	// 3. PHP installation
	run("add-apt-repository", "ppa:ondrej/php", "-y")
	run("apt", "update")
	phpPkgs := []string{
		fmt.Sprintf("php%s", phpVersion),
		fmt.Sprintf("php%s-fpm", phpVersion),
		fmt.Sprintf("php%s-cli", phpVersion),
		fmt.Sprintf("php%s-mbstring", phpVersion),
		fmt.Sprintf("php%s-xml", phpVersion),
		fmt.Sprintf("php%s-bcmath", phpVersion),
		fmt.Sprintf("php%s-curl", phpVersion),
		fmt.Sprintf("php%s-zip", phpVersion),
		fmt.Sprintf("php%s-gd", phpVersion),
		fmt.Sprintf("php%s-intl", phpVersion),
		fmt.Sprintf("php%s-mysql", phpVersion),
	}
	run("apt", append([]string{"install", "-y"}, phpPkgs...)...)

	// 4. Composer install (optional)
	run("php", "-r", "copy('https://getcomposer.org/installer','composer-setup.php');")
	run("php", "composer-setup.php", "--install-dir=/usr/local/bin", "--filename=composer")
	run("php", "-r", "unlink('composer-setup.php');")

	// 5. Clone project
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		run("git", "clone", "-b", gitBranch, gitRepo, appDir)
	} else {
		run("git", "-C", appDir, "fetch", "--all")
		run("git", "-C", appDir, "checkout", gitBranch)
		run("git", "-C", appDir, "pull", "origin", gitBranch)
	}

	// 6. Set permissions (CodeIgniter writable folder)
	run("chown", "-R", "www-data:www-data", appDir)
	run("find", appDir, "-type", "f", "-exec", "chmod", "644", "{}", ";")
	run("find", appDir, "-type", "d", "-exec", "chmod", "755", "{}", ";")
	run("chmod", "-R", "775", fmt.Sprintf("%s/writable", appDir))

	// 7. Nginx configuration
	nginxConfig := fmt.Sprintf("/etc/nginx/sites-available/%s", appName)
	configContent := fmt.Sprintf(`
server {
    listen 80;
    server_name %s www.%s;

    root %s/public;
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php%s-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
    }

    location ~ /\.ht {
        deny all;
    }
}
`, domain, domain, appDir, phpVersion)
	os.WriteFile(nginxConfig, []byte(configContent), 0644)
	run("ln", "-sf", nginxConfig, "/etc/nginx/sites-enabled/")
	run("nginx", "-t")
	run("systemctl", "reload", "nginx")

	// 8. SSL with Certbot
	run("apt", "install", "-y", "certbot", "python3-certbot-nginx")
	run("certbot", "--nginx", "-d", domain, "-d", "www."+domain, "--non-interactive", "--agree-tos", "-m", "admin@"+domain)
	run("systemctl", "enable", "certbot.timer")
	run("systemctl", "start", "certbot.timer")

	fmt.Printf("CodeIgniter deployment completed at %s!\n", appDir)
	return nil
}

// DetectPHPVersion automatically detects installed PHP version.
func DetectPHPVersion() (string, error) {
	// 1. Try "php -v"
	out, err := exec.Command("php", "-v").Output()
	if err == nil {
		version := parsePHPVersion(string(out))
		if version != "" {
			return version, nil
		}
	}

	// 2. Try update-alternatives list
	out, err = exec.Command("update-alternatives", "--list", "php").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "php") {
				v := extractVersionFromPath(line)
				if v != "" {
					return v, nil
				}
			}
		}
	}

	// 3. Try scanning PHP-FPM services
	out, err = exec.Command("bash", "-c", "ls /run/php/").Output()
	if err == nil {
		files := strings.Split(string(out), "\n")
		for _, f := range files {
			if strings.Contains(f, "php") && strings.Contains(f, "fpm.sock") {
				v := extractVersionFromPath(f)
				if v != "" {
					return v, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not detect PHP version")
}

// parsePHPVersion extracts PHP version from "php -v" output
func parsePHPVersion(s string) string {
	// Example: "PHP 8.3.2 ..."
	parts := strings.Split(s, " ")
	for _, p := range parts {
		if strings.Count(p, ".") >= 1 && strings.HasPrefix(p, "8.") {
			return p[:3] // "8.3"
		}
		if strings.Count(p, ".") >= 1 && strings.HasPrefix(p, "7.") {
			return p[:3]
		}
	}
	return ""
}

// extractVersionFromPath detects "php8.3", "php8.2-fpm", etc.
func extractVersionFromPath(path string) string {
	// Examples:
	// /usr/bin/php8.3
	// php8.2-fpm.sock
	for _, p := range strings.Split(path, "/") {
		if strings.HasPrefix(p, "php") {
			trimmed := strings.TrimPrefix(p, "php")
			if len(trimmed) >= 3 {
				return trimmed[:3] // returns "8.3"
			}
		}
	}
	return ""
}
