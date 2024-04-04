package api

import "time"

type Franchise struct {
	ID                 string
	CreatedAt          time.Time
	UpdatedAt          *time.Time
	LeagueID           string
	NBATeamID          int
	City               string
	State              string
	Country            string
	Name               string
	Nickname           string
	StartYear          int
	EndYear            int
	Years              int
	Games              int
	Wins               int
	Losses             int
	PlayoffAppearances int
	DivisionTitles     int
	ConferenceTitles   int
	LeagueTitles       int
	Active             bool
}
