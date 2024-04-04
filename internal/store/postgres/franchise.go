package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/franchise"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (d DB) UpdateFranchises(ctx context.Context, franchiseUpdates []franchise.FranchiseUpdate) ([]api.Franchise, error) {
	ctx, span := otel.Tracer("team").Start(ctx, "postgres.DB.UpdateFranchises")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction to update franchises: %w", err)
	}
	defer tx.Rollback(ctx)

	insertGame := `
		INSERT INTO nba.franchise
			(name, nickname, city, state, country, league_id, nba_team_id, start_year, end_year, years, games, wins, losses, playoff_appearances, division_titles, conference_titles, league_titles, active)
		VALUES ($1, $2, $3, $4, $5, (SELECT id FROM nba.League WHERE nba_league_id = $6), $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (nba_team_id) DO UPDATE
		SET 
			name = coalesce(excluded.name, nba.franchise.name),
			nickname = coalesce(excluded.nickname, nba.franchise.nickname),
			city = coalesce(excluded.city, nba.franchise.city),
			state = coalesce(excluded.state, nba.franchise.state),
			country = coalesce(excluded.country, nba.franchise.country),
			league_id = coalesce(excluded.league_id, nba.franchise.league_id),
			nba_team_id = coalesce(excluded.nba_team_id, nba.franchise.nba_team_id),
			start_year = coalesce(excluded.start_year, nba.franchise.start_year),
			end_year = coalesce(excluded.end_year, nba.franchise.end_year),
			years = coalesce(excluded.years, nba.franchise.years),
			games = coalesce(excluded.games, nba.franchise.games),
			wins = coalesce(excluded.wins, nba.franchise.wins),
			losses = coalesce(excluded.losses, nba.franchise.losses),
			playoff_appearances = coalesce(excluded.playoff_appearances, nba.franchise.playoff_appearances),
			division_titles = coalesce(excluded.division_titles, nba.franchise.division_titles),
			conference_titles = coalesce(excluded.conference_titles, nba.franchise.conference_titles),
			league_titles = coalesce(excluded.league_titles, nba.franchise.league_titles),
			active = coalesce(excluded.active, nba.franchise.active)
		RETURNING nba.franchise.*`

	bp := &pgx.Batch{}

	for _, franchiseUpdate := range franchiseUpdates {
		bp.Queue(insertGame,
			franchiseUpdate.Name,
			franchiseUpdate.Nickname,
			franchiseUpdate.City,
			franchiseUpdate.State,
			franchiseUpdate.Country,
			franchiseUpdate.NBALeagueID,
			franchiseUpdate.NBATeamID,
			franchiseUpdate.StartYear,
			franchiseUpdate.EndYear,
			franchiseUpdate.Years,
			franchiseUpdate.Games,
			franchiseUpdate.Wins,
			franchiseUpdate.Losses,
			franchiseUpdate.PlayoffAppearances,
			franchiseUpdate.DivisionTitles,
			franchiseUpdate.ConferenceTitles,
			franchiseUpdate.LeagueTitles,
			franchiseUpdate.Active)
	}

	batchResults := tx.SendBatch(ctx, bp)

	var franchises []api.Franchise

	for _, _ = range franchiseUpdates {
		var fr api.Franchise
		err := batchResults.QueryRow().Scan(
			&fr.ID,
			&fr.CreatedAt,
			&fr.UpdatedAt,
			&fr.Name,
			&fr.Nickname,
			&fr.City,
			&fr.State,
			&fr.Country,
			&fr.LeagueID,
			&fr.NBATeamID,
			&fr.StartYear,
			&fr.EndYear,
			&fr.Years,
			&fr.Games,
			&fr.Wins,
			&fr.Losses,
			&fr.PlayoffAppearances,
			&fr.DivisionTitles,
			&fr.ConferenceTitles,
			&fr.LeagueTitles,
			&fr.Active,
		)
		if err != nil {
			batchResults.Close()
			return nil, err
		}

		franchises = append(franchises, fr)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return franchises, nil
}
