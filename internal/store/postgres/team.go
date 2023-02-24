package postgres

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/internal/team"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

func (d DB) GetTeamWithID(ctx context.Context, teamID string) (api.Team, error) {
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
		team := api.Team{}
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
	query := `
		SELECT id, t.name, nickname, city, city_alternate, state, country, franchise_id, nba_url_name, nba_short_name, nba_team_id, created_at, updated_at
		FROM nba.team t
		WHERE nba_team_id = ANY($1)`

	rows, err := d.pgxPool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	teams := []api.Team{}

	for rows.Next() {
		team := api.Team{}
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
	query := `
		SELECT id, t.name, nickname, city, city_alternate, state, country, nba_url_name, nba_short_name, nba_team_id, created_at, updated_at
		FROM nba.team t`

	rows, err := d.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	teams := []api.Team{}

	for rows.Next() {
		team := api.Team{}
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
	tx, err := d.pgxPool.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

	insertTeam := `
		INSERT INTO nba.team
			as t(name, nickname, city, city_alternate, nba_url_name, nba_short_name, nba_team_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (nba_team_id) DO UPDATE
		SET 
			name = coalesce(excluded.name, t.name),
			nickname = coalesce(excluded.nickname, t.nickname),
			city = coalesce(excluded.city, t.city),
			city_alternate = coalesce(excluded.city_alternate, t.city_alternate),
			nba_url_name = coalesce(excluded.nba_url_name, t.nba_url_name),
			nba_short_name = coalesce(excluded.nba_short_name, t.nba_short_name),
			nba_team_id = coalesce(excluded.nba_team_id, t.nba_team_id)
		RETURNING t.id`

	b := &pgx.Batch{}

	for _, team := range teams {
		b.Queue(
			insertTeam,
			team.Name,
			team.Nickname,
			team.City,
			team.AlternateCity,
			team.NBAURLName,
			team.NBAShortName,
			team.NBATeamID)
	}

	batchResults := tx.SendBatch(ctx, b)

	insertedTeamIDs := []uuid.UUID{}

	for _ = range teams {
		var id uuid.UUID
		err := batchResults.QueryRow().Scan(&id)
		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedTeamIDs = append(insertedTeamIDs, id)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return d.GetTeamsWithIDs(ctx, insertedTeamIDs)
}

func (d DB) NBATeamIDMappings(ctx context.Context) (map[string]string, error) {
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
