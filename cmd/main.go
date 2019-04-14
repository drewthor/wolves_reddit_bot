package main

import (
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/pkg/gfunctions"
)

func main() {
	gfunctions.CreatePostGameThread(nba.MinnesotaTimberwolves)
}
