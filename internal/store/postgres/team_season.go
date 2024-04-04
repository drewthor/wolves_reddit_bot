package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/team_season"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (d DB) GetTeamSeasonsWithIDs(ctx context.Context, teamSeasonIDs []string) ([]api.TeamSeason, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.GetTeamSeasonsWithIDs")
	defer span.End()

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
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.UpdateTeamSeasons")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction with error: %w", err)
	}
	defer tx.Rollback(ctx)

	insertQuery := `
		INSERT INTO nba.team_season
			(team_id, league_id, season_id, conference_id, division_id, name, city)
		VALUES ((SELECT id FROM nba.team WHERE nba_team_id = $1), (SELECT id FROM nba.league WHERE nba_league_id = $2), (SELECT id FROM nba.season WHERE start_year = $3), (SELECT id FROM nba.conference WHERE name = $4), (SELECT id FROM nba.division WHERE name = $5), $6, $7)
		ON CONFLICT (team_id, league_id, season_id) DO UPDATE
		SET
			team_id = coalesce(excluded.team_id, nba.team_season.team_id),
			league_id = coalesce(excluded.league_id, nba.team_season.league_id),
			season_id = coalesce(excluded.season_id, nba.team_season.season_id),
			conference_id = coalesce(excluded.conference_id, nba.team_season.conference_id),
			division_id = coalesce(excluded.division_id, nba.team_season.division_id),
			name = coalesce(excluded.name, nba.team_season.name),
			city = coalesce(excluded.city, nba.team_season.city)
		RETURNING nba.team_season.*`

	b := &pgx.Batch{}

	for _, teamSeasonUpdate := range teamSeasonUpdates {
		b.Queue(
			insertQuery,
			teamSeasonUpdate.NBATeamID,
			teamSeasonUpdate.NBALeagueID,
			teamSeasonUpdate.SeasonStartYear,
			teamSeasonUpdate.ConferenceName,
			teamSeasonUpdate.DivisionName,
			teamSeasonUpdate.Name,
			teamSeasonUpdate.City,
		)
	}

	batchResults := tx.SendBatch(ctx, b)

	var updatedTeamSeasons []api.TeamSeason
	for _ = range teamSeasonUpdates {
		var teamSeason api.TeamSeason
		err := batchResults.QueryRow().Scan(
			&teamSeason.ID,
			&teamSeason.TeamID,
			&teamSeason.SeasonID,
			&teamSeason.LeagueID,
			&teamSeason.ConferenceID,
			&teamSeason.DivisionID,
			&teamSeason.CreatedAt,
			&teamSeason.UpdatedAt,
			&teamSeason.City,
			&teamSeason.Name,
		)
		if err != nil {
			batchResults.Close()
			return nil, err
		}

		updatedTeamSeasons = append(updatedTeamSeasons, teamSeason)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return updatedTeamSeasons, nil
}
