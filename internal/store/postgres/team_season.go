package postgres

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/team_season"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

func (d DB) GetTeamSeasonsWithIDs(ctx context.Context, teamSeasonIDs []string) ([]api.TeamSeason, error) {
	query := `
		SELECT ts.id, ts.team_id, ts.league_id, ts.season_id, ts.conference_id, ts.division_id, ts.created_at, ts.updated_at
		FROM nba.team_season ts
		WHERE id = ANY($1)`

	rows, err := d.pgxPool.Query(ctx, query, teamSeasonIDs)
	if err != nil {
		return nil, err
	}

	teamSeasons := []api.TeamSeason{}

	for rows.Next() {
		teamSeason := api.TeamSeason{}
		err = rows.Scan(
			&teamSeason.ID,
			&teamSeason.TeamID,
			&teamSeason.LeagueID,
			&teamSeason.SeasonID,
			&teamSeason.ConferenceID,
			&teamSeason.DivisionID,
			&teamSeason.CreatedAt,
			&teamSeason.UpdatedAt)
		if err != nil {
			return nil, err
		}
		teamSeasons = append(teamSeasons, teamSeason)
	}

	return teamSeasons, nil
}

func (d DB) UpdateTeamSeasons(ctx context.Context, teamSeasonUpdates []team_season.TeamSeasonUpdate) ([]api.TeamSeason, error) {
	tx, err := d.pgxPool.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

	insertQuery := `
		INSERT INTO nba.team_season
			AS ts(team_id, league_id, season_id, conference_id, division_id)
		VALUES ($1, (SELECT id FROM nba.league WHERE name = $2), (SELECT id FROM nba.season WHERE start_year = $3), (SELECT id FROM nba.conference WHERE name = $4), (SELECT id FROM nba.division WHERE name = $5))
		ON CONFLICT (team_id, league_id, season_id) DO UPDATE
		SET
			team_id = coalesce(excluded.team_id, ts.team_id),
			league_id = coalesce(excluded.league_id, ts.league_id),
			season_id = coalesce(excluded.season_id, ts.season_id),
			conference_id = coalesce(excluded.conference_id, ts.conference_id),
			division_id = coalesce(excluded.division_id, ts.division_id)
		RETURNING ts.id`

	b := &pgx.Batch{}

	for _, teamSeasonUpdate := range teamSeasonUpdates {
		b.Queue(
			insertQuery,
			teamSeasonUpdate.TeamID,
			teamSeasonUpdate.LeagueName,
			teamSeasonUpdate.SeasonStartYear,
			teamSeasonUpdate.ConferenceName,
			teamSeasonUpdate.DivisionName)
	}

	batchResults := tx.SendBatch(ctx, b)

	insertedTeamSeasonIDs := []string{}

	for _ = range teamSeasonUpdates {
		id := ""
		err := batchResults.QueryRow().Scan(&id)
		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedTeamSeasonIDs = append(insertedTeamSeasonIDs, id)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return d.GetTeamSeasonsWithIDs(ctx, insertedTeamSeasonIDs)
}
