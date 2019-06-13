package main

import (
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/pkg/gfunctions/gt"
)

func main() {
	gt.CreateGameThread(nba.MinnesotaTimberwolves)
}
