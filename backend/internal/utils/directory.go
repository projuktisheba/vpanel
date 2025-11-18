package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/projuktisheba/vpanel/backend/internal/models"
)



func GetTemplateDirectory(projectFramework, projectName string) string {
	lang, ok := models.FrameworkMap[projectFramework]
	if !ok {
		lang = projectFramework
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "projuktisheba", "templates", "projects", lang, projectFramework, projectName)
}
func GetTemplatePath(projectFramework, projectName, extension string) string {
	lang, ok := models.FrameworkMap[projectFramework]
	if !ok {
		lang = projectFramework
	}
	homeDir, _ := os.UserHomeDir()
	templatePath :=filepath.Join(homeDir, "projuktisheba", "templates", "projects", lang, projectFramework, projectName, projectName)
	if strings.TrimSpace(extension) != ""{
		return  templatePath+extension
	}
	return templatePath
}

func GetProjectDirectory(domainName string) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "projuktisheba", "bin", domainName)
}
// GetWordpressProjectDirectory returns the full path for a WordPress project
// based on the provided domain name.
// It also empty string when error occurs
func GetWordpressProjectDirectory(domainName string) string {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return ""
    }
    return filepath.Join(homeDir, "projuktisheba", "bin", "wordpress", domainName)
}
// GetWordpressProjectName returns a unique name for a WordPress project
// based on the provided domain name.
// It also empty string when error occurs
func GetWordpressProjectName(domainName string) string {
	// trim leading and trailing dots
	domainName = strings.Trim(domainName, ".")
    // replace the dots with underscore (example: subdomain.domain.com will be subdomain_domain_com)
    return strings.ReplaceAll(domainName, ".", "_")
}
