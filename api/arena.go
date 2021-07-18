package api

import "time"

type Arena struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	City      *string    `json:"city"`
	State     *string    `json:"state"`
	Country   string     `json:"country"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
