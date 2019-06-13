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
	gt.CreateGameThread(nba.GoldenStateWarriors)
	pgt.CreatePostGameThread(nba.GoldenStateWarriors)
	log.Printf("ran post game thread checker")
	return nil
}
