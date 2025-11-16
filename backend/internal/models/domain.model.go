package models

import "time"

type Domain struct {
	ID             int64     `json:"id"`
	Domain         string    `json:"domain"`
	DomainProvider string    `json:"domainProvider"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
