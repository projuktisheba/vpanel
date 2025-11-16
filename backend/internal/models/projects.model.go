package models

import "time"

type Project struct {
	ID               int64     `json:"id"`
	ProjectName      string    `json:"projectName"`
	DomainID         int64     `json:"domainId"`
	DatabaseID       int64     `json:"databaseId"`
	Status           string    `json:"status"`           // live, inactive, development
	ProjectFramework string    `json:"projectFramework"` // Laravel, Angular, etc.
	RootDirectory    string    `json:"rootDirectory"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Domain           *Domain   `json:"domain"`
	Database         *Database `json:"database"`
}
