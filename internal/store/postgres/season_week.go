package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/internal/season"
	"github.com/jackc/pgx/v4"
)

func (d DB) UpdateSeasonWeeks(ctx context.Context, seasonWeekUpdates []season.SeasonWeekUpdate) ([]season.SeasonWeek, error) {
	tx, err := d.pgxPool.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update season weeks with error: %w", err)
	}

	insertGame := `
		INSERT INTO nba.season_week
			as s(season_id, start_date, end_date)
		SELECT nba.season.id, $2, $3
		    FROM nba.season
		    WHERE nba.season.start_year = $1
		ON CONFLICT (season_id, start_date) DO UPDATE
		SET 
			season_id = coalesce(excluded.season_id, s.season_id),
			start_date = coalesce(excluded.start_date, s.start_date),
			end_date = coalesce(excluded.end_date, s.end_date)
		RETURNING s.*`

	bp := &pgx.Batch{}

	for _, seasonWeekUpdate := range seasonWeekUpdates {
		bp.Queue(insertGame,
			seasonWeekUpdate.SeasonStartYear,
			seasonWeekUpdate.StartDate,
			seasonWeekUpdate.EndDate)
	}

	batchResults := tx.SendBatch(ctx, bp)

	insertedSeasonWeeks := []season.SeasonWeek{}

	for _, _ = range seasonWeekUpdates {
		var seasonWeek season.SeasonWeek
		err := batchResults.QueryRow().Scan(&seasonWeek.ID,
			&seasonWeek.CreatedAt,
			&seasonWeek.UpdatedAt,
			&seasonWeek.SeasonID,
			&seasonWeek.StartDate,
			&seasonWeek.EndDate)
		if err != nil {
			batchResults.Close()
			return nil, fmt.Errorf("error scanning batch results updating season weeks: %w", err)
		}

		insertedSeasonWeeks = append(insertedSeasonWeeks, seasonWeek)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, fmt.Errorf("error when closing batch results for updating season weeks: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("error when committing tx when updating season weeks: %w", err)
	}

	return insertedSeasonWeeks, nil
}
