package dao

import (
	"context"
	"log"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type RefereeDAO struct {
	DB *pgxpool.Pool
}

type RefereeUpdate struct {
	FirstName string
	LastName  string
}

func (rd *RefereeDAO) UpdateReferees(refereeUpdates []RefereeUpdate) ([]api.Referee, error) {
	tx, err := rd.DB.Begin(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

	insertReferee := `
		INSERT INTO nba.referee
			as r(first_name, last_name)
		VALUES ($1, $2)
		ON CONFLICT (first_name, last_name) DO UPDATE
		SET 
			first_name = coalesce(excluded.first_name, r.first_name),
			last_name = coalesce(excluded.last_name, r.last_name)
		RETURNING r.id, r.first_name, r.last_name, r.created_at, r.updated_at`

	bp := &pgx.Batch{}

	for _, refereeUpdate := range refereeUpdates {
		bp.Queue(insertReferee,
			refereeUpdate.FirstName,
			refereeUpdate.LastName)
	}

	batchResults := tx.SendBatch(context.Background(), bp)

	insertedReferees := []api.Referee{}

	for _, _ = range insertedReferees {
		referee := api.Referee{}

		err := batchResults.QueryRow().Scan(
			&referee.ID,
			&referee.FirstName,
			&referee.LastName,
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
