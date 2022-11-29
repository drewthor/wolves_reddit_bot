package referee

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
)

type Store interface {
	UpdateReferees(ctx context.Context, refereeUpdates []RefereeUpdate) ([]api.Referee, error)
}

type RefereeUpdate struct {
	NBARefereeID int
	FirstName    string
	LastName     string
	JerseyNumber int
}
