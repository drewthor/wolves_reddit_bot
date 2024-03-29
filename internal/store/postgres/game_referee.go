package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/internal/game_referee"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (d DB) UpdateGameReferees(ctx context.Context, gameRefereeUpdates []game_referee.GameRefereeUpdate) ([]game_referee.GameReferee, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.UpdateGameReferees")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction when updating game referees: %w", err)
	}
	defer tx.Rollback(ctx)

	insertGameReferee := `
		INSERT INTO nba.game_referee
			as gr(game_id, referee_id, assignment)
		VALUES ((SELECT id FROM nba.Game WHERE nba_game_id = $1), (SELECT id FROM nba.Referee WHERE nba_referee_id = $2), $3)
		ON CONFLICT (game_id, referee_id) DO UPDATE
		SET 
			game_id = coalesce(excluded.game_id, gr.game_id),
			referee_id = coalesce(excluded.referee_id, gr.referee_id),
			assignment = coalesce(excluded.assignment, gr.assignment)
		RETURNING game_id, referee_id, assignment, created_at, updated_at`

	bp := &pgx.Batch{}

	for _, gameRefereeUpdate := range gameRefereeUpdates {
		bp.Queue(insertGameReferee,
			gameRefereeUpdate.NBAGameID,
			gameRefereeUpdate.NBARefereeID,
			gameRefereeUpdate.Assignment)
	}

	batchResults := tx.SendBatch(ctx, bp)

	insertedGameReferees := []game_referee.GameReferee{}

	for _, _ = range insertedGameReferees {
		gameReferee := game_referee.GameReferee{}

		err := batchResults.QueryRow().Scan(
			&gameReferee.GameID,
			&gameReferee.RefereeID,
			&gameReferee.Assignment,
			&gameReferee.CreatedAt,
			&gameReferee.UpdatedAt)

		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedGameReferees = append(insertedGameReferees, gameReferee)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return insertedGameReferees, nil
}
