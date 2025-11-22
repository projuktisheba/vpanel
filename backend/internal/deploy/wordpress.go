package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/projuktisheba/vpanel/backend/internal/config"
	"github.com/projuktisheba/vpanel/backend/internal/pkg/ssl"
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

	// 7 Detect PHP-FPM socket
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

	// 8 Create Nginx config
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

	// 9 Install Certbot and obtain SSL
	fmt.Println("Installing obtaining SSL...")
	ssl.SetupSSL(context.Background(), domain, config.Email, true)

	if err := exec.Command("sudo", "systemctl", "reload", "nginx").Run(); err != nil {
		return err
	}

	fmt.Println("============================================")
	fmt.Println("WordPress deployed successfully with SSL!")
	fmt.Printf("Domain: https://%s\n", domain)
	fmt.Printf("Folder: %s\n", wpPath)
	fmt.Printf("PHP Version Used: %s\n", phpVer)
	fmt.Printf("Nginx Config: %s\n", nginxConfName)
	fmt.Println("REMINDER: Create your MySQL database and user manually.")
	fmt.Println("============================================")

	return nil
}

// SuspendWordpressSite creates a temporary Nginx block that includes the necessary SSL paths
// extracted from the original configuration file to pass the 'nginx -t' test.
func SuspendWordpressSite(domain string) error {
	siteAvailablePath := "/etc/nginx/sites-available/"
	siteEnabledPath := "/etc/nginx/sites-enabled/"

	originalConfPath := siteAvailablePath + fmt.Sprintf("%s.conf", domain)
	suspendedConfName := fmt.Sprintf("%s.conf.suspended", domain)
	suspendedConfPath := siteAvailablePath + suspendedConfName
	symlinkPath := siteEnabledPath + fmt.Sprintf("%s.conf", domain)

	// --- 1. Extract SSL Configuration from the Original File ---
	// Command: sudo grep -E 'ssl_certificate|ssl_certificate_key' /path/to/original.conf
	
	fmt.Printf("Extracting SSL paths from: %s\n", originalConfPath)
	
	sslCmd := exec.Command("sudo", "grep", "-E", "ssl_certificate|ssl_certificate_key", originalConfPath)
	
	// Use CombinedOutput for simplicity, but only stderr/err is expected on failure.
	output, err := sslCmd.CombinedOutput()
	
	var sslBlock string
	if err != nil {
		// If grep fails (e.g., file not found, or command fails), we assume the original site 
		// didn't have an SSL block and proceed without SSL lines.
		// However, since we know this site uses SSL, we return a detailed error.
		if strings.Contains(string(output), "No such file or directory") {
			return fmt.Errorf("original Nginx config %s not found", originalConfPath)
		}
		// If grep succeeds, output contains the lines. If it fails due to a syntax error 
		// or permission error, we catch it here.
		fmt.Printf("Warning: Failed to grep SSL block, assuming plain HTTP suspend. Error: %v, output: %s\n", err, output)
        sslBlock = "" // Proceed without SSL block if grep fails (less secure, but handles errors)
	} else {
		// If grep succeeds, the output contains the necessary lines.
		sslBlock = string(output)
	}
	
	// --- 2. Define the Suspension Block Content ---
	// We use 'listen 443 ssl;' and inject the extracted SSL paths.
	blockContent := fmt.Sprintf(`
server {
    listen 80;
    listen 443 ssl;
    server_name %s;

%s

    # Explicitly return 403 (Forbidden) to prevent any unwanted redirects.
    return 403 "Site has been temporarily suspended.";
}
`, domain, sslBlock) // Inject the extracted SSL lines here!

    // --- Step A: Write the new suspension block content to a file (using pipe and tee) ---
    // (Retaining the existing logic for writing with sudo tee)
    
    // 1. Create a command to echo the content
    cmdEcho := exec.Command("echo", "-e", blockContent)
    
    // 2. Create a command to use sudo tee to write the file
    cmdTee := exec.Command("sudo", "tee", suspendedConfPath)

    // 3. Pipe the output of 'echo' into 'sudo tee'
    cmdTee.Stdin, _ = cmdEcho.StdoutPipe()
    
    // 4. Start both commands and wait for them to finish
    // ... (Error handling for Start and Wait commands omitted for brevity, 
    //      but use the exact logic from your previous function) ...
    if err := cmdEcho.Start(); err != nil {
        return fmt.Errorf("failed to start echo command: %v", err)
    }
    if err := cmdTee.Start(); err != nil {
        return fmt.Errorf("failed to start sudo tee command: %v", err)
    }
    if err := cmdEcho.Wait(); err != nil {
        return fmt.Errorf("echo command failed: %v", err)
    }
    if err := cmdTee.Wait(); err != nil {
        return fmt.Errorf("sudo tee command failed: %v", err)
    }
    
	// --- Step B: Create the symlink in sites-enabled (sudo ln -s -f) ---
	// Use -f to overwrite any existing symlink.
	if err := exec.Command("sudo","ln", "-s", "-f", suspendedConfPath, symlinkPath).Run(); err != nil {
		return fmt.Errorf("failed to create symlink for suspended block: %v", err)
	}

	// --- Step C: Reload Nginx (sudo systemctl reload nginx) ---
	if err :=  exec.Command("sudo","systemctl", "reload", "nginx").Run(); err != nil {
		return fmt.Errorf("site suspended, but Nginx reload failed: %v", err)
	}

	return nil
}

