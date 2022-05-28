package referee

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Store interface {
	UpdateReferees(refereeUpdates []RefereeUpdate) ([]api.Referee, error)
}

func NewStore(db *pgxpool.Pool) Store {
	return &store{DB: db}
}

type store struct {
	DB *pgxpool.Pool
}

type RefereeUpdate struct {
	NBARefereeID int
	FirstName    string
	LastName     string
	JerseyNumber int
}

func (s *store) UpdateReferees(refereeUpdates []RefereeUpdate) ([]api.Referee, error) {
	tx, err := s.DB.Begin(context.Background())
	defer tx.Rollback(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

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

	batchResults := tx.SendBatch(context.Background(), bp)

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

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return insertedReferees, nil
}
