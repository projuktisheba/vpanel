package models

import "time"

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
