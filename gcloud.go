package gcloud

import (
	"context"
	"log"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/pkg/gfunctions/gt"
	"github.com/drewthor/wolves_reddit_bot/pkg/gfunctions/pgt"
)

type event struct {
	Data []byte
}

func Receive(ctx context.Context, e event) error {
	//var wg sync.WaitGroup
	//wg.Add(2)
	gt.CreateGameThread(nba.MinnesotaTimberwolves /*, &wg*/)
	pgt.CreatePostGameThread(nba.MinnesotaTimberwolves /*, &wg*/)
	//wg.Wait()
	log.Printf("ran post game thread checker")
	return nil
}
