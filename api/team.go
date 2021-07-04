package api

import "time"

type Team struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Nickname      string     `json:"nickname"`
	City          string     `json:"city"`
	AlternateCity string     `json:"alternate_city"`
	State         *string    `json:"state"`
	Country       *string    `json:"country"`
	League        string     `json:"league"`
	Season        string     `json:"season"`
	Conference    string     `json:"conference"`
	Division      string     `json:"division"`
	NBAURLName    string     `json:"nba_url_name"`
	NBAShortName  string     `json:"nba_short_name"`
	NBATeamID     string     `json:"nba_team_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}
