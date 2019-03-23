package gcloud

import (
	"context"
	"github.com/drewthor/wolves_reddit_bot/pkg/pgt"
	"log"
)

type event struct {
	Data []byte
}

func Receive(ctx context.Context, e event) error {
	pgt.CreatePostGameThread()
	log.Printf("ran post game thread checker")
	return nil
}
