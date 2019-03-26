package main

import (
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/pkg/pgt"
)

func main() {
	pgt.CreatePostGameThread(nba.MinnesotaTimberwolves)
	pgt.CreatePostGameThread(nba.MilwaukeeBucks)
}
