package models

import (
	"database/sql"
	"time"
)

// TableMeta holds metadata info for each table.
type TableMeta struct {
	TableName      string
	Engine         sql.NullString
	RowFormat      sql.NullString
	TableCollation sql.NullString
	TableRows      sql.NullInt64
	DataLengthMB   sql.NullFloat64
	IndexLengthMB  sql.NullFloat64
	CreateTime     sql.NullTime
	TableComment   sql.NullString
}

type DatabaseMeta struct {
	DBName      string     `json:"db_name"`
	TableCount  int        `json:"table_count"`
	DataSizeMB  float64    `json:"data_size_mb"`
	IndexSizeMB float64    `json:"index_size_mb"`
	CreateTime  *time.Time `json:"create_time,omitempty"` // MySQL doesn’t store DB create time directly
	LastUpdate  *time.Time `json:"last_update,omitempty"` // Latest table update
	Users       []string   `json:"users,omitempty"`       // Users with privileges (optional)
}
