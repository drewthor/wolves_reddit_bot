package gcloud

import (
	"context"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/pkg/gfunctions"
	"log"
)

type event struct {
	Data []byte
}

func Receive(ctx context.Context, e event) error {
	gfunctions.CreatePostGameThread(nba.MilwaukeeBucks)
	log.Printf("ran post game thread checker")
	return nil
}
