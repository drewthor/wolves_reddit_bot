package postgres

import (
	"context"
	"log/slog"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/team"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (d DB) GetTeamWithID(ctx context.Context, teamID string) (api.Team, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.GetTeamWithID")
	defer span.End()

	query := `
		SELECT id, t.name, nickname, city, city_alternate, state, country, franchise_id, nba_url_name, nba_short_name, nba_team_id, created_at, updated_at
		FROM nba.team t 
		WHERE id = $1`

	// LATERAL (
	// 	SELECT l.name
	// 	FROM nba.league l
	// 	WHERE l.id = t.league_id
	// ) league,
	// LATERAL (
	//         SELECT CONCAT(s.start_year, '-', s.end_year)
	// 		FROM nba.season s
	// 		WHERE s.id = t.season_id
	// ) season,
	// LATERAL (
	//         SELECT c.name
	// 		FROM nba.conference c
	// 		WHERE c.id = t.season_id
	// ) conference,
	// LATERAL (
	//         SELECT d.name
	// 		FROM nba.division d
	// 		WHERE d.id = t.season_id
	// ) division
	// WHERE id = $1`

	team := api.Team{}

	err := d.pgxPool.QueryRow(ctx, query, teamID).Scan(
		&team.ID,
		&team.Name,
		&team.Nickname,
		&team.City,
		&team.AlternateCity,
		&team.State,
		&team.Country,
		&team.FranchiseID,
		&team.NBAURLName,
		&team.NBAShortName,
		&team.NBATeamID,
		&team.CreatedAt,
		&team.UpdatedAt)

	if err != nil {
		return api.Team{}, err
	}

	return team, nil
}

func (d DB) GetTeamsWithIDs(ctx context.Context, ids []uuid.UUID) ([]api.Team, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.GetTeamsWithIDs")
	defer span.End()

	query := `
		SELECT id, t.name, nickname, city, city_alternate, state, country, franchise_id, nba_url_name, nba_short_name, nba_team_id, created_at, updated_at
		FROM nba.team t
		WHERE id = ANY($1)`

	rows, err := d.pgxPool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	teams := []api.Team{}

	for rows.Next() {
		var team api.Team
		err = rows.Scan(
			&team.ID,
			&team.Name,
			&team.Nickname,
			&team.City,
			&team.AlternateCity,
			&team.State,
			&team.Country,
			&team.FranchiseID,
			&team.NBAURLName,
			&team.NBAShortName,
			&team.NBATeamID,
			&team.CreatedAt,
			&team.UpdatedAt)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (d DB) GetTeamsWithNBAIDs(ctx context.Context, ids []int) ([]api.Team, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.GetTeamsWithNBAIDs")
	defer span.End()

	query := `
		SELECT id, t.name, nickname, city, city_alternate, state, country, franchise_id, nba_url_name, nba_short_name, nba_team_id, created_at, updated_at
		FROM nba.team t
		WHERE nba_team_id = ANY($1)`

	rows, err := d.pgxPool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	var teams []api.Team

	for rows.Next() {
		var team api.Team
		err = rows.Scan(
			&team.ID,
			&team.Name,
			&team.Nickname,
			&team.City,
			&team.AlternateCity,
			&team.State,
			&team.Country,
			&team.FranchiseID,
			&team.NBAURLName,
			&team.NBAShortName,
			&team.NBATeamID,
			&team.CreatedAt,
			&team.UpdatedAt)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (d DB) ListTeams(ctx context.Context) ([]api.Team, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.ListTeams")
	defer span.End()

	query := `
		SELECT id, t.name, nickname, city, city_alternate, state, country, nba_url_name, nba_short_name, nba_team_id, created_at, updated_at
		FROM nba.team t`

	rows, err := d.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	teams := []api.Team{}

	for rows.Next() {
		var team api.Team
		err = rows.Scan(
			&team.ID,
			&team.Name,
			&team.Nickname,
			&team.City,
			&team.AlternateCity,
			&team.State,
			&team.Country,
			&team.FranchiseID,
			&team.NBAURLName,
			&team.NBAShortName,
			&team.NBATeamID,
			&team.CreatedAt,
			&team.UpdatedAt)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (d DB) UpdateTeams(ctx context.Context, teams []team.TeamUpdate) ([]api.Team, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.UpdateTeams")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		slog.Error("could not start db transaction with error: %v", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	insertTeam := `
		INSERT INTO nba.team
			(name, nickname, city, city_alternate, franchise_id, nba_url_name, nba_short_name, nba_team_id)
		VALUES ($1, $2, $3, $4, (SELECT id FROM nba.franchise WHERE nba_team_id = $5), $6, $7, $5)
		ON CONFLICT (nba_team_id) DO UPDATE
		SET 
			name = coalesce(excluded.name, nba.team.name),
			nickname = coalesce(excluded.nickname, nba.team.nickname),
			city = coalesce(excluded.city, nba.team.city),
			city_alternate = coalesce(excluded.city_alternate, nba.team.city_alternate),
			franchise_id = coalesce(excluded.franchise_id, nba.team.franchise_id),
			nba_url_name = coalesce(excluded.nba_url_name, nba.team.nba_url_name),
			nba_short_name = coalesce(excluded.nba_short_name, nba.team.nba_short_name),
			nba_team_id = coalesce(excluded.nba_team_id, nba.team.nba_team_id)
		RETURNING nba.team.*`

	b := &pgx.Batch{}

	for _, team := range teams {
		b.Queue(
			insertTeam,
			team.Name,
			team.Nickname,
			team.City,
			team.AlternateCity,
			team.NBATeamID,
			team.NBAURLName,
			team.NBAShortName)
	}

	batchResults := tx.SendBatch(ctx, b)

	var updatedTeams []api.Team

	for _ = range teams {
		var t api.Team
		err = batchResults.QueryRow().Scan(
			&t.ID,
			&t.Name,
			&t.Nickname,
			&t.City,
			&t.AlternateCity,
			&t.State,
			&t.Country,
			&t.FranchiseID,
			&t.NBAURLName,
			&t.NBAShortName,
			&t.NBATeamID,
			&t.CreatedAt,
			&t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		updatedTeams = append(updatedTeams, t)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return updatedTeams, nil
}

func (d DB) NBATeamIDMappings(ctx context.Context) (map[string]string, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.NBATeamIDMappings")
	defer span.End()

	query := `
		SELECT id, nba_team_id
		FROM nba.team t`

	rows, err := d.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	mappings := map[string]string{}

	for rows.Next() {
		teamID := ""
		nbaTeamID := ""
		err = rows.Scan(&teamID, &nbaTeamID)
		if err != nil {
			return nil, err
		}
		mappings[nbaTeamID] = teamID
	}

	return mappings, nil
}
