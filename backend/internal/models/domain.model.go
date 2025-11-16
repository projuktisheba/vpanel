package models

import "time"

type Domain struct {
	ID            int64      `json:"id"`
	Domain        string     `json:"domain"`
	SSLUpdateDate *time.Time `json:"ssl_update_date"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
