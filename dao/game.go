package dao

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type GameDAO struct {
	DB *pgxpool.Pool
}

func (gd *GameDAO) GetByIDs(ids []string) ([]api.Game, error) {
	query := `
		SELECT id, home_team_id, away_team_id, home_team_points, away_team_points, game_status.name, arena_id, attendance, season.name, season_stage.name, period, period_time_remaining_tenth_seconds, duration_seconds, start_time, end_time, nba_game_id, created_at, updated_at
		FROM nba.game g, 
		LATERAL (
		        SELECT name
				FROM nba.game_status gs
				WHERE gs.id = g.game_status_id
		) game_status,
		LATERAL (
		        SELECT CONCAT(s.start_year, '-', s.end_year) as name
				FROM nba.season s
				WHERE s.id = g.season_id
		) season,
		LATERAL (
		        SELECT name
				FROM nba.season_stage ss
				WHERE ss.id = g.season_stage_id
        ) season_stage
		WHERE id = ANY($1)`

	rows, err := gd.DB.Query(context.Background(), query, ids)
	if err != nil {
		return nil, err
	}

	games := []api.Game{}

	for rows.Next() {
		game := api.Game{}
		err = rows.Scan(
			&game.ID,
			&game.HomeTeamID,
			&game.AwayTeamID,
			&game.HomeTeamPoints,
			&game.AwayTeamPoints,
			&game.Status,
			&game.ArenaID,
			&game.Attendance,
			&game.Season,
			&game.SeasonStage,
			&game.Period,
			&game.PeriodTimeRemaining,
			&game.Duration,
			&game.StartTime,
			&game.EndTime,
			&game.NBAGameID,
			&game.CreatedAt,
			&game.UpdatedAt)
		if err != nil {
			return nil, err
		}

		games = append(games, game)
	}

	return games, nil
}

type GameUpdate struct {
	NBAHomeTeamID                   string
	NBAAwayTeamID                   string
	HomeTeamPoints                  sql.NullInt64
	AwayTeamPoints                  sql.NullInt64
	GameStatusName                  string
	ArenaName                       string
	Attendance                      int
	SeasonStartYear                 string
	SeasonStageName                 string
	Period                          int
	PeriodTimeRemainingTenthSeconds int
	DurationSeconds                 *int
	StartTime                       time.Time
	EndTime                         *time.Time
	NBAGameID                       string
}

func (gd *GameDAO) UpdateGames(gameUpdates []GameUpdate) ([]api.Game, error) {
	tx, err := gd.DB.Begin(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

	insertGame := `
		INSERT INTO nba.game
			as g(home_team_id, away_team_id, home_team_points, away_team_points, game_status_id, arena_id, attendance, season_id, season_stage_id, period, period_time_remaining_tenth_seconds, duration_seconds, start_time, end_time, nba_game_id)
		VALUES ((SELECT id FROM nba.Team WHERE nba_team_id = $1), (SELECT id FROM nba.Team WHERE nba_team_id = $2), $3, $4, (SELECT id FROM nba.game_status WHERE name = $5), (SELECT id FROM nba.arena WHERE name = $6), $7, (SELECT id FROM nba.season WHERE start_year = $8), (SELECT id FROM nba.season_stage WHERE name = $9), $10, $11, $12, $13, $14, $15)
		ON CONFLICT (nba_game_id) DO UPDATE
		SET 
			home_team_id = coalesce(excluded.home_team_id, g.home_team_id),
			away_team_id = coalesce(excluded.away_team_id, g.away_team_id),
			home_team_points = coalesce(excluded.home_team_points, g.home_team_points),
			away_team_points = coalesce(excluded.away_team_points, g.away_team_points),
			game_status_id = coalesce(excluded.game_status_id, g.game_status_id),
			arena_id = coalesce(excluded.arena_id, g.arena_id),
			attendance = coalesce(excluded.attendance, g.attendance),
			season_id = coalesce(excluded.season_id, g.season_id),
			season_stage_id = coalesce(excluded.season_stage_id, g.season_stage_id),
			period = coalesce(excluded.period, g.period),
			period_time_remaining_tenth_seconds = coalesce(excluded.period_time_remaining_tenth_seconds, g.period_time_remaining_tenth_seconds),
			duration_seconds = coalesce(excluded.duration_seconds, g.duration_seconds),
			start_time = coalesce(excluded.start_time, g.start_time),
			end_time = coalesce(excluded.end_time, g.end_time),
			nba_game_id = coalesce(excluded.nba_game_id, g.nba_game_id)
		RETURNING g.id`

	bp := &pgx.Batch{}

	for _, gameUpdate := range gameUpdates {
		bp.Queue(insertGame,
			gameUpdate.NBAHomeTeamID,
			gameUpdate.NBAAwayTeamID,
			gameUpdate.HomeTeamPoints,
			gameUpdate.AwayTeamPoints,
			gameUpdate.GameStatusName,
			gameUpdate.ArenaName,
			gameUpdate.Attendance,
			gameUpdate.SeasonStartYear,
			gameUpdate.SeasonStageName,
			gameUpdate.Period,
			gameUpdate.PeriodTimeRemainingTenthSeconds,
			gameUpdate.DurationSeconds,
			gameUpdate.StartTime,
			gameUpdate.EndTime,
			gameUpdate.NBAGameID)
	}

	batchResults := tx.SendBatch(context.Background(), bp)

	insertedGameIDs := []string{}

	for _, _ = range gameUpdates {
		id := ""
		err := batchResults.QueryRow().Scan(&id)
		if err != nil {
			return nil, err
		}

		insertedGameIDs = append(insertedGameIDs, id)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return gd.GetByIDs(insertedGameIDs)
}
