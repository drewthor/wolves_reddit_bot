package gcloud

import (
	"context"
	"log"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/pkg/gfunctions"
)

type event struct {
	Data []byte
}

func Receive(ctx context.Context, e event) error {
	gfunctions.CreatePostGameThread(nba.MilwaukeeBucks)
	gfunctions.CreatePostGameThread(nba.OklahomaCityThunder)
	gfunctions.CreatePostGameThread(nba.PortlandTrailblazers)
	gfunctions.CreatePostGameThread(nba.HoustonRockets)
	gfunctions.CreatePostGameThread(nba.TorontoRaptors)
	log.Printf("ran post game thread checker")
	return nil
}
