package models

import (
	"time"
)

type DBUser struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	UserType  string     `json:"userType"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type Database struct {
	ID             int64      `json:"id"`
	DBName         string     `json:"dbName"`
	DBType         string     `json:"dbType"`
	TableCount     int        `json:"tableCount"`
	DatabaseSizeMB float64    `json:"databaseSizeMB"`
	UserID         int64      `json:"userId"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
	User           *DBUser    `json:"user,omitempty"`
}
