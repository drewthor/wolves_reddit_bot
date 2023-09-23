package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/playbyplay"
	"go.opentelemetry.io/otel"
)

func (d DB) UpdatePlayByPlays(ctx context.Context, playByPlayUpdates []playbyplay.PlayByPlayUpdate) ([]api.PlayByPlay, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.UpdatePlayByPlays")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update play by plays with error: %w", err)
	}
	defer tx.Rollback(ctx)

	/*
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

		bp := &pgxutil.Batch{}

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
		}*/

	return nil, nil
}
