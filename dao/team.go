package dao

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TeamDAO struct {
	DB *pgxpool.Pool
}

func (td *TeamDAO) Get(teamID string) (api.Team, error) {
	query := `
		SELECT id, t.name, nickname, city, city_alternate, state, country, league.name, season.name, conference.name, division.name, nba_url_name, nba_short_name, nba_team_id, created_at, updated_at
		FROM nba.team t, 
		LATERAL (
			SELECT l.name
			FROM nba.league l
			WHERE l.id = t.league_id
        ) league,
		LATERAL (
		        SELECT CONCAT(s.start_year, '-', s.end_year)
				FROM nba.season s
				WHERE s.id = t.season_id
		) season,
		LATERAL (
		        SELECT c.name
				FROM nba.conference c
				WHERE c.id = t.season_id
        ) conference,
		LATERAL (
		        SELECT d.name
				FROM nba.division d
				WHERE d.id = t.season_id
        ) division
		WHERE id = $1`

	team := api.Team{}

	err := td.DB.QueryRow(context.Background(), query, teamID).Scan(
		&team.ID,
		&team.Name,
		&team.Nickname,
		&team.City,
		&team.AlternateCity,
		&team.State,
		&team.Country,
		&team.League,
		&team.Season,
		&team.Conference,
		&team.Division,
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

func (td *TeamDAO) GetByIDs(ids []string) ([]api.Team, error) {
	query := `
		SELECT id, t.name, nickname, city, city_alternate, state, country, league.name, season.name, conference.name, division.name, nba_url_name, nba_short_name, nba_team_id, created_at, updated_at
		FROM nba.team t, 
		LATERAL (
			SELECT l.name
			FROM nba.league l
			WHERE l.id = t.league_id
        ) league,
		LATERAL (
		        SELECT CONCAT(s.start_year, '-', s.end_year) as name
				FROM nba.season s
				WHERE s.id = t.season_id
		) season,
		LATERAL (
		        SELECT c.name
				FROM nba.conference c
				WHERE c.id = t.conference_id
        ) conference,
		LATERAL (
		        SELECT d.name
				FROM nba.division d
				WHERE d.id = t.division_id
        ) division
		WHERE id = ANY($1)`

	rows, err := td.DB.Query(context.Background(), query, ids)
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
			&team.League,
			&team.Season,
			&team.Conference,
			&team.Division,
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

func (td *TeamDAO) GetAll() ([]api.Team, error) {
	query := `
		SELECT id, t.name, nickname, city, city_alternate, state, country, league.name, season.name, conference.name, division.name, nba_url_name, nba_short_name, nba_team_id, created_at, updated_at
		FROM nba.team t, 
		LATERAL (
			SELECT l.name
			FROM nba.league l
			WHERE l.id = t.league_id
        ) league,
		LATERAL (
		        SELECT CONCAT(s.start_year, '-', s.end_year) as name
				FROM nba.season s
				WHERE s.id = t.season_id
		) season,
		LATERAL (
		        SELECT c.name
				FROM nba.conference c
				WHERE c.id = t.conference_id
        ) conference,
		LATERAL (
		        SELECT d.name
				FROM nba.division d
				WHERE d.id = t.division_id
        ) division`

	rows, err := td.DB.Query(context.Background(), query)
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
			&team.League,
			&team.Season,
			&team.Conference,
			&team.Division,
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

func (td *TeamDAO) UpdateTeams(teams []api.Team) ([]api.Team, error) {
	tx, err := td.DB.Begin(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

	insertTeam := `
		INSERT INTO nba.team
			as t(name, nickname, city, city_alternate, league_id, season_id, conference_id, division_id, nba_url_name, nba_short_name, nba_team_id)
		VALUES ($1, $2, $3, $4, (SELECT id FROM nba.league WHERE name = $5), (SELECT id FROM nba.season WHERE start_year = $6), (SELECT id FROM nba.conference WHERE name = $7), (SELECT id FROM nba.division WHERE name = $8), $9, $10, $11)
		ON CONFLICT (nba_team_id, season_id) DO UPDATE
		SET 
			name = coalesce(excluded.name, t.name),
			nickname = coalesce(excluded.nickname, t.nickname),
			city = coalesce(excluded.city, t.city),
			city_alternate = coalesce(excluded.city_alternate, t.city_alternate),
			league_id = coalesce(excluded.league_id, t.league_id),
			season_id = coalesce(excluded.season_id, t.season_id),
			conference_id = coalesce(excluded.conference_id, t.conference_id),
			division_id = coalesce(excluded.division_id, t.division_id),
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
			team.League,
			team.Season,
			team.Conference,
			team.Division,
			team.NBAURLName,
			team.NBAShortName,
			team.NBATeamID)
	}

	batchResults := tx.SendBatch(context.Background(), b)

	insertedTeamIDs := []string{}

	for _ = range teams {
		id := ""
		err := batchResults.QueryRow().Scan(&id)
		if err != nil {
			return nil, err
		}

		insertedTeamIDs = append(insertedTeamIDs, id)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return td.GetByIDs(insertedTeamIDs)
}

func (td *TeamDAO) NBATeamIDMappings() (map[string]string, error) {
	query := `
		SELECT id, nba_team_id
		FROM nba.team t`

	rows, err := td.DB.Query(context.Background(), query)
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