// RestartWordpressSite performs cleanup, restores the active site symlink, and reloads Nginx.
func RestartWordpressSite(domain string) error {
	siteAvailablePath := "/etc/nginx/sites-available/"
	siteEnabledPath := "/etc/nginx/sites-enabled/"

	suspendedConfName := fmt.Sprintf("%s.conf.suspended", domain)
	suspendedConfPath := siteAvailablePath + suspendedConfName
	actualConfPath := siteAvailablePath + fmt.Sprintf("%s.conf", domain)
	symlinkPath := siteEnabledPath + fmt.Sprintf("%s.conf", domain)
    
    // --- Step 1: Remove the temporary suspension config file (sudo rm) ---
	fmt.Printf("Attempting to remove temporary suspension config: %s\n", suspendedConfPath)
	if err := exec.Command("sudo","rm", suspendedConfPath).Run(); err != nil {
        fmt.Printf("Warning: Failed to remove suspension config (may already be gone): %v\n", err)
	}
    
    // --- NEW STEP: Remove all broken symlinks from sites-enabled ---
    fmt.Println("Cleaning up broken symlinks in sites-enabled...")
    // Command: sudo find /etc/nginx/sites-enabled/ -type l ! -exec test -e {} \; -delete
    // We target only symbolic links (-type l) that do not exist (! -exec test -e)
    
    // Note: The 'find' command can be complex to run via exec.Command. 
    // We use 'sh -c' to ensure the complex shell logic is interpreted correctly.
    cleanupCmd := fmt.Sprintf("find %s -type l ! -exec test -e {} \\; -delete", siteEnabledPath)
    
    // Execute the complex command using sh -c
    if err := exec.Command("sudo", "sh", "-c", cleanupCmd).Run(); err != nil {
        // If find fails, we warn but continue, as the core task is restoration
        fmt.Printf("Warning: Failed to clean up broken symlinks: %v", err)
    }

    // --- Step 2: Restore the symlink for the actual site config (sudo ln -s -f) ---
    // This is the core action: ensure the primary symlink points to the active config.
    fmt.Printf("Attempting to restore actual site symlink: %s -> %s\n", actualConfPath, symlinkPath)
	if err := exec.Command("sudo","ln", "-s", "-f", actualConfPath, symlinkPath).Run(); err != nil {
		return fmt.Errorf("failed to restore actual site symlink: %v", err)
	}

    // --- Step 3: Reload Nginx (sudo systemctl reload nginx) ---
	fmt.Println("Reloading Nginx to apply reactivation...")
	if err := exec.Command("sudo","systemctl", "reload", "nginx").Run(); err != nil {
		return fmt.Errorf("site reactivated, but Nginx reload failed: %v", err)
	}

	return nil
}

// DeleteWordpressSite deletes the project folder and Nginx configuration
func DeleteWordpressSite(projectName, domain string, projectRoot string) error {
    if domain == "" || projectRoot == "" {
        return fmt.Errorf("domain and project root cannot be empty")
    }

    projectRoot = filepath.Clean(projectRoot)
    projectFolder := filepath.Join(projectRoot, domain)

    // Delete project folder using sudo rm -rf
    cmd := exec.Command("sudo", "rm", "-rf", projectFolder)
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("failed to remove project folder: %v | %s", err, string(out))
    }
    fmt.Printf("Deleted project folder: %s\n", projectFolder)

    // Delete Nginx config
    nginxConf := "/etc/nginx/sites-available/" + fmt.Sprintf("%s.conf", projectName)
    cmd = exec.Command("sudo", "rm", "-f", nginxConf)
    cmd.Run()

    // Delete symlink
    symlink := "/etc/nginx/sites-enabled/" + fmt.Sprintf("%s.conf", projectName)
    cmd = exec.Command("sudo", "rm", "-f", symlink)
    cmd.Run()

    // Reload nginx
    cmd = exec.Command("sudo", "nginx", "-t")
    if err := cmd.Run(); err != nil {
        fmt.Println("⚠ Nginx test failed after deletion, please check manually")
    } else {
        exec.Command("sudo", "systemctl", "reload", "nginx").Run()
        fmt.Println("🔄 Nginx reloaded")
    }

    // Delete certbot certificate (optional cleanup)
    certPath1 := "/etc/letsencrypt/live/" + domain
    certPath2 := "/etc/letsencrypt/archive/" + domain
    certPath3 := "/etc/letsencrypt/renewal/" + domain + ".conf"

    exec.Command("sudo", "rm", "-rf", certPath1).Run()
    exec.Command("sudo", "rm", "-rf", certPath2).Run()
    exec.Command("sudo", "rm", "-f", certPath3).Run()

    fmt.Printf("✅ Site %s fully deleted.\n", domain)
    return nil
}

