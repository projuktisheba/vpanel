package utils

import (
	"os"
	"path/filepath"

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
	return filepath.Join(homeDir, "projuktisheba", "templates", "projects", lang, projectFramework, projectName, projectName+extension)
}

func GetProjectDirectory(domainName string) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "projuktisheba", "bin", domainName)
}
