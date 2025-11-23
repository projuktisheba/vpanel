package utils

import (
	"net"
)

/**
 * isDomainConnectedToIP looks up the IP addresses for a domain and checks
 * if any of the resolved IPv4 addresses match the targetedIPV4 string.
 *
 * @param domain The hostname or domain name (e.g., "google.com").
 * @param targetedIPV4 The specific IPv4 address to check against.
 * @return bool True if the domain resolves to the targeted IP, false otherwise.
 * @return error Returns an error if the DNS lookup fails.
 */
func IsDomainConnectedToIP(domain string, targetedIPV4 string) bool {
	// 1. Perform DNS resolution for both IPv4 and IPv6
	ips, err := net.LookupIP(domain)
	if err != nil {
		// Return the error from the network operation
		return false
	}

	// 2. Iterate through resolved IPs and check for a match
	for _, ip := range ips {
		// ip.To4() returns a non-nil value only if the IP address is an IPv4 address.
		// We only want to compare IPv4 addresses.
		if ip.To4() != nil {
			// Compare the string representation of the resolved IPv4 with the targeted IP
			if ip.String() == targetedIPV4 {
				// Found a match
				return true
			}
		}
	}

	// 3. If the loop finishes without finding a match
	return false
}
