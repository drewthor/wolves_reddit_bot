package main

import (
	"sync"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/pkg/gfunctions/gt"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go gt.CreateGameThread(nba.MinnesotaTimberwolves, &wg)
	wg.Wait()
}
