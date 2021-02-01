package services

import (
	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
)

type GameService struct{}

func (gs GameService) Get(gameDate string) api.Games {
	gameScoreboards := nba.GetGameScoreboards(gameDate)
	return api.Games{Games: gameScoreboards.Games}
}
