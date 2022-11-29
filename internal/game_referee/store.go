package game_referee

import (
	"context"
	"time"
)

type Store interface {
	UpdateGameReferees(ctx context.Context, gameRefereeUpdates []GameRefereeUpdate) ([]GameReferee, error)
}

type GameRefereeUpdate struct {
	NBAGameID    string
	NBARefereeID int
	Assignment   string
}

type GameReferee struct {
	GameID     string
	RefereeID  string
	Assignment string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}
