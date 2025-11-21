package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func DeployWordPress(domain string, projectRoot string) error {
	if domain == "" || projectRoot == "" {
		return fmt.Errorf("domain and project root cannot be empty")
	}

	// Remove trailing slash
	projectRoot = strings.TrimRight(projectRoot, "/")
	projectFolder := filepath.Join(projectRoot, domain)
	nginxConfName := strings.ReplaceAll(domain, ".", "_") + ".conf"

	fmt.Printf("Project folder: %s\n", projectFolder)
	fmt.Printf("Nginx config filename: %s\n", nginxConfName)

	// 1 Update server
	fmt.Println("Updating server...")
	cmd := exec.Command("sudo", "apt", "update", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// 2 Detect installed PHP version
	fmt.Println("Detecting installed PHP versions...")
	phpDirOutput, err := exec.Command("ls", "/etc/php/").Output()
	if err != nil {
		return err
	}

	phpVersions := strings.Fields(string(phpDirOutput))
	var phpVer string
	if len(phpVersions) == 0 {
		fmt.Println("⚠ No PHP detected. Installing latest PHP...")
		cmds := [][]string{
			{"sudo", "apt", "install", "-y", "software-properties-common"},
			{"sudo", "add-apt-repository", "-y", "ppa:ondrej/php"},
			{"sudo", "apt", "update", "-y"},
			{"sudo", "apt", "install", "-y", "php-fpm", "php-mysql", "php-curl", "php-gd", "php-mbstring", "php-xml", "php-zip", "php-soap", "php-intl"},
		}
		for _, c := range cmds {
			cmd := exec.Command(c[0], c[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
		// Recheck PHP version
		phpDirOutput2, err := exec.Command("ls", "/etc/php/").Output()
		if err != nil {
			return err
		}
		phpVersions2 := strings.Fields(string(phpDirOutput2))
		phpVer = phpVersions2[len(phpVersions2)-1]
		fmt.Printf("✅ Installed PHP %s\n", phpVer)
	} else {
		phpVer = phpVersions[len(phpVersions)-1]
		fmt.Printf("✅ Detected PHP version: %s\n", phpVer)
	}

	// 3 Install Nginx, MySQL client, tools
	fmt.Println("Installing Nginx, MySQL client, unzip, wget, curl...")
	cmds := [][]string{
		{"sudo", "apt", "install", "-y", "nginx", "mysql-client", "unzip", "wget", "curl"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// 4 Prepare project folder
	fmt.Println("Creating project folder...")
	if err := os.MkdirAll(projectFolder, 0755); err != nil {
		return err
	}
	if err := exec.Command("sudo", "chown", "-R", fmt.Sprintf("%s:%s", os.Getenv("USER"), os.Getenv("USER")), projectFolder).Run(); err != nil {
		return err
	}

	// 5 Download WordPress
	fmt.Println("Downloading latest WordPress...")
	tmpZip := "/tmp/wordpress.zip"
	cmd = exec.Command("wget", "-q", "https://wordpress.org/latest.zip", "-O", tmpZip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("unzip", "-q", tmpZip, "-d", "/tmp")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// 6 Backup existing WordPress
	wpPath := filepath.Join(projectFolder, "wordpress")
	if _, err := os.Stat(wpPath); err == nil {
		backupPath := fmt.Sprintf("%s-backup-%d", wpPath, time.Now().Unix())
		fmt.Printf("⚠ Existing WordPress found. Backing up to %s\n", backupPath)
		if err := os.Rename(wpPath, backupPath); err != nil {
			return err
		}
	}

	// Move downloaded WordPress
	if err := os.Rename("/tmp/wordpress", wpPath); err != nil {
		return err
	}

	// 7️⃣ Detect PHP-FPM socket
	socketOutput, err := exec.Command("ls", "/run/php/").Output()
	if err != nil {
		return err
	}
	var phpSock string
	for _, line := range strings.Fields(string(socketOutput)) {
		if strings.HasPrefix(line, fmt.Sprintf("php%s-fpm.sock", phpVer)) {
			phpSock = "/run/php/" + line
			break
		}
	}
	if phpSock == "" {
		return fmt.Errorf("PHP-FPM socket for PHP %s not found", phpVer)
	}
	fmt.Printf("Using PHP-FPM socket: %s\n", phpSock)

	// 8️⃣ Create Nginx config
	nginxConfPath := "/etc/nginx/sites-available/" + nginxConfName
	nginxConf := fmt.Sprintf(`server {
    listen 80;
    server_name %s www.%s;

    root %s;
    index index.php index.html;

	# Allow large file uploads (2048M)
    client_max_body_size 2048M;

    location / {
        try_files $uri $uri/ /index.php?$args;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:%s;
		# Increase timeout for long processes
        # Increase timeouts for long restorations
        fastcgi_read_timeout 1800; 
        fastcgi_send_timeout 1800;
        fastcgi_connect_timeout 1800;
    }

    location ~ /\.ht {
        deny all;
    }
}`, domain, domain, wpPath, phpSock)

	if err := os.WriteFile("/tmp/nginx_tmp.conf", []byte(nginxConf), 0644); err != nil {
		return err
	}
	if err := exec.Command("sudo", "mv", "/tmp/nginx_tmp.conf", nginxConfPath).Run(); err != nil {
		return err
	}

	// Enable site
	if err := exec.Command("sudo", "ln", "-sf", nginxConfPath, "/etc/nginx/sites-enabled/").Run(); err != nil {
		return err
	}

	// Test Nginx and reload
	if err := exec.Command("sudo", "nginx", "-t").Run(); err != nil {
		return fmt.Errorf("nginx test failed: %v", err)
	}
	if err := exec.Command("sudo", "systemctl", "reload", "nginx").Run(); err != nil {
		return err
	}

	// Set permissions
	if err := exec.Command("sudo", "chown", "-R", "www-data:www-data", wpPath).Run(); err != nil {
		return err
	}

	// 9️⃣ Install Certbot and obtain SSL
	fmt.Println("Installing Certbot and obtaining SSL...")
	cmdsCert := [][]string{
		{"sudo", "apt", "install", "-y", "certbot", "python3-certbot-nginx"},
		{"sudo", "certbot", "--nginx", "-d", domain, "-d", "www." + domain, "--non-interactive", "--agree-tos", "-m", "admin@" + domain},
	}
	for _, c := range cmdsCert {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if err := exec.Command("sudo", "systemctl", "reload", "nginx").Run(); err != nil {
		return err
	}

	fmt.Println("============================================")
	fmt.Println("✅ WordPress deployed successfully with SSL!")
	fmt.Printf("🌍 Domain: https://%s\n", domain)
	fmt.Printf("📂 Folder: %s\n", wpPath)
	fmt.Printf("🧩 PHP Version Used: %s\n", phpVer)
	fmt.Printf("📝 Nginx Config: %s\n", nginxConfName)
	fmt.Println("⚠️ REMINDER: Create your MySQL database and user manually.")
	fmt.Println("============================================")

	return nil
}

// SuspendSite temporarily disables a site by commenting out Nginx config
func SuspendSite(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	nginxConf := "/etc/nginx/sites-available/" + fmt.Sprintf("%s.conf", domain)
	if _, err := os.Stat(nginxConf); os.IsNotExist(err) {
		return fmt.Errorf("nginx config for domain %s does not exist", domain)
	}

	// Rename config to indicate suspended
	suspendedConf := nginxConf + ".suspended"
	if err := os.Rename(nginxConf, suspendedConf); err != nil {
		return fmt.Errorf("failed to suspend site: %v", err)
	}

	// Remove symlink in sites-enabled
	symlink := "/etc/nginx/sites-enabled/" + fmt.Sprintf("%s.conf", domain)
	_ = os.Remove(symlink)

	// Reload Nginx
	if err := exec.Command("sudo", "nginx", "-t").Run(); err != nil {
		return fmt.Errorf("nginx test failed: %v", err)
	}
	if err := exec.Command("sudo", "systemctl", "reload", "nginx").Run(); err != nil {
		return fmt.Errorf("failed to reload nginx: %v", err)
	}

	fmt.Printf("✅ Site %s has been suspended.\n", domain)
	return nil
}

// RestartSite re-enables a suspended site
func RestartSite(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	suspendedConf := "/etc/nginx/sites-available/" + fmt.Sprintf("%s.conf.suspended", domain)
	if _, err := os.Stat(suspendedConf); os.IsNotExist(err) {
		return fmt.Errorf("suspended nginx config for domain %s does not exist", domain)
	}

	// Rename back to normal
	nginxConf := "/etc/nginx/sites-available/" + fmt.Sprintf("%s.conf", domain)
	if err := os.Rename(suspendedConf, nginxConf); err != nil {
		return fmt.Errorf("failed to restore site config: %v", err)
	}

	// Recreate symlink in sites-enabled
	symlink := "/etc/nginx/sites-enabled/" + fmt.Sprintf("%s.conf", domain)
	if err := os.Symlink(nginxConf, symlink); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	// Reload Nginx
	if err := exec.Command("sudo", "nginx", "-t").Run(); err != nil {
		return fmt.Errorf("nginx test failed: %v", err)
	}
	if err := exec.Command("sudo", "systemctl", "reload", "nginx").Run(); err != nil {
		return fmt.Errorf("failed to reload nginx: %v", err)
	}

	fmt.Printf("✅ Site %s has been reactivated.\n", domain)
	return nil
}

// DeleteProject deletes the project folder and Nginx configuration
func DeleteProject(domain string, projectRoot string) error {
	if domain == "" || projectRoot == "" {
		return fmt.Errorf("domain and project root cannot be empty")
	}
	projectRoot = filepath.Clean(projectRoot)
	projectFolder := filepath.Join(projectRoot, domain)

	// Remove project folder
	if _, err := os.Stat(projectFolder); !os.IsNotExist(err) {
		if err := os.RemoveAll(projectFolder); err != nil {
			return fmt.Errorf("failed to remove project folder: %v", err)
		}
		fmt.Printf("✅ Deleted project folder: %s\n", projectFolder)
	}

	// Remove Nginx config and symlink
	nginxConf := "/etc/nginx/sites-available/" + fmt.Sprintf("%s.conf", domain)
	symlink := "/etc/nginx/sites-enabled/" + fmt.Sprintf("%s.conf", domain)
	_ = os.Remove(symlink)
	_ = os.Remove(nginxConf)

	// Reload Nginx
	if err := exec.Command("sudo", "nginx", "-t").Run(); err != nil {
		fmt.Println("⚠ Nginx test failed after deletion, please check manually")
	} else {
		_ = exec.Command("sudo", "systemctl", "reload", "nginx").Run()
	}

	// Remove cert records
	fmt.Printf("✅ Site %s fully deleted.\n", domain)
	return nil
}
