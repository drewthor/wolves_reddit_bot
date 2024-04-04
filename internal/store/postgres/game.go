package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/game"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (d DB) List(ctx context.Context) ([]api.Game, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.List")
	defer span.End()

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
        ) season_stage`

	rows, err := d.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var games []api.Game
	for rows.Next() {
		g := api.Game{}
		err = rows.Scan(
			&g.ID,
			&g.HomeTeamID,
			&g.AwayTeamID,
			&g.HomeTeamPoints,
			&g.AwayTeamPoints,
			&g.Status,
			&g.ArenaID,
			&g.Attendance,
			&g.Season,
			&g.SeasonStage,
			&g.Period,
			&g.PeriodTimeRemaining,
			&g.Duration,
			&g.StartTime,
			&g.EndTime,
			&g.NBAGameID,
			&g.CreatedAt,
			&g.UpdatedAt)
		if err != nil {
			return nil, err
		}

		games = append(games, g)
	}

	return games, nil

}

func (d DB) GetGameWithID(ctx context.Context, id string) (api.Game, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.GetGameWithID")
	defer span.End()

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
		WHERE id = $1`

	g := api.Game{}

	err := d.pgxPool.QueryRow(ctx, query, id).Scan(
		&g.ID,
		&g.HomeTeamID,
		&g.AwayTeamID,
		&g.HomeTeamPoints,
		&g.AwayTeamPoints,
		&g.Status,
		&g.ArenaID,
		&g.Attendance,
		&g.Season,
		&g.SeasonStage,
		&g.Period,
		&g.PeriodTimeRemaining,
		&g.Duration,
		&g.StartTime,
		&g.EndTime,
		&g.NBAGameID,
		&g.CreatedAt,
		&g.UpdatedAt)
	if err != nil {
		return api.Game{}, err
	}

	return g, nil
}

func (d DB) GetGamesWithIDs(ctx context.Context, ids []string) ([]api.Game, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.GetGamesWithIDs")
	defer span.End()

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

	rows, err := d.pgxPool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	var games []api.Game
	for rows.Next() {
		g := api.Game{}
		err = rows.Scan(
			&g.ID,
			&g.HomeTeamID,
			&g.AwayTeamID,
			&g.HomeTeamPoints,
			&g.AwayTeamPoints,
			&g.Status,
			&g.ArenaID,
			&g.Attendance,
			&g.Season,
			&g.SeasonStage,
			&g.Period,
			&g.PeriodTimeRemaining,
			&g.Duration,
			&g.StartTime,
			&g.EndTime,
			&g.NBAGameID,
			&g.CreatedAt,
			&g.UpdatedAt)
		if err != nil {
			return nil, err
		}

		games = append(games, g)
	}

	return games, nil
}

