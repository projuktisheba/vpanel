package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type ComposerJSON struct {
	Require map[string]string `json:"require"`
}

func DeployPHPProject(appRootDirectory, appFramework, domain string) error {
	switch appFramework {
	case "Laravel":
		return deployLaravelHandler(appRootDirectory, appFramework, domain)
	case "CodeIgniter":
		return fmt.Errorf("Application framework %s does not supported", appFramework)
	case "Symfony":
		return fmt.Errorf("Application framework %s is not supported", appFramework)
	}
	return nil
}

func appNameGenerator(domain string) string {
	parts := strings.Split(domain, ".")
	return strings.Join(parts, "_")
}

// deployLaravelHandler deploys or updates a Laravel project at appRootDirectory
func deployLaravelHandler(appRootDirectory, appFramework, domain string) error {
	fmt.Println("Starting Laravel deployment...")

	appName := appNameGenerator(domain) // your app name generator

	// ---------------------------------------------------------
	// Helper to run commands through SUDO automatically
	// ---------------------------------------------------------
	run := func(cmd string, args ...string) error {
		full := append([]string{cmd}, args...)
		fmt.Printf("Running: sudo %s %s\n", cmd, strings.Join(args, " "))
		c := exec.Command("sudo", full...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	// ---------------------------------------------------------
	// 1. Update system packages (optional, safe)
	// ---------------------------------------------------------
	run("apt", "update")
	run("apt", "upgrade", "-y")

	// ---------------------------------------------------------
	// 2. Install required packages
	// ---------------------------------------------------------
	pkgs := []string{"nginx", "mysql-server", "unzip", "curl", "software-properties-common"}
	run("apt", append([]string{"install", "-y"}, pkgs...)...)

	// ---------------------------------------------------------
	// 3. Add PHP PPA + install PHP packages
	// ---------------------------------------------------------
	run("add-apt-repository", "ppa:ondrej/php", "-y")
	run("apt", "update")

	phpVersionRaw, err := RequiredPHPVersion(appRootDirectory)
	if err != nil {
		return fmt.Errorf("cannot detect PHP version: %v", err)
	}
	phpVersion := parsePHPVersion(phpVersionRaw) // ensure format like "8.1"
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
	run("apt", append([]string{"install", "-y"}, phpPkgs...)...)

	// ---------------------------------------------------------
	// 4. Install Composer if not exists
	// ---------------------------------------------------------
	if _, err := exec.LookPath("composer"); err != nil {
		run("php", "-r", "copy('https://getcomposer.org/installer','composer-setup.php');")
		run("php", "composer-setup.php", "--install-dir=/usr/local/bin", "--filename=composer")
		run("php", "-r", "unlink('composer-setup.php');")
	}

	// ---------------------------------------------------------
	// 5. Ensure appRootDirectory exists
	// ---------------------------------------------------------
	if err := os.MkdirAll(appRootDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create app directory: %v", err)
	}

	// ---------------------------------------------------------
	// 6. Set permissions only for required dirs
	// ---------------------------------------------------------
	run("chown", "-R", "www-data:www-data", appRootDirectory+"/storage")
	run("chown", "-R", "www-data:www-data", appRootDirectory+"/bootstrap/cache")
	run("chmod", "-R", "775", appRootDirectory+"/storage")
	run("chmod", "-R", "775", appRootDirectory+"/bootstrap/cache")

	// ---------------------------------------------------------
	// 7. Composer install if vendor missing
	// ---------------------------------------------------------
	if _, err := os.Stat(filepath.Join(appRootDirectory, "vendor")); os.IsNotExist(err) {
		if err := run("composer", "install", "--optimize-autoloader", "--no-dev", "--working-dir="+appRootDirectory); err != nil {
			return fmt.Errorf("composer install failed: %v", err)
		}
	}

	// ---------------------------------------------------------
	// 8. Artisan commands (idempotent)
	// ---------------------------------------------------------
	envFile := filepath.Join(appRootDirectory, ".env")
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// copy .env.example if .env missing
		exampleEnv := filepath.Join(appRootDirectory, ".env.example")
		if _, err := os.Stat(exampleEnv); err == nil {
			run("cp", exampleEnv, envFile)
		}
	}

	// key:generate only if APP_KEY is empty
	run("php", filepath.Join(appRootDirectory, "artisan"), "key:generate")
	run("php", filepath.Join(appRootDirectory, "artisan"), "migrate", "--force")
	run("php", filepath.Join(appRootDirectory, "artisan"), "db:seed", "--force")
	run("php", filepath.Join(appRootDirectory, "artisan"), "storage:link")
	run("php", filepath.Join(appRootDirectory, "artisan"), "config:cache")
	run("php", filepath.Join(appRootDirectory, "artisan"), "route:cache")
	run("php", filepath.Join(appRootDirectory, "artisan"), "view:cache")

	// ---------------------------------------------------------
	// 9. Nginx config (back up if exists)
	// ---------------------------------------------------------
	nginxConfig := fmt.Sprintf("/etc/nginx/sites-available/%s", appName)
	if _, err := os.Stat(nginxConfig); err == nil {
		run("mv", nginxConfig, nginxConfig+".bak")
	}

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
`, domain, domain, appRootDirectory, phpVersion)

	// Write using sudo tee
	cmd := exec.Command("sudo", "tee", nginxConfig)
	cmd.Stdin = strings.NewReader(configContent)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write nginx config: %w", err)
	}
	run("ln", "-sf", nginxConfig, "/etc/nginx/sites-enabled/")
	run("nginx", "-t")
	run("systemctl", "reload", "nginx")

	// ---------------------------------------------------------
	// 10. Certbot SSL
	// ---------------------------------------------------------
	run("apt", "install", "-y", "certbot", "python3-certbot-nginx")
	run("certbot", "--nginx", "-d", domain, "-d", "www."+domain,
		"--non-interactive", "--agree-tos", "-m", "admin@"+domain)
	run("systemctl", "enable", "certbot.timer")
	run("systemctl", "start", "certbot.timer")

	fmt.Printf("\nDeployment completed for %s at %s!\n", appName, appRootDirectory)
	return nil
}

func RequiredPHPVersion(projectDir string) (string, error) {
	composerPath := filepath.Join(projectDir, "composer.json")

	// 1. Check composer.json
	if _, err := os.Stat(composerPath); err == nil {
		data, err := os.ReadFile(composerPath)
		if err != nil {
			return "", fmt.Errorf("failed to read composer.json: %w", err)
		}

		var composer ComposerJSON
		if err := json.Unmarshal(data, &composer); err != nil {
			return "", fmt.Errorf("invalid composer.json: %w", err)
		}

		// Extract PHP version
		if phpVersion, ok := composer.Require["php"]; ok {
			return phpVersion, nil
		}
	}

	// 2. Search WordPress `$required_php_version`
	wpLoadPath := filepath.Join(projectDir, "wp-includes", "load.php")
	if _, err := os.Stat(wpLoadPath); err == nil {
		data, _ := os.ReadFile(wpLoadPath)
		re := regexp.MustCompile(`\$required_php_version\s*=\s*'([^']+)'`)
		match := re.FindStringSubmatch(string(data))
		if len(match) > 1 {
			return match[1], nil
		}
	}

	// 3. Search CodeIgniter VERSION
	ciConst := filepath.Join(projectDir, "application", "config", "constants.php")
	if _, err := os.Stat(ciConst); err == nil {
		data, _ := os.ReadFile(ciConst)
		re := regexp.MustCompile(`CI_VERSION',\s*'([^']+)'`)
		match := re.FindStringSubmatch(string(data))
		if len(match) > 1 {
			ciVersion := match[1]
			return mapCItoPHP(ciVersion), nil
		}
	}

	// Nothing found
	return "", fmt.Errorf("could not detect PHP version")
}

