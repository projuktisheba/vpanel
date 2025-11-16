package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DeployLaravelHandler deploys a Laravel app to a custom folder
func DeployLaravelHandler(
	appName, appDir, domain, phpVersion, dbName, dbUser, dbPass, gitRepo, gitBranch, sqlFile string,
) error {
	fmt.Println("Starting Laravel deployment...")

	// Helper to run commands
	run := func(name string, args ...string) error {
		fmt.Printf("Running: %s %s\n", name, strings.Join(args, " "))
		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// 1. Update system
	if err := run("sudo", "apt", "update"); err != nil {
		return err
	}
	if err := run("sudo", "apt", "upgrade", "-y"); err != nil {
		return err
	}

	// 2. Install required packages
	pkgs := []string{"nginx", "mysql-server", "unzip", "git", "curl", "software-properties-common"}

	args := append([]string{"install", "-y"}, pkgs...) // correctly build args

	if err := run("apt", args...); err != nil {
		return err
	}

	// 3. Add PHP PPA and install PHP
	if err := run("sudo", "add-apt-repository", "ppa:ondrej/php", "-y"); err != nil {
		return err
	}
	if err := run("sudo", "apt", "update"); err != nil {
		return err
	}

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
		fmt.Sprintf("php%s-soap", phpVersion),
	}
	args = append([]string{"install", "-y"}, phpPkgs...)
	if err := run("apt", args...); err != nil {
		return err
	}

	// 4. Install Composer
	if err := run("php", "-r", "copy('https://getcomposer.org/installer','composer-setup.php');"); err != nil {
		return err
	}
	if err := run("php", "composer-setup.php", "--install-dir=/usr/local/bin", "--filename=composer"); err != nil {
		return err
	}
	if err := run("php", "-r", "unlink('composer-setup.php');"); err != nil {
		return err
	}

	// 5. Set permissions
	if err := run("sudo", "chown", "-R", "www-data:www-data", appDir); err != nil {
		return err
	}
	if err := run("sudo", "find", appDir, "-type", "f", "-exec", "chmod", "644", "{}", ";"); err != nil {
		return err
	}
	if err := run("sudo", "find", appDir, "-type", "d", "-exec", "chmod", "755", "{}", ";"); err != nil {
		return err
	}
	if err := run("sudo", "chmod", "-R", "775", fmt.Sprintf("%s/storage %s/bootstrap/cache", appDir, appDir)); err != nil {
		return err
	}

	// 8. Composer install
	if err := run("composer", "install", "--optimize-autoloader", "--no-dev"); err != nil {
		return err
	}

	// 9. Configure Laravel artisan commands
	run("php", fmt.Sprintf("%s/artisan", appDir), "key:generate")
	run("php", fmt.Sprintf("%s/artisan", appDir), "migrate", "--force")
	run("php", fmt.Sprintf("%s/artisan", appDir), "db:seed", "--force")
	run("php", fmt.Sprintf("%s/artisan", appDir), "storage:link")
	run("php", fmt.Sprintf("%s/artisan", appDir), "config:cache")
	run("php", fmt.Sprintf("%s/artisan", appDir), "route:cache")
	run("php", fmt.Sprintf("%s/artisan", appDir), "view:cache")

	// 10. Create Nginx config
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
}`, domain, domain, appDir, phpVersion)
	os.WriteFile(nginxConfig, []byte(configContent), 0644)
	run("sudo", "ln", "-sf", nginxConfig, "/etc/nginx/sites-enabled/")
	run("sudo", "nginx", "-t")
	run("sudo", "systemctl", "reload", "nginx")

	fmt.Printf("Deployment completed for %s at %s!\n", appName, appDir)
	return nil
}
