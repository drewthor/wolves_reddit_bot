package season

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type SeasonUpdate struct {
	LeagueID uuid.UUID
}

type SeasonWeekUpdate struct {
	SeasonStartYear int
	StartDate       time.Time
	EndDate         time.Time
}

type Season struct {
	LeagueID uuid.UUID
}

type SeasonWeek struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	SeasonID  uuid.UUID
	StartDate time.Time
	EndDate   time.Time
}

type Store interface {
	UpdateSeasons(ctx context.Context, seasonUpdates []SeasonUpdate) ([]Season, error)
	UpdateSeasonWeeks(ctx context.Context, seasonWeekUpdates []SeasonWeekUpdate) ([]SeasonWeek, error)
}