func mapCItoPHP(ciVersion string) string {
	// Basic mapping
	switch {
	case ciVersion >= "4.0.0":
		return ">=7.4"
	case ciVersion >= "3.1.0":
		return ">=5.6"
	default:
		return ">=5.2"
	}
}

// parsePHPVersion extracts PHP version (major.minor) from a string.
// It works for composer constraints like "^8.1", "8.1.*" or php -v output like "PHP 8.3.2".
func parsePHPVersion(s string) string {
	// Remove any leading caret "^" or tilde "~"
	s = strings.TrimPrefix(s, "^")
	s = strings.TrimPrefix(s, "~")

	// Find first substring that looks like major.minor, e.g., "8.1" or "7.4"
	parts := strings.Fields(s)
	for _, part := range parts {
		// Remove any trailing ".*" or similar
		part = strings.TrimSuffix(part, ".*")

		// Check if it matches 7.x or 8.x pattern
		if len(part) >= 3 && (strings.HasPrefix(part, "7.") || strings.HasPrefix(part, "8.")) {
			pv := strings.Split(part, ".")
			if len(pv) >= 2 {
				return pv[0] + "." + pv[1] // major.minor
			}
			return part
		}
	}

	// fallback: check for "7.x" or "8.x" anywhere in string
	for _, c := range strings.Split(s, ".") {
		if c == "7" || c == "8" {
			return c + ".0"
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
