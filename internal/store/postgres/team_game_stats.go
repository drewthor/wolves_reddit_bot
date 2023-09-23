package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/internal/team_game_stats"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (d DB) UpdateTeamGameStatsTotals(ctx context.Context, teamGameStatsTotalsUpdates []team_game_stats.TeamGameStatsTotalUpdate) ([]team_game_stats.TeamGameStatsTotal, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.UpdateTeamGameStatsTotals")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update team game stats totals: %w", err)
	}
	defer tx.Rollback(ctx)

	insertTeamGameStatsTotal := `
		INSERT INTO nba.team_game_stats_total
			as tgst(
			        game_id, 
			        team_id, 
			        game_time_played_seconds,
			        total_player_time_played_seconds, 
			        points, 
			        points_against,
			        assists, 
			        personal_turnovers, 
			        team_turnovers, 
			        total_turnovers, 
			        steals, 
			        three_pointers_attempted, 
			        three_pointers_made, 
			        field_goals_attempted, 
			        field_goals_made,
			        effective_adjusted_field_goals,
			        free_throws_attempted,
			        free_throws_made,
			        blocks,
			        times_blocked,
			        personal_offensive_rebounds,
			        personal_defensive_rebounds,
			        total_personal_rebounds,
			        team_rebounds,
			        team_offensive_rebounds,
			        team_defensive_rebounds,
			        total_offensive_rebounds,
			        total_defensive_rebounds,
			        total_rebounds,
			        personal_fouls,
			        offensive_fouls,
			        fouls_drawn,
			        team_fouls,
			        personal_technical_fouls,
			        team_technical_fouls,
			        full_timeouts_remaining,
			        short_timeouts_remaining,
			        total_timeouts_remaining,
			        fast_break_points,
			        fast_break_points_attempted,
			        fast_break_points_made,
			        points_in_paint,
			        points_in_paint_attempted,
			        points_in_paint_made,
			        second_chance_points,
			        second_chance_points_attempted,
			        second_chance_points_made,
			        points_off_turnovers,
			        biggest_lead,
			        biggest_lead_score,
			        biggest_scoring_run,
			        biggest_scoring_run_score,
			        time_leading_tenth_seconds,
			        lead_changes,
			        times_tied,
			        true_shooting_attempts,
			        true_shooting_percentage,
			        bench_points)
		VALUES ((SELECT id FROM nba.game WHERE nba_game_id = $1), (SELECT id FROM nba.team WHERE nba_team_id = $2), $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58)
		ON CONFLICT (game_id, team_id) DO UPDATE
		SET
			game_id = coalesce(excluded.game_id, tgst.game_id),
			team_id = coalesce(excluded.team_id, tgst.team_id),
			game_time_played_seconds = coalesce(excluded.game_time_played_seconds, tgst.game_time_played_seconds),
			total_player_time_played_seconds = coalesce(excluded.total_player_time_played_seconds, tgst.total_player_time_played_seconds),
			points = coalesce(excluded.points, tgst.points),
			points_against = coalesce(excluded.points_against, tgst.points_against),
			assists = coalesce(excluded.assists, tgst.assists),
			personal_turnovers = coalesce(excluded.personal_turnovers, tgst.personal_turnovers),
			team_turnovers = coalesce(excluded.team_turnovers, tgst.team_turnovers),
			total_turnovers = coalesce(excluded.total_turnovers, tgst.total_turnovers),
			steals = coalesce(excluded.steals, tgst.steals),
			three_pointers_attempted = coalesce(excluded.three_pointers_attempted, tgst.three_pointers_attempted),
			three_pointers_made = coalesce(excluded.three_pointers_made, tgst.three_pointers_made),
			field_goals_attempted = coalesce(excluded.field_goals_attempted, tgst.field_goals_attempted),
			field_goals_made = coalesce(excluded.field_goals_made, tgst.field_goals_made),
			effective_adjusted_field_goals = coalesce(excluded.effective_adjusted_field_goals, tgst.effective_adjusted_field_goals),
			free_throws_attempted = coalesce(excluded.free_throws_attempted, tgst.free_throws_attempted),
			free_throws_made = coalesce(excluded.free_throws_made, tgst.free_throws_made),
			blocks = coalesce(excluded.blocks, tgst.blocks),
			times_blocked = coalesce(excluded.times_blocked, tgst.times_blocked),
			personal_offensive_rebounds = coalesce(excluded.personal_offensive_rebounds, tgst.personal_offensive_rebounds),
			personal_defensive_rebounds = coalesce(excluded.personal_defensive_rebounds, tgst.personal_defensive_rebounds),
			total_personal_rebounds = coalesce(excluded.total_personal_rebounds, tgst.total_personal_rebounds),
			team_rebounds = coalesce(excluded.team_rebounds, tgst.team_rebounds),
			team_offensive_rebounds = coalesce(excluded.team_offensive_rebounds, tgst.team_offensive_rebounds),
			team_defensive_rebounds = coalesce(excluded.team_defensive_rebounds, tgst.team_defensive_rebounds),
			total_offensive_rebounds = coalesce(excluded.total_offensive_rebounds, tgst.total_offensive_rebounds),
			total_defensive_rebounds = coalesce(excluded.total_defensive_rebounds, tgst.total_defensive_rebounds),
			total_rebounds = coalesce(excluded.total_rebounds, tgst.total_rebounds),
			personal_fouls = coalesce(excluded.personal_fouls, tgst.personal_fouls),
			offensive_fouls = coalesce(excluded.offensive_fouls, tgst.offensive_fouls),
			fouls_drawn = coalesce(excluded.fouls_drawn, tgst.fouls_drawn),
			team_fouls = coalesce(excluded.team_fouls, tgst.team_fouls),
			personal_technical_fouls = coalesce(excluded.personal_technical_fouls, tgst.personal_technical_fouls),
			team_technical_fouls = coalesce(excluded.team_technical_fouls, tgst.team_technical_fouls),
			full_timeouts_remaining = coalesce(excluded.full_timeouts_remaining, tgst.full_timeouts_remaining),
			short_timeouts_remaining = coalesce(excluded.short_timeouts_remaining, tgst.short_timeouts_remaining),
			total_timeouts_remaining = coalesce(excluded.total_timeouts_remaining, tgst.total_timeouts_remaining),
			fast_break_points = coalesce(excluded.fast_break_points, tgst.fast_break_points),
			fast_break_points_attempted = coalesce(excluded.fast_break_points_attempted, tgst.fast_break_points_attempted),
			fast_break_points_made = coalesce(excluded.fast_break_points_made, tgst.fast_break_points_made),
			points_in_paint = coalesce(excluded.points_in_paint, tgst.points_in_paint),
			points_in_paint_attempted = coalesce(excluded.points_in_paint_attempted, tgst.points_in_paint_attempted),
			points_in_paint_made = coalesce(excluded.points_in_paint_made, tgst.points_in_paint_made),
			second_chance_points = coalesce(excluded.second_chance_points, tgst.second_chance_points),
			second_chance_points_attempted = coalesce(excluded.second_chance_points_attempted, tgst.second_chance_points_attempted),
			second_chance_points_made = coalesce(excluded.second_chance_points_made, tgst.second_chance_points_made),
			points_off_turnovers = coalesce(excluded.points_off_turnovers, tgst.points_off_turnovers),
			biggest_lead = coalesce(excluded.biggest_lead, tgst.biggest_lead),
			biggest_lead_score = coalesce(excluded.biggest_lead_score, tgst.biggest_lead_score),
			biggest_scoring_run = coalesce(excluded.biggest_scoring_run, tgst.biggest_scoring_run),
			biggest_scoring_run_score = coalesce(excluded.biggest_scoring_run_score, tgst.biggest_scoring_run_score),
			time_leading_tenth_seconds = coalesce(excluded.time_leading_tenth_seconds, tgst.time_leading_tenth_seconds),
			lead_changes = coalesce(excluded.lead_changes, tgst.lead_changes),
			times_tied = coalesce(excluded.times_tied, tgst.times_tied),
			true_shooting_attempts = coalesce(excluded.true_shooting_attempts, tgst.true_shooting_attempts),
			true_shooting_percentage = coalesce(excluded.true_shooting_percentage, tgst.true_shooting_percentage),
			bench_points = coalesce(excluded.bench_points, tgst.bench_points)
		RETURNING tgst.*`

	bp := &pgx.Batch{}

	for _, teamGameStatsTotalUpdate := range teamGameStatsTotalsUpdates {
		bp.Queue(
			insertTeamGameStatsTotal,
			teamGameStatsTotalUpdate.NBAGameID,
			teamGameStatsTotalUpdate.NBATeamID,
			teamGameStatsTotalUpdate.GameTimePlayedSeconds,
			teamGameStatsTotalUpdate.TotalPlayerTimePlayedSeconds,
			teamGameStatsTotalUpdate.Points,
			teamGameStatsTotalUpdate.PointsAgainst,
			teamGameStatsTotalUpdate.Assists,
			teamGameStatsTotalUpdate.PersonalTurnovers,
			teamGameStatsTotalUpdate.TeamTurnovers,
			teamGameStatsTotalUpdate.TotalTurnovers,
			teamGameStatsTotalUpdate.Steals,
			teamGameStatsTotalUpdate.ThreePointersAttempted,
			teamGameStatsTotalUpdate.ThreePointersMade,
			teamGameStatsTotalUpdate.FieldGoalsAttempted,
			teamGameStatsTotalUpdate.FieldGoalsMade,
			teamGameStatsTotalUpdate.EffectiveAdjustedFieldGoals,
			teamGameStatsTotalUpdate.FreeThrowsAttempted,
			teamGameStatsTotalUpdate.FreeThrowsMade,
			teamGameStatsTotalUpdate.Blocks,
			teamGameStatsTotalUpdate.TimesBlocked,
			teamGameStatsTotalUpdate.PersonalOffensiveRebounds,
			teamGameStatsTotalUpdate.PersonalDefensiveRebounds,
			teamGameStatsTotalUpdate.TotalPersonalRebounds,
			teamGameStatsTotalUpdate.TeamRebounds,
			teamGameStatsTotalUpdate.TeamOffensiveRebounds,
			teamGameStatsTotalUpdate.TeamDefensiveRebounds,
			teamGameStatsTotalUpdate.TotalOffensiveRebounds,
			teamGameStatsTotalUpdate.TotalDefensiveRebounds,
			teamGameStatsTotalUpdate.TotalRebounds,
			teamGameStatsTotalUpdate.PersonalFouls,
			teamGameStatsTotalUpdate.OffensiveFouls,
			teamGameStatsTotalUpdate.FoulsDrawn,
			teamGameStatsTotalUpdate.TeamFouls,
			teamGameStatsTotalUpdate.PersonalTechnicalFouls,
			teamGameStatsTotalUpdate.TeamTechnicalFouls,
			teamGameStatsTotalUpdate.FullTimeoutsRemaining,
			teamGameStatsTotalUpdate.ShortTimeoutsRemaining,
			teamGameStatsTotalUpdate.TotalTimeoutsRemaining,
			teamGameStatsTotalUpdate.FastBreakPoints,
			teamGameStatsTotalUpdate.FastBreakPointsAttempted,
			teamGameStatsTotalUpdate.FastBreakPointsMade,
			teamGameStatsTotalUpdate.PointsInPaint,
			teamGameStatsTotalUpdate.PointsInPaintAttempted,
			teamGameStatsTotalUpdate.PointsInPaintMade,
			teamGameStatsTotalUpdate.SecondChancePoints,
			teamGameStatsTotalUpdate.SecondChancePointsAttempted,
			teamGameStatsTotalUpdate.SecondChancePointsMade,
			teamGameStatsTotalUpdate.PointsOffTurnovers,
			teamGameStatsTotalUpdate.BiggestLead,
			teamGameStatsTotalUpdate.BiggestLeadScore,
			teamGameStatsTotalUpdate.BiggestScoringRun,
			teamGameStatsTotalUpdate.BiggestScoringRunScore,
			teamGameStatsTotalUpdate.TimeLeadingTenthSeconds,
			teamGameStatsTotalUpdate.LeadChanges,
			teamGameStatsTotalUpdate.TimesTied,
			teamGameStatsTotalUpdate.TrueShootingAttempts,
			teamGameStatsTotalUpdate.TrueShootingPercentage,
			teamGameStatsTotalUpdate.BenchPoints,
		)
	}

	batchResults := tx.SendBatch(ctx, bp)

	insertedTeamGameStatsTotals := []team_game_stats.TeamGameStatsTotal{}

	for _, _ = range teamGameStatsTotalsUpdates {
		t := team_game_stats.TeamGameStatsTotal{}

		err := batchResults.QueryRow().Scan(
			&t.ID,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.GameID,
			&t.TeamID,
			&t.GameTimePlayedSeconds,
			&t.TotalPlayerTimePlayedSeconds,
			&t.Points,
			&t.PointsAgainst,
			&t.Assists,
			&t.PersonalTurnovers,
			&t.TeamTurnovers,
			&t.TotalTurnovers,
			&t.Steals,
			&t.ThreePointersAttempted,
			&t.ThreePointersMade,
			&t.FieldGoalsAttempted,
			&t.FieldGoalsMade,
			&t.EffectiveAdjustedFieldGoals,
			&t.FreeThrowsAttempted,
			&t.FreeThrowsMade,
			&t.Blocks,
			&t.TimesBlocked,
			&t.PersonalOffensiveRebounds,
			&t.PersonalDefensiveRebounds,
			&t.TotalPersonalRebounds,
			&t.TeamRebounds,
			&t.TeamOffensiveRebounds,
			&t.TeamDefensiveRebounds,
			&t.TotalOffensiveRebounds,
			&t.TotalDefensiveRebounds,
			&t.TotalRebounds,
			&t.PersonalFouls,
			&t.OffensiveFouls,
			&t.FoulsDrawn,
			&t.TeamFouls,
			&t.PersonalTechnicalFouls,
			&t.TeamTechnicalFouls,
			&t.FullTimeoutsRemaining,
			&t.ShortTimeoutsRemaining,
			&t.TotalTimeoutsRemaining,
			&t.FastBreakPoints,
			&t.FastBreakPointsAttempted,
			&t.FastBreakPointsMade,
			&t.PointsInPaint,
			&t.PointsInPaintAttempted,
			&t.PointsInPaintMade,
			&t.SecondChancePoints,
			&t.SecondChancePointsAttempted,
			&t.SecondChancePointsMade,
			&t.PointsOffTurnovers,
			&t.BiggestLead,
			&t.BiggestLeadScore,
			&t.BiggestScoringRun,
			&t.BiggestScoringRunScore,
			&t.TimeLeadingTenthSeconds,
			&t.LeadChanges,
			&t.TimesTied,
			&t.TrueShootingAttempts,
			&t.TrueShootingPercentage,
			&t.BenchPoints,
		)

		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedTeamGameStatsTotals = append(insertedTeamGameStatsTotals, t)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return insertedTeamGameStatsTotals, nil
}

