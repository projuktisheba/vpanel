package models

import (
	"time"
)

const (
	// Technical Status : User-Friendly Description
	StatusInit         = "Initialized"       // "Preparing your project"
	StatusFileUploaded = "Files Uploaded"    // "Your files have been received"
	StatusBuild        = "Building"          // "We are setting up your project"
	StatusRunning      = "Running"           // "Your project is live and running"
	StatusSuspended    = "Suspended"         // "Your project is temporarily paused"
	StatusClosed       = "Closed"            // "Your project has been closed"
	StatusError        = "Error"             // "Something went wrong, please check"
	StatusDeploying    = "Deploying"         // "Project is being deployed"
	StatusCompleted    = "Completed"         // "Project setup is complete"
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
