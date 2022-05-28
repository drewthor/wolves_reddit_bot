package game_referee

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Store interface {
	UpdateGameReferees(gameRefereeUpdates []GameRefereeUpdate) ([]GameReferee, error)
}

func NewStore(db *pgxpool.Pool) Store {
	return &store{DB: db}
}

type store struct {
	DB *pgxpool.Pool
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

func (s *store) UpdateGameReferees(gameRefereeUpdates []GameRefereeUpdate) ([]GameReferee, error) {
	tx, err := s.DB.Begin(context.Background())
	defer tx.Rollback(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

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

	batchResults := tx.SendBatch(context.Background(), bp)

	insertedGameReferees := []GameReferee{}

	for _, _ = range insertedGameReferees {
		gameReferee := GameReferee{}

		err := batchResults.QueryRow().Scan(
			&gameReferee.GameID,
			&gameReferee.RefereeID,
			&gameReferee.Assignment,
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
