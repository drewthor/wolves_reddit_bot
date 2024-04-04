package api

import "time"

type TeamSeason struct {
	ID           string     `json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	TeamID       string     `json:"team_id"`
	LeagueID     string     `json:"league_id"`
	SeasonID     string     `json:"season_id"`
	ConferenceID *string    `json:"conference_id"`
	DivisionID   *string    `json:"division_id"`
	City         string
	Name         string
}
