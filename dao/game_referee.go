package dao

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type GameRefereeDAO struct {
	DB *pgxpool.Pool
}

type GameRefereeUpdate struct {
	NBAGameID string
	FirstName string
	LastName  string
}

type GameReferee struct {
	GameID    string
	RefereeID string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

func (grd *GameRefereeDAO) UpdateGameReferees(gameRefereeUpdates []GameRefereeUpdate) ([]GameReferee, error) {
	tx, err := grd.DB.Begin(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

	insertGameReferee := `
		INSERT INTO nba.game_referee
			as gr(game_id, referee_id)
		VALUES ((SELECT id FROM nba.Game WHERE nba_game_id = $1), (SELECT id FROM nba.Referee WHERE first_name = $2 AND last_name = $3))
		ON CONFLICT (game_id, referee_id) DO UPDATE
		SET 
			game_id = coalesce(excluded.game_id, gr.game_id),
			referee_id = coalesce(excluded.referee_id, gr.referee_id)
		RETURNING game_id, referee_id, created_at, updated_at`

	bp := &pgx.Batch{}

	for _, gameRefereeUpdate := range gameRefereeUpdates {
		bp.Queue(insertGameReferee,
			gameRefereeUpdate.NBAGameID,
			gameRefereeUpdate.FirstName,
			gameRefereeUpdate.LastName)
	}

	batchResults := tx.SendBatch(context.Background(), bp)

	insertedGameReferees := []GameReferee{}

	for _, _ = range insertedGameReferees {
		gameReferee := GameReferee{}

		err := batchResults.QueryRow().Scan(
			&gameReferee.GameID,
			&gameReferee.RefereeID,
			&gameReferee.CreatedAt,
			&gameReferee.UpdatedAt)

		if err != nil {
			return nil, err
		}

		insertedGameReferees = append(insertedGameReferees, gameReferee)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return insertedGameReferees, nil
}
