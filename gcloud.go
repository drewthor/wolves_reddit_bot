package gcloud

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/pkg/gfunctions/gt"
	"github.com/drewthor/wolves_reddit_bot/pkg/gfunctions/pgt"
)

type event struct {
	Data []byte
}

func Receive(ctx context.Context, e event) error {
	// google cloud scheduler follows cron which has a limitation of minute granularity
	// manufacture more granular control via sleeping; cloud scheduler triggers this event every
	// minute so we will run 4 times every trigger with 15 second pauses in-between
	var wg sync.WaitGroup
	wg.Add(2)
	go gt.CreateGameThread(nba.MinnesotaTimberwolves, &wg)
	go pgt.CreatePostGameThread(nba.MinnesotaTimberwolves, &wg)

	time.Sleep(15 * time.Second)
	wg.Add(2)
	go gt.CreateGameThread(nba.MinnesotaTimberwolves, &wg)
	go pgt.CreatePostGameThread(nba.MinnesotaTimberwolves, &wg)

	time.Sleep(15 * time.Second)
	wg.Add(2)
	go gt.CreateGameThread(nba.MinnesotaTimberwolves, &wg)
	go pgt.CreatePostGameThread(nba.MinnesotaTimberwolves, &wg)

	time.Sleep(15 * time.Second)
	wg.Add(2)
	go gt.CreateGameThread(nba.MinnesotaTimberwolves, &wg)
	go pgt.CreatePostGameThread(nba.MinnesotaTimberwolves, &wg)

	wg.Wait()
	log.Printf("ran post game thread checker")
	return nil
}
