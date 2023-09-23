package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/arena"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (d DB) UpdateArenas(ctx context.Context, arenas []arena.ArenaUpdate) ([]api.Arena, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.UpdateArenas")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start db transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	insertArena := `
		INSERT INTO nba.arena
			as a(name, city, state, country, nba_arena_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (nba_arena_id) DO UPDATE
		SET
			name = coalesce(excluded.name, a.name),
			city = coalesce(excluded.city, a.city),
			state = coalesce(excluded.state, a.state),
			country = coalesce(excluded.country, a.country),
			nba_arena_id = coalesce(excluded.nba_arena_id, a.nba_arena_id)
		RETURNING a.id, a.name, a.city, a.state, a.country, a.nba_arena_id, a.created_at, a.updated_at`

	bp := &pgx.Batch{}

	for _, arena := range arenas {
		bp.Queue(insertArena,
			arena.Name,
			arena.City,
			arena.State,
			arena.Country,
			arena.NBAArenaID)
	}

	batchResults := tx.SendBatch(ctx, bp)

	insertedArenas := []api.Arena{}

	for _, _ = range arenas {
		a := api.Arena{}

		err := batchResults.QueryRow().Scan(&a.ID,
			&a.Name,
			&a.City,
			&a.State,
			&a.Country,
			&a.NBAArenaID,
			&a.CreatedAt,
			&a.UpdatedAt)

		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedArenas = append(insertedArenas, a)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, fmt.Errorf("could not close batchResults when updating arenas: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not commit transaction when updating arenas: %w", err)
	}

	return insertedArenas, nil
}
