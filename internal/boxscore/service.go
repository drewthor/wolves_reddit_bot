package boxscore

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/util"
)

type Service interface {
	Get(gameID, gameDate string) (api.Boxscore, error)
}

func NewService() Service {
	return &service{}
}

type service struct{}

func (s *service) Get(gameID, gameDate string) (api.Boxscore, error) {
	// TODO: fix this, this is just straight wrong and will likely panic
	boxscore, err := nba.GetCurrentSeasonBoxscore(context.Background(), cloudflare.Client{}, util.NBAR2Bucket, gameID, gameDate)
	return api.Boxscore{Boxscore: boxscore}, err
}
