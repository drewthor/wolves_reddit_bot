package referee

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
	"go.opentelemetry.io/otel"
)

type Service interface {
	UpdateReferees(ctx context.Context, refereeUpdates []RefereeUpdate) ([]api.Referee, error)
}

func NewService(refereeStore Store) Service {
	return &service{RefereeStore: refereeStore}
}

type service struct {
	RefereeStore Store
}

func (s service) UpdateReferees(ctx context.Context, refereeUpdates []RefereeUpdate) ([]api.Referee, error) {
	ctx, span := otel.Tracer("referee").Start(ctx, "referee.service.UpdateReferees")
	defer span.End()

	return s.RefereeStore.UpdateReferees(ctx, refereeUpdates)
}
