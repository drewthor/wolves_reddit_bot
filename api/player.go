package api

import (
	"time"
)

type Player struct {
	ID              string     `json:"id"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Birthdate       *time.Time `json:"birthdate"`
	HeightFeet      *int       `json:"height_feet"`
	HeightInches    *int       `json:"height_inches"`
	HeightMeters    *float64   `json:"height_meters"`
	WeightPounds    *int       `json:"weight_pounds"`
	WeightKilograms *float64   `json:"weight_kilograms"`
	JerseyNumber    *int       `json:"jersey_number"`
	Positions       []string   `json:"positions"`
	Active          bool       `json:"active"`
	YearsPro        *int       `json:"years_pro"`
	NBADebutYear    *int       `json:"nba_debut_year"`
	NBAPlayerID     int        `json:"nba_player_id"`
	Country         *string    `json:"country"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}
