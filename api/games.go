package api

import "github.com/drewthor/wolves_reddit_bot/apis/nba"

type Games struct {
	Games []nba.GameScoreboard `json:"games"`
}
