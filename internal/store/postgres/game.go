package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/game"
	"github.com/jackc/pgx/v4"
)

func (d DB) GetGameWithID(ctx context.Context, id string) (api.Game, error) {
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

	games := []api.Game{}

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

func (d DB) UpdateGamesOld(ctx context.Context, gameUpdates []game.GameUpdateOld) ([]api.Game, error) {
	if len(gameUpdates) == 0 {
		return nil, nil
	}
	tx, err := d.pgxPool.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update games old with error: %w", err)
	}

	insertGame := `
		INSERT INTO nba.game
			as g(home_team_id, away_team_id, home_team_points, away_team_points, game_status_id, attendance, season_id, season_stage_id, period, period_time_remaining_tenth_seconds, duration_seconds, start_time, end_time, nba_game_id)
		VALUES ((SELECT id FROM nba.Team WHERE nba_team_id = $1), (SELECT id FROM nba.Team WHERE nba_team_id = $2), $3, $4, (SELECT id FROM nba.game_status WHERE name = $5), $6, (SELECT id FROM nba.season WHERE start_year = $7), (SELECT id FROM nba.season_stage WHERE name = $8), $9, $10, $11, $12, $13, $14)
		ON CONFLICT (nba_game_id) DO UPDATE
		SET 
			home_team_id = coalesce(excluded.home_team_id, g.home_team_id),
			away_team_id = coalesce(excluded.away_team_id, g.away_team_id),
			home_team_points = coalesce(excluded.home_team_points, g.home_team_points),
			away_team_points = coalesce(excluded.away_team_points, g.away_team_points),
			game_status_id = coalesce(excluded.game_status_id, g.game_status_id),
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

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return d.GetGamesWithIDs(ctx, insertedGameIDs)
}

func (d DB) UpdateGames(ctx context.Context, gameUpdates []game.GameUpdate) ([]api.Game, error) {
	tx, err := d.pgxPool.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update games with error: %w", err)
	}

	insertGame := `
		INSERT INTO nba.game
			as g(home_team_id, away_team_id, home_team_points, away_team_points, game_status_id, arena_id, attendance, season_id, season_stage_id, sellout, period, period_time_remaining_tenth_seconds, duration_seconds, start_time, end_time, regulation_periods, nba_game_id)
		SELECT
			( SELECT id FROM nba.team WHERE nba_team_id = $1 ) as home_team_id,
			( SELECT id FROM nba.team WHERE nba_team_id = $2 ) as away_team_id,
		    $3,
		    $4,
			( SELECT id FROM nba.game_status WHERE name = $5 ) as game_status_id,
			( SELECT id FROM nba.arena WHERE nba_arena_id = $6 ) as arena_id,
			$7,
			( SELECT nba.season.id FROM nba.season JOIN nba.season_week ON nba.season.id = nba.season_week.season_id WHERE nba.season_week.start_date <= $8 AND nba.season_week.end_date > $8) as season_id,
			( SELECT id FROM nba.season_stage WHERE name = 'regular' ) as season_stage_id,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16
		ON CONFLICT (nba_game_id) DO UPDATE
		SET
			home_team_id = coalesce(excluded.home_team_id, g.home_team_id),
			away_team_id = coalesce(excluded.away_team_id, g.away_team_id),
			home_team_points = coalesce(excluded.home_team_points, g.home_team_points),
			away_team_points = coalesce(excluded.away_team_points, g.away_team_points),
			game_status_id = coalesce(excluded.game_status_id, g.game_status_id),
			arena_id = coalesce(excluded.arena_id, g.arena_id),
			season_id = coalesce(excluded.season_id, g.season_id),
			season_stage_id = coalesce(excluded.season_stage_id, g.season_stage_id),
			attendance = coalesce(excluded.attendance, g.attendance),
			sellout = coalesce(excluded.sellout, g.sellout),
			period = coalesce(excluded.period, g.period),
			period_time_remaining_tenth_seconds = coalesce(excluded.period_time_remaining_tenth_seconds, g.period_time_remaining_tenth_seconds),
			duration_seconds = coalesce(excluded.duration_seconds, g.duration_seconds),
			start_time = g.start_time,
			end_time = coalesce(excluded.end_time, g.end_time),
			regulation_periods = coalesce(excluded.regulation_periods, g.regulation_periods),
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
			gameUpdate.NBAArenaID,
			gameUpdate.Attendance,
			gameUpdate.StartTime,
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

	insertedGameIDs := []string{}

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
