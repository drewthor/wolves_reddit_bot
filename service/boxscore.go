package service

import (
	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
)

type BoxscoreService struct{}

func (bs BoxscoreService) Get(gameID, gameDate string) (api.Boxscore, error) {
	boxscore, err := nba.GetBoxscore(nba.GetDailyAPIPaths().Boxscore, gameDate, gameID)
	return api.Boxscore{Boxscore: boxscore}, err
}
