package franchise

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
)

type Store interface {
	UpdateFranchises(ctx context.Context, franchises []FranchiseUpdate) ([]api.Franchise, error)
}

type FranchiseUpdate struct {
	NBALeagueID        string
	NBATeamID          int
	Name               string
	Nickname           string
	City               string
	State              string
	Country            string
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
