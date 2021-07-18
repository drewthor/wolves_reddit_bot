package dao

import (
	"context"
	"database/sql"
	"log"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ArenaDAO struct {
	DB *pgxpool.Pool
}

type ArenaUpdate struct {
	Name    string
	City    sql.NullString
	State   sql.NullString
	Country string
}

func (ad *ArenaDAO) UpdateArenas(arenas []ArenaUpdate) ([]api.Arena, error) {
	tx, err := ad.DB.Begin(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

	insertArena := `
		INSERT INTO nba.arena
			as a(name, city, state, country)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE
		SET
			name = coalesce(excluded.name, a.name),
			city = coalesce(excluded.city, a.city),
			state = coalesce(excluded.state, a.state),
			country = coalesce(excluded.country, a.country)
		RETURNING a.id, a.name, a.city, a.state, a.country, a.created_at, a.updated_at`

	bp := &pgx.Batch{}

	for _, arena := range arenas {
		bp.Queue(insertArena,
			arena.Name,
			arena.City,
			arena.State,
			arena.Country)
	}

	batchResults := tx.SendBatch(context.Background(), bp)

	insertedArenas := []api.Arena{}

	for _, _ = range arenas {
		arena := api.Arena{}

		err := batchResults.QueryRow().Scan(&arena.ID,
			&arena.Name,
			&arena.City,
			&arena.State,
			&arena.Country,
			&arena.CreatedAt,
			&arena.UpdatedAt)

		if err != nil {
			return nil, err
		}

		insertedArenas = append(insertedArenas, arena)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return insertedArenas, nil
}
