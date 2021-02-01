package api

import "github.com/drewthor/wolves_reddit_bot/apis/nba"

type Boxscore struct {
	Boxscore nba.Boxscore `json:"boxscore"`
}
