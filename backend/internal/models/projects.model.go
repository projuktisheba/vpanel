package models

import (
	"time"
)

const (
	// Technical Status : User-Friendly Description
	ProjectStatusInit         = "Initialized"       // "Preparing your project"
	ProjectStatusFileUploaded = "Files Uploaded"    // "Your files have been received"
	ProjectStatusRunning      = "Running"           // "Your project is live and running"
	ProjectStatusSuspended    = "Suspended"         // "Your project is temporarily paused"
	ProjectStatusClosed       = "Closed"            // "Your project has been closed"
	ProjectStatusError        = "Error"             // "Something went wrong, please check"
	ProjectStatusDeploying    = "Deploying"         // "Project is being deployed"
)


type Project struct {
	ID               int64     `json:"id"`
	ProjectName      string    `json:"projectName"`
	DomainName       string    `json:"domainName"`
	DBName           string    `json:"dbName"`
	ProjectFramework string    `json:"projectFramework"`
	TemplatePath     string    `json:"templatePath"`
	ProjectDirectory string    `json:"projectDirectory"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	DomainInfo       *Domain   `json:"domainInfo"`
	DatabaseInfo     *Database `json:"databaseInfo"`
}