func (d DB) GetGameWithNBAID(ctx context.Context, nbaID string) (api.Game, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.GetGameWithNBAID")
	defer span.End()

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
		WHERE nba_game_id = $1`

	g := api.Game{}

	err := d.pgxPool.QueryRow(ctx, query, nbaID).Scan(
		&g.ID,
		&g.HomeTeamID,
		&g.AwayTeamID,
		&g.HomeTeamPoints,
		&g.AwayTeamPoints,
		&g.Status,
		&g.ArenaID,
		&g.Attendance,
		&g.Season,
		&g.SeasonStage,
		&g.Period,
		&g.PeriodTimeRemaining,
		&g.Duration,
		&g.StartTime,
		&g.EndTime,
		&g.NBAGameID,
		&g.CreatedAt,
		&g.UpdatedAt)
	if err != nil {
		return api.Game{}, err
	}

	return g, nil
}

func (d DB) UpdateGamesSummary(ctx context.Context, gameUpdates []game.GameSummaryUpdate) ([]api.Game, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.UpdateGamesSummary")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update games old with error: %w", err)
	}
	defer tx.Rollback(ctx)

	insertGame := `
		INSERT INTO nba.game
			(home_team_id, away_team_id, home_team_points, away_team_points, game_status_id, attendance, season_id, season_stage_id, period, period_time_remaining_tenth_seconds, duration_seconds, start_time, end_time, nba_game_id)
		VALUES ((SELECT id FROM nba.Team WHERE nba_team_id = $1), (SELECT id FROM nba.Team WHERE nba_team_id = $2), $3, $4, (SELECT id FROM nba.game_status WHERE name = $5), $6, (SELECT id FROM nba.season WHERE start_year = $7), (SELECT id FROM nba.season_stage WHERE name = $8), $9, $10, $11, $12, $13, $14)
		ON CONFLICT (nba_game_id) DO UPDATE
		SET 
			home_team_id = coalesce(excluded.home_team_id, nba.game.home_team_id),
			away_team_id = coalesce(excluded.away_team_id, nba.game.away_team_id),
			home_team_points = coalesce(excluded.home_team_points, nba.game.home_team_points),
			away_team_points = coalesce(excluded.away_team_points, nba.game.away_team_points),
			game_status_id = coalesce(excluded.game_status_id, nba.game.game_status_id),
			attendance = coalesce(excluded.attendance, nba.game.attendance),
			season_id = coalesce(excluded.season_id, nba.game.season_id),
			period = coalesce(excluded.period, nba.game.period),
			period_time_remaining_tenth_seconds = coalesce(excluded.period_time_remaining_tenth_seconds, nba.game.period_time_remaining_tenth_seconds),
			duration_seconds = coalesce(excluded.duration_seconds, nba.game.duration_seconds),
			start_time = coalesce(excluded.start_time, nba.game.start_time),
			end_time = coalesce(excluded.end_time, nba.game.end_time),
			nba_game_id = coalesce(excluded.nba_game_id, nba.game.nba_game_id)
		RETURNING nba.game.id`

	bp := &pgx.Batch{}

	for _, gameUpdate := range gameUpdates {
		bp.Queue(insertGame,
			gameUpdate.NBAHomeTeamID,
			gameUpdate.NBAAwayTeamID,
			gameUpdate.HomeTeamPoints,
			gameUpdate.AwayTeamPoints,
			gameUpdate.GameStatusName,
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

	batchResults := tx.SendBatch(ctx, bp)

	var insertedGameIDs []string

	for _, _ = range gameUpdates {
		id := ""
		err := batchResults.QueryRow().Scan(&id)
		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedGameIDs = append(insertedGameIDs, id)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return d.GetGamesWithIDs(ctx, insertedGameIDs)
}

func (d DB) UpdateGames(ctx context.Context, gameUpdates []game.GameUpdate) ([]api.Game, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.UpdateGames")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update games with error: %w", err)
	}
	defer tx.Rollback(ctx)

	insertGame := `
		INSERT INTO nba.game
			(home_team_id, away_team_id, home_team_points, away_team_points, game_status_id, arena_id, attendance, season_id, season_stage_id, sellout, period, period_time_remaining_tenth_seconds, duration_seconds, start_time, end_time, regulation_periods, nba_game_id)
		VALUES (
			(SELECT id FROM nba.team WHERE nba_team_id = $1),
			(SELECT id FROM nba.team WHERE nba_team_id = $2),
		    $3,
		    $4,
			(SELECT id FROM nba.game_status WHERE name = $5),
			(SELECT id FROM nba.arena WHERE nba_arena_id = $6),
			$7,
			(SELECT nba.season.id FROM nba.season WHERE nba.season.start_year = $8),
			(SELECT id FROM nba.season_stage WHERE name = 'regular'),
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16
		)
		ON CONFLICT (nba_game_id) DO UPDATE
		SET
			home_team_id = coalesce(excluded.home_team_id, nba.game.home_team_id),
			away_team_id = coalesce(excluded.away_team_id, nba.game.away_team_id),
			home_team_points = coalesce(excluded.home_team_points, nba.game.home_team_points),
			away_team_points = coalesce(excluded.away_team_points, nba.game.away_team_points),
			game_status_id = coalesce(excluded.game_status_id, nba.game.game_status_id),
			arena_id = coalesce(excluded.arena_id, nba.game.arena_id),
			season_id = coalesce(excluded.season_id, nba.game.season_id),
			season_stage_id = coalesce(excluded.season_stage_id, nba.game.season_stage_id),
			attendance = coalesce(excluded.attendance, nba.game.attendance),
			sellout = coalesce(excluded.sellout, nba.game.sellout),
			period = coalesce(excluded.period, nba.game.period),
			period_time_remaining_tenth_seconds = coalesce(excluded.period_time_remaining_tenth_seconds, nba.game.period_time_remaining_tenth_seconds),
			duration_seconds = coalesce(excluded.duration_seconds, nba.game.duration_seconds),
			start_time = excluded.start_time,
			end_time = coalesce(excluded.end_time, nba.game.end_time),
			regulation_periods = coalesce(excluded.regulation_periods, nba.game.regulation_periods),
			nba_game_id = coalesce(excluded.nba_game_id, nba.game.nba_game_id)
		RETURNING nba.game.id`

	bp := &pgx.Batch{}

	for _, gameUpdate := range gameUpdates {
		bp.Queue(insertGame,
			gameUpdate.NBAHomeTeamID,
			gameUpdate.NBAAwayTeamID,
			gameUpdate.HomeTeamPoints,
			gameUpdate.AwayTeamPoints,
			gameUpdate.GameStatusName,
			gameUpdate.NBAArenaID,
			gameUpdate.Attendance,
			gameUpdate.SeasonStartYear,
			gameUpdate.Sellout,
			gameUpdate.Period,
			gameUpdate.PeriodTimeRemainingTenthSeconds,
			gameUpdate.DurationSeconds,
			gameUpdate.StartTime,
			gameUpdate.EndTime,
			gameUpdate.RegulationPeriods,
			gameUpdate.NBAGameID)
	}

	batchResults := tx.SendBatch(ctx, bp)

	var insertedGameIDs []string

	for _, _ = range gameUpdates {
		id := ""
		err := batchResults.QueryRow().Scan(&id)
		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedGameIDs = append(insertedGameIDs, id)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return d.GetGamesWithIDs(ctx, insertedGameIDs)
}

func (d DB) UpdateScheduledGames(ctx context.Context, gameUpdates []game.GameScheduledUpdate) ([]api.Game, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.UpdateGames")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update games with error: %w", err)
	}
	defer tx.Rollback(ctx)

	insertGame := `
		INSERT INTO nba.game
			(home_team_id, away_team_id, home_team_points, away_team_points, game_status_id, arena_id, season_id, season_stage_id, start_time, nba_game_id)
		VALUES (
			(SELECT id FROM nba.team WHERE nba_team_id = $1),
			(SELECT id FROM nba.team WHERE nba_team_id = $2),
		    $3,
		    $4,
			(SELECT id FROM nba.game_status WHERE name = $5),
			(SELECT id FROM nba.arena WHERE name = $6),
			(SELECT nba.season.id FROM nba.season WHERE nba.season.start_year = $7),
			(SELECT id FROM nba.season_stage WHERE name = 'regular'),
			$8,
			$9
		)
		ON CONFLICT (nba_game_id) DO UPDATE
		SET
			home_team_id = coalesce(excluded.home_team_id, nba.game.home_team_id),
			away_team_id = coalesce(excluded.away_team_id, nba.game.away_team_id),
			home_team_points = coalesce(excluded.home_team_points, nba.game.home_team_points),
			away_team_points = coalesce(excluded.away_team_points, nba.game.away_team_points),
			game_status_id = coalesce(excluded.game_status_id, nba.game.game_status_id),
			arena_id = coalesce(excluded.arena_id, nba.game.arena_id),
			season_id = coalesce(excluded.season_id, nba.game.season_id),
			season_stage_id = coalesce(excluded.season_stage_id, nba.game.season_stage_id),
			start_time = excluded.start_time,
			nba_game_id = coalesce(excluded.nba_game_id, nba.game.nba_game_id)
		RETURNING nba.game.id`

	bp := &pgx.Batch{}

	for _, gameUpdate := range gameUpdates {
		bp.Queue(insertGame,
			gameUpdate.NBAHomeTeamID,
			gameUpdate.NBAAwayTeamID,
			gameUpdate.HomeTeamPoints,
			gameUpdate.AwayTeamPoints,
			gameUpdate.GameStatusName,
			gameUpdate.NBAArenaName,
			gameUpdate.SeasonStartYear,
			gameUpdate.StartTime,
			gameUpdate.NBAGameID)
	}

	batchResults := tx.SendBatch(ctx, bp)

	var insertedGameIDs []string

	for _, _ = range gameUpdates {
		id := ""
		err := batchResults.QueryRow().Scan(&id)
		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedGameIDs = append(insertedGameIDs, id)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return d.GetGamesWithIDs(ctx, insertedGameIDs)
}
