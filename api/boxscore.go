package api

import "github.com/drewthor/wolves_reddit_bot/apis/nba"

type Boxscore struct {
	Boxscore nba.BoxscoreOld `json:"boxscore"`
}
