package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/referee"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (d DB) UpdateReferees(ctx context.Context, refereeUpdates []referee.RefereeUpdate) ([]api.Referee, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.UpdateReferees")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction with error: %w", err)
	}
	defer tx.Rollback(ctx)

	insertReferee := `
		INSERT INTO nba.referee
			as r(first_name, last_name, jersey_number, nba_referee_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (nba_referee_id) DO UPDATE
		SET 
			first_name = coalesce(excluded.first_name, r.first_name),
			last_name = coalesce(excluded.last_name, r.last_name),
			jersey_number = coalesce(excluded.jersey_number, r.jersey_number),
			nba_referee_id = coalesce(excluded.nba_referee_id, r.nba_referee_id)
		RETURNING r.id, r.first_name, r.last_name, r.jersey_number, r.nba_referee_id, r.created_at, r.updated_at`

	bp := &pgx.Batch{}

	for _, refereeUpdate := range refereeUpdates {
		bp.Queue(insertReferee,
			refereeUpdate.FirstName,
			refereeUpdate.LastName,
			refereeUpdate.JerseyNumber,
			refereeUpdate.NBARefereeID)
	}

	batchResults := tx.SendBatch(ctx, bp)

	insertedReferees := []api.Referee{}

	for _, _ = range insertedReferees {
		referee := api.Referee{}

		err := batchResults.QueryRow().Scan(
			&referee.ID,
			&referee.FirstName,
			&referee.LastName,
			&referee.JerseyNumber,
			&referee.NBARefereeID,
			&referee.CreatedAt,
			&referee.UpdatedAt)

		if err != nil {
			return nil, err
		}

		insertedReferees = append(insertedReferees, referee)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return insertedReferees, nil
}
