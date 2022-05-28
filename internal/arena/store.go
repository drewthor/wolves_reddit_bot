package arena

import (
	"context"
	"database/sql"

	log "github.com/sirupsen/logrus"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Store interface {
	UpdateArenas(arenas []ArenaUpdate) ([]api.Arena, error)
}

func NewStore(db *pgxpool.Pool) Store {
	return &store{DB: db}
}

type store struct {
	DB *pgxpool.Pool
}

type ArenaUpdate struct {
	NBAArenaID int
	Name       string
	City       sql.NullString
	State      sql.NullString
	Country    string
}

func (s *store) UpdateArenas(arenas []ArenaUpdate) ([]api.Arena, error) {
	log.Info("updating arenas")
	tx, err := s.DB.Begin(context.Background())
	defer tx.Rollback(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

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

	log.Info("queued arenas for insert")

	batchResults := tx.SendBatch(context.Background(), bp)

	insertedArenas := []api.Arena{}

	for _, _ = range arenas {
		arena := api.Arena{}

		err := batchResults.QueryRow().Scan(&arena.ID,
			&arena.Name,
			&arena.City,
			&arena.State,
			&arena.Country,
			&arena.NBAArenaID,
			&arena.CreatedAt,
			&arena.UpdatedAt)

		log.WithError(err).Info("inserted arena")
		if err != nil {
			return nil, err
		}

		insertedArenas = append(insertedArenas, arena)
	}

	err = batchResults.Close()
	if err != nil {
		log.WithError(err).Error("could not close batchResults when updating arenas")
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		log.WithError(err).Error("could not commit transaction when updating arenas")
		return nil, err
	}

	return insertedArenas, nil
}
