package api

import "time"

type Referee struct {
	ID           string     `json:"id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	JerseyNumber int        `json:"jersey_number"`
	NBARefereeID int        `json:"nba_referee_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}