func (d DB) UpdateTeamGameStatsTotalsOld(ctx context.Context, teamGameStatsTotalsUpdates []team_game_stats.TeamGameStatsTotalUpdateOld) ([]team_game_stats.TeamGameStatsTotal, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.UpdateTeamGameStatsTotalsOld")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update team game stats totals: %w", err)
	}
	defer tx.Rollback(ctx)

	insertTeamGameStatsTotal := `
		INSERT INTO nba.team_game_stats_total
			as tgst(
			        game_id,
			        team_id,
			        game_time_played_seconds,
			        total_player_time_played_seconds,
			        points,
			        points_against,
			        assists,
			        total_turnovers,
			        steals,
			        three_pointers_attempted,
			        three_pointers_made,
			        field_goals_attempted,
			        field_goals_made,
			        free_throws_attempted,
			        free_throws_made,
			        blocks,
			        times_blocked,
			        total_offensive_rebounds,
			        total_defensive_rebounds,
			        total_rebounds,
			        personal_fouls,
			        team_fouls,
			        total_timeouts_remaining,
			        fast_break_points,
			        points_in_paint,
			        second_chance_points,
			        points_off_turnovers,
			        biggest_lead,
			        lead_changes,
			        times_tied)
			VALUES ((SELECT id FROM nba.game WHERE nba_game_id = $1), (SELECT id FROM nba.team WHERE nba_team_id = $2), $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
		ON CONFLICT (game_id, team_id) DO UPDATE
		SET
			game_id = coalesce(excluded.game_id, tgst.game_id),
			team_id = coalesce(excluded.team_id, tgst.team_id),
			game_time_played_seconds = coalesce(excluded.game_time_played_seconds, tgst.game_time_played_seconds),
			total_player_time_played_seconds = coalesce(excluded.total_player_time_played_seconds, tgst.total_player_time_played_seconds),
			points = coalesce(excluded.points, tgst.points),
			points_against = coalesce(excluded.points_against, tgst.points_against),
			assists = coalesce(excluded.assists, tgst.assists),
			total_turnovers = coalesce(excluded.total_turnovers, tgst.total_turnovers),
			steals = coalesce(excluded.steals, tgst.steals),
			three_pointers_attempted = coalesce(excluded.three_pointers_attempted, tgst.three_pointers_attempted),
			three_pointers_made = coalesce(excluded.three_pointers_made, tgst.three_pointers_made),
			field_goals_attempted = coalesce(excluded.field_goals_attempted, tgst.field_goals_attempted),
			field_goals_made = coalesce(excluded.field_goals_made, tgst.field_goals_made),
			free_throws_attempted = coalesce(excluded.free_throws_attempted, tgst.free_throws_attempted),
			free_throws_made = coalesce(excluded.free_throws_made, tgst.free_throws_made),
			blocks = coalesce(excluded.blocks, tgst.blocks),
			times_blocked = coalesce(excluded.times_blocked, tgst.times_blocked),
			total_offensive_rebounds = coalesce(excluded.total_offensive_rebounds, tgst.total_offensive_rebounds),
			total_defensive_rebounds = coalesce(excluded.total_defensive_rebounds, tgst.total_defensive_rebounds),
			total_rebounds = coalesce(excluded.total_rebounds, tgst.total_rebounds),
			personal_fouls = coalesce(excluded.personal_fouls, tgst.personal_fouls),
			team_fouls = coalesce(excluded.team_fouls, tgst.team_fouls),
			total_timeouts_remaining = coalesce(excluded.total_timeouts_remaining, tgst.total_timeouts_remaining),
			fast_break_points = coalesce(excluded.fast_break_points, tgst.fast_break_points),
			points_in_paint = coalesce(excluded.points_in_paint, tgst.points_in_paint),
			second_chance_points = coalesce(excluded.second_chance_points, tgst.second_chance_points),
			points_off_turnovers = coalesce(excluded.points_off_turnovers, tgst.points_off_turnovers),
			biggest_lead = coalesce(excluded.biggest_lead, tgst.biggest_lead),
			lead_changes = coalesce(excluded.lead_changes, tgst.lead_changes),
			times_tied = coalesce(excluded.times_tied, tgst.times_tied)
		RETURNING tgst.*`

	bp := &pgx.Batch{}

	for _, teamGameStatsTotalUpdate := range teamGameStatsTotalsUpdates {
		bp.Queue(
			insertTeamGameStatsTotal,
			teamGameStatsTotalUpdate.NBAGameID,
			teamGameStatsTotalUpdate.NBATeamID,
			teamGameStatsTotalUpdate.GameTimePlayedSeconds,
			teamGameStatsTotalUpdate.TotalPlayerTimePlayedSeconds,
			teamGameStatsTotalUpdate.Points,
			teamGameStatsTotalUpdate.PointsAgainst,
			teamGameStatsTotalUpdate.Assists,
			teamGameStatsTotalUpdate.TotalTurnovers,
			teamGameStatsTotalUpdate.Steals,
			teamGameStatsTotalUpdate.ThreePointersAttempted,
			teamGameStatsTotalUpdate.ThreePointersMade,
			teamGameStatsTotalUpdate.FieldGoalsAttempted,
			teamGameStatsTotalUpdate.FieldGoalsMade,
			teamGameStatsTotalUpdate.FreeThrowsAttempted,
			teamGameStatsTotalUpdate.FreeThrowsMade,
			teamGameStatsTotalUpdate.Blocks,
			teamGameStatsTotalUpdate.TimesBlocked,
			teamGameStatsTotalUpdate.TotalOffensiveRebounds,
			teamGameStatsTotalUpdate.TotalDefensiveRebounds,
			teamGameStatsTotalUpdate.TotalRebounds,
			teamGameStatsTotalUpdate.PersonalFouls,
			teamGameStatsTotalUpdate.TeamFouls,
			teamGameStatsTotalUpdate.TotalTimeoutsRemaining,
			teamGameStatsTotalUpdate.FastBreakPoints,
			teamGameStatsTotalUpdate.PointsInPaint,
			teamGameStatsTotalUpdate.SecondChancePoints,
			teamGameStatsTotalUpdate.PointsOffTurnovers,
			teamGameStatsTotalUpdate.BiggestLead,
			teamGameStatsTotalUpdate.LeadChanges,
			teamGameStatsTotalUpdate.TimesTied,
		)
	}

	batchResults := tx.SendBatch(ctx, bp)

	insertedTeamGameStatsTotals := []team_game_stats.TeamGameStatsTotal{}

	for _, _ = range teamGameStatsTotalsUpdates {
		t := team_game_stats.TeamGameStatsTotal{}

		err := batchResults.QueryRow().Scan(
			&t.ID,
			&t.GameID,
			&t.TeamID,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.GameTimePlayedSeconds,
			&t.TotalPlayerTimePlayedSeconds,
			&t.Points,
			&t.PointsAgainst,
			&t.Assists,
			&t.PersonalTurnovers,
			&t.TeamTurnovers,
			&t.TotalTurnovers,
			&t.Steals,
			&t.ThreePointersAttempted,
			&t.ThreePointersMade,
			&t.FieldGoalsAttempted,
			&t.FieldGoalsMade,
			&t.EffectiveAdjustedFieldGoals,
			&t.FreeThrowsAttempted,
			&t.FreeThrowsMade,
			&t.Blocks,
			&t.TimesBlocked,
			&t.PersonalOffensiveRebounds,
			&t.PersonalDefensiveRebounds,
			&t.TotalPersonalRebounds,
			&t.TeamRebounds,
			&t.TeamOffensiveRebounds,
			&t.TeamDefensiveRebounds,
			&t.TotalOffensiveRebounds,
			&t.TotalDefensiveRebounds,
			&t.TotalRebounds,
			&t.PersonalFouls,
			&t.OffensiveFouls,
			&t.FoulsDrawn,
			&t.TeamFouls,
			&t.PersonalTechnicalFouls,
			&t.TeamTechnicalFouls,
			&t.FullTimeoutsRemaining,
			&t.ShortTimeoutsRemaining,
			&t.TotalTimeoutsRemaining,
			&t.FastBreakPoints,
			&t.FastBreakPointsAttempted,
			&t.FastBreakPointsMade,
			&t.PointsInPaint,
			&t.PointsInPaintAttempted,
			&t.PointsInPaintMade,
			&t.SecondChancePoints,
			&t.SecondChancePointsAttempted,
			&t.SecondChancePointsMade,
			&t.PointsOffTurnovers,
			&t.BiggestLead,
			&t.BiggestLeadScore,
			&t.BiggestScoringRun,
			&t.BiggestScoringRunScore,
			&t.TimeLeadingTenthSeconds,
			&t.LeadChanges,
			&t.TimesTied,
			&t.TrueShootingAttempts,
			&t.TrueShootingPercentage,
			&t.BenchPoints,
		)

		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedTeamGameStatsTotals = append(insertedTeamGameStatsTotals, t)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return insertedTeamGameStatsTotals, nil
}
