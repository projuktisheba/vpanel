package ssl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/projuktisheba/vpanel/backend/internal/pkg/syscmd"
)

// --------------------
// Local Certificate Check
// --------------------

// sslExistsAtPath checks if both PEM cert and key files exist and are non-empty
func sslExistsAtPath(certPath, keyPath string) bool {
	if stat, err := os.Stat(certPath); err != nil || stat.Size() < 10 {
		return false
	}
	if stat, err := os.Stat(keyPath); err != nil || stat.Size() < 10 {
		return false
	}
	return true
}

// sslExistsP12 checks if a PKCS#12 (.p12/.pfx) file exists and is non-empty
func sslExistsP12(path string) bool {
	if stat, err := os.Stat(path); err != nil || stat.Size() < 10 {
		return false
	}
	return true
}

// sslExistsLocal checks standard local paths for domain certificates
func sslExistsLocal(domain string) bool {
	possiblePaths := []struct{ cert, key string }{
		{filepath.Join("/etc/letsencrypt/live", domain, "fullchain.pem"),
			filepath.Join("/etc/letsencrypt/live", domain, "privkey.pem")},
		{filepath.Join("/etc/ssl", domain, "cert.pem"),
			filepath.Join("/etc/ssl", domain, "key.pem")},
		{filepath.Join("/home/user/certs", domain, "cert.pem"),
			filepath.Join("/home/user/certs", domain, "key.pem")},
	}

	for _, p := range possiblePaths {
		if sslExistsAtPath(p.cert, p.key) {
			return true
		}
	}

	// Check P12/PFX
	p12Paths := []string{
		filepath.Join("/etc/ssl", domain, "cert.p12"),
		filepath.Join("/etc/ssl", domain, "cert.pfx"),
		filepath.Join("/home/user/certs", domain, "cert.p12"),
		filepath.Join("/home/user/certs", domain, "cert.pfx"),
	}
	for _, path := range p12Paths {
		if sslExistsP12(path) {
			return true
		}
	}

	return false
}

// --------------------
// Remote Certificate Check
// --------------------

// sslExistsRemote checks if a domain has an SSL certificate on port 443
func sslExistsRemote(domain string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use "echo | openssl s_client" to prevent hanging stdin
	cmd := exec.CommandContext(ctx, "sh", "-c",
		fmt.Sprintf("echo | openssl s_client -connect %s:443 -servername %s -showcerts", domain, domain),
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return false
	}

	output := out.String()
	if strings.Contains(output, "-----BEGIN CERTIFICATE-----") {
		return true
	}

	return false
}

// --------------------
// Automatic Decision
// --------------------

// CheckSSL determines if SSL exists locally or remotely.
// Returns true if SSL exists, false if issuance is needed.
func CheckSSL(domain string) (sslExists bool) {
	// 1. Check local files first
	if sslExistsLocal(domain) {
		return true
	}

	// 2. Check remote SSL
	timeout := 2 * time.Minute
	if sslExistsRemote(domain, timeout) {
		return true
	}
	//SSL record doesn't exist
	return false
}


// SetupSSL sets up SSL for the given domain using Certbot.
// useNginx: true = use nginx plugin, false = standalone
func SetupSSL(ctx context.Context, domain, email string, useNginx bool) error {
	if domain == "" || email == "" {
		return errors.New("domain and email are required")
	}

	// ----------------------------------------
	// 1. CHECK & INSTALL CERTBOT IF NEEDED
	// ----------------------------------------
	fmt.Println("Checking Certbot installation...")

	if _, err := exec.LookPath("certbot"); err != nil {
		fmt.Println("Certbot not found. Installing...")

		if err := syscmd.RunSudoCmd(ctx, "apt", "update"); err != nil {
			return fmt.Errorf("failed apt update: %v", err)
		}

		if useNginx {
			if err := syscmd.RunSudoCmd(ctx, "apt", "install", "-y", "certbot", "python3-certbot-nginx"); err != nil {
				return fmt.Errorf("failed to install certbot with nginx: %v", err)
			}
		} else {
			if err := syscmd.RunSudoCmd(ctx, "apt", "install", "-y", "certbot"); err != nil {
				return fmt.Errorf("failed to install certbot: %v", err)
			}
		}
	}

	// ----------------------------------------
	// 2. RUN CERTBOT
	// ----------------------------------------
	fmt.Println("Starting SSL setup for domain:", domain)

	args := []string{}
	if useNginx {
		args = []string{
			"--nginx",
			"-d", domain,
			"-d", "www." + domain,
			"--non-interactive",
			"--agree-tos",
			"--email", email,
		}
	} else {
		args = []string{
			"certonly",
			"--standalone",
			"--non-interactive",
			"--agree-tos",
			"--email", email,
			"-d", domain,
		}
	}

	// Run certbot via sudo and stream output
	if err := syscmd.RunSudoCmd(ctx, "certbot", args...); err != nil {
		return fmt.Errorf("certbot failed: %v", err)
	}

	// ----------------------------------------
	// 3. RELOAD NGINX IF USED
	// ----------------------------------------
	if useNginx {
		if err := syscmd.RunSudoCmd(ctx, "systemctl", "reload", "nginx"); err != nil {
			return fmt.Errorf("failed to reload nginx: %v", err)
		}
	}

	fmt.Println("\n------------------------------------------------")
	fmt.Println("SSL Certificate obtained successfully!")
	fmt.Printf("Cert Path: /etc/letsencrypt/live/%s/fullchain.pem\n", domain)
	fmt.Printf("Key Path : /etc/letsencrypt/live/%s/privkey.pem\n", domain)
	fmt.Println("------------------------------------------------")

	return nil
}

