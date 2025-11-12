package models

import (
	"time"
)

// TableMeta holds metadata info for each table.
type TableMeta struct {
	TableName      string    `json:"tableName"`
	Engine         string    `json:"engine"`
	RowFormat      string    `json:"rowFormat"`
	TableCollation string    `json:"tableCollation"`
	TableRows      int64     `json:"table_rows"`
	DataLengthMB   float64   `json:"dataLengthMB"`
	IndexLengthMB  float64   `json:"indexLength"`
	CreateAt       time.Time `json:"createdAt"`
}

type DatabaseMeta struct {
	DBName         string     `json:"dbName"`
	TableCount     int        `json:"tableCount"`
	DatabaseSizeMB float64    `json:"databaseSizeMB"`
	CreatedAt      *time.Time `json:"createdAt"` // MySQL doesn’t store DB create time directly
	UpdatedAt      *time.Time `json:"updatedAt"` // Latest table update
	Users          []string   `json:"users"`     // Users with privileges (optional)
}
