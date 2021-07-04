package api

import (
	"time"
)

type Player struct {
	ID              string     `json:"id"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Birthdate       time.Time  `json:"birthdate"`
	HeightFeet      int        `json:"height_feet"`
	HeightInches    int        `json:"height_inches"`
	HeightMeters    float64    `json:"height_meters"`
	WeightPounds    int        `json:"weight_pounds"`
	WeightKilograms float64    `json:"weight_kilograms"`
	JerseyNumber    *int       `json:"jersey_number"`
	Positions       []string   `json:"positions"`
	CurrentlyInNBA  bool       `json:"currently_in_nba"`
	YearsPro        int        `json:"years_pro"`
	NBADebutYear    *int       `json:"nba_debut_year"`
	NBAPlayerID     string     `json:"nba_player_id"`
	Country         string     `json:"country"`
	TimeCreated     time.Time  `json:"time_created"`
	TimeModified    *time.Time `json:"time_modified"`
}
