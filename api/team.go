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
	FranchiseID   *string    `json:"franchise_id"`
	NBAURLName    string     `json:"nba_url_name"`
	NBAShortName  string     `json:"nba_short_name"`
	NBATeamID     int        `json:"nba_team_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}
