package ssl

import (
	"fmt"
	"os"
	"os/exec"
)

// SetupSSL runs the certbot command to obtain a certificate for the given domain.
// It uses the --standalone plugin, which requires port 80 to be free.
func SetupSSL(domain string, email string) error {
	// 1. Check if certbot is installed
	_, err := exec.LookPath("certbot")
	if err != nil {
		return fmt.Errorf("certbot is not installed on this system. Please install it first")
	}

	fmt.Printf("Starting SSL setup for domain: %s...\n", domain)

	// 2. Construct the command
	// certbot certonly --standalone --non-interactive --agree-tos --email user@example.com -d example.com
	cmd := exec.Command("certbot", "certonly",
		"--standalone",      // Use built-in web server for verification (Port 80 must be free)
		"--non-interactive", // Do not ask for user input
		"--agree-tos",       // Agree to Terms of Service
		"--email", email,    // Email for urgent renewal and security notices
		"-d", domain, // The domain name
	)

	// 3. Pipe stdout and stderr to see certbot's output in real-time
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 4. Execute the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run certbot: %v", err)
	}

	// 5. Success message
	fmt.Println("\n------------------------------------------------")
	fmt.Println("SSL Certificate obtained successfully!")
	fmt.Printf("Certificate: /etc/letsencrypt/live/%s/fullchain.pem\n", domain)
	fmt.Printf("Private Key: /etc/letsencrypt/live/%s/privkey.pem\n", domain)
	fmt.Println("------------------------------------------------")

	return nil
}

// func main() {
// 	// Example Usage
// 	// NOTE: This program usually needs to be run with 'sudo' permissions
// 	// to write to /etc/letsencrypt/

// 	if os.Geteuid() != 0 {
// 		fmt.Println("WARNING: This script usually requires root privileges to run Certbot.")
// 	}

// 	domain := "yourdomain.com" // Replace with your domain
// 	email := "you@example.com" // Replace with your email

// 	// Call the setup function
// 	err := SetupSSL(domain, email)
// 	if err != nil {
// 		log.Fatalf("Error setting up SSL: %v", err)
// 	}
// }
