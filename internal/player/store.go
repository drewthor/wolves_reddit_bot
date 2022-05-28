package player

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Store interface {
	Get(playerID string) (api.Player, error)
	GetAll() ([]api.Player, error)
	GetByIDs(ids []string) ([]api.Player, error)
	UpdatePlayers(players []api.Player) ([]api.Player, error)
}

func NewStore(db *pgxpool.Pool) Store {
	return &store{DB: db}
}

type store struct {
	DB *pgxpool.Pool
}

func (s *store) Get(playerID string) (api.Player, error) {
	player := api.Player{}

	query := `
		SELECT id, first_name, last_name, birthdate, height_feet, height_inches, height_meters, weight_pounds, weight_kilograms, jersey_number, positions.pos_array, currently_in_nba, years_pro, nba_debut_year, nba_player_id, country, time_created, time_modified 
		FROM nba.player p, LATERAL (
		        SELECT ARRAY (
		            SELECT pos.name 
					FROM nba.position pos
		            JOIN nba.player_position pp ON pos.id = pp.position_id
		            WHERE pp.player_id = p.id
		            ORDER BY pp.priority asc
                ) as pos_array
        ) positions
		WHERE p.id = $1`

	err := s.DB.QueryRow(context.Background(), query, playerID).Scan(
		&player.ID,
		&player.FirstName,
		&player.LastName,
		&player.Birthdate,
		&player.HeightFeet,
		&player.HeightInches,
		&player.HeightMeters,
		&player.WeightPounds,
		&player.WeightKilograms,
		&player.JerseyNumber,
		&player.Positions,
		&player.CurrentlyInNBA,
		&player.YearsPro,
		&player.NBADebutYear,
		&player.NBAPlayerID,
		&player.Country,
		&player.TimeCreated,
		&player.TimeModified)
	if err != nil {
		return api.Player{}, err
	}

	return player, nil
}

func (s *store) GetAll() ([]api.Player, error) {
	query := `
		SELECT id, first_name, last_name, birthdate, height_feet, height_inches, height_meters, weight_pounds, weight_kilograms, jersey_number, positions.pos_array, currently_in_nba, years_pro, nba_debut_year, nba_player_id, country, time_created, time_modified 
		FROM nba.player p, LATERAL (
		        SELECT ARRAY (
		            SELECT pos.name 
					FROM nba.position pos
		            JOIN nba.player_position pp ON pos.id = pp.position_id
		            WHERE pp.player_id = p.id
		            ORDER BY pp.priority asc
                ) as pos_array
        ) positions`

	rows, err := s.DB.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	players := []api.Player{}

	for rows.Next() {
		player := api.Player{}
		err = rows.Scan(
			&player.ID,
			&player.FirstName,
			&player.LastName,
			&player.Birthdate,
			&player.HeightFeet,
			&player.HeightInches,
			&player.HeightMeters,
			&player.WeightPounds,
			&player.WeightKilograms,
			&player.JerseyNumber,
			&player.Positions,
			&player.CurrentlyInNBA,
			&player.YearsPro,
			&player.NBADebutYear,
			&player.NBAPlayerID,
			&player.Country,
			&player.TimeCreated,
			&player.TimeModified)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func (s *store) GetByIDs(ids []string) ([]api.Player, error) {
	query := `
		SELECT id, first_name, last_name, birthdate, height_feet, height_inches, height_meters, weight_pounds, weight_kilograms, jersey_number, positions.pos_array, currently_in_nba, years_pro, nba_debut_year, nba_player_id, country, time_created, time_modified 
		FROM nba.player p, LATERAL (
		        SELECT ARRAY (
		            SELECT pos.name 
					FROM nba.position pos
		            JOIN nba.player_position pp ON pos.id = pp.position_id
		            WHERE pp.player_id = p.id
		            ORDER BY pp.priority asc
                ) as pos_array
        ) positions
		WHERE id = ANY($1)`

	rows, err := s.DB.Query(context.Background(), query, ids)
	if err != nil {
		return nil, err
	}

	players := []api.Player{}

	for rows.Next() {
		player := api.Player{}
		err = rows.Scan(
			&player.ID,
			&player.FirstName,
			&player.LastName,
			&player.Birthdate,
			&player.HeightFeet,
			&player.HeightInches,
			&player.HeightMeters,
			&player.WeightPounds,
			&player.WeightKilograms,
			&player.JerseyNumber,
			&player.Positions,
			&player.CurrentlyInNBA,
			&player.YearsPro,
			&player.NBADebutYear,
			&player.NBAPlayerID,
			&player.Country,
			&player.TimeCreated,
			&player.TimeModified)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func (s *store) UpdatePlayers(players []api.Player) ([]api.Player, error) {
	tx, err := s.DB.Begin(context.Background())
	defer tx.Rollback(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

	insertPlayer := `
		INSERT INTO nba.player
			as p(first_name, last_name, birthdate, height_feet, height_inches, height_meters, weight_pounds, weight_kilograms, jersey_number, currently_in_nba, years_pro, nba_debut_year, nba_player_id, country)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (nba_player_id) DO UPDATE
		SET 
			first_name = coalesce(excluded.first_name, p.first_name),
			last_name = coalesce(excluded.last_name, p.last_name),
			birthdate = coalesce(excluded.birthdate, p.birthdate),
			height_feet = coalesce(excluded.height_feet, p.height_feet),
			height_inches = coalesce(excluded.height_inches, p.height_inches),
			height_meters = coalesce(excluded.height_meters, p.height_meters),
			weight_pounds = coalesce(excluded.weight_pounds, p.weight_pounds),
			weight_kilograms = coalesce(excluded.weight_kilograms, p.weight_kilograms),
			jersey_number = coalesce(excluded.jersey_number, p.jersey_number),
			currently_in_nba = excluded.currently_in_nba,
			years_pro = coalesce(excluded.years_pro, p.years_pro),
			nba_debut_year = coalesce(excluded.nba_debut_year, p.nba_debut_year),
			nba_player_id = coalesce(excluded.nba_player_id, p.nba_player_id),
			country = coalesce(excluded.country, p.country)
		RETURNING p.id`

	removeExistingPlayerPositions := `
	DELETE FROM nba.player_position
	WHERE player_id = $1`

	insertPlayerPositions := `
		INSERT INTO nba.player_position
			as pp(player_id, position_id, priority)
		VALUES ($1, (SELECT id FROM nba.position WHERE name = $2), $3)`

	bp := &pgx.Batch{}

	for _, player := range players {
		bp.Queue(insertPlayer,
			player.FirstName,
			player.LastName,
			player.Birthdate,
			player.HeightFeet,
			player.HeightInches,
			player.HeightMeters,
			player.WeightPounds,
			player.WeightKilograms,
			player.JerseyNumber,
			player.CurrentlyInNBA,
			player.YearsPro,
			player.NBADebutYear,
			player.NBAPlayerID,
			player.Country)
	}

	bpp := &pgx.Batch{}

	batchResults := tx.SendBatch(context.Background(), bp)

	insertedPlayerIDs := []string{}
	numPlayerPositions := 0

	for _, player := range players {
		id := ""
		err := batchResults.QueryRow().Scan(&id)
		if err != nil {
			return nil, err
		}

		insertedPlayerIDs = append(insertedPlayerIDs, id)

		bpp.Queue(removeExistingPlayerPositions, id)
		for j := range player.Positions {
			bpp.Queue(insertPlayerPositions, id, player.Positions[j], j)
			numPlayerPositions++
		}
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	batchResults = tx.SendBatch(context.Background(), bpp)

	for i := 0; i < numPlayerPositions; i++ {
		_, err = batchResults.Exec()
		if err != nil {
			return nil, err
		}
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return s.GetByIDs(insertedPlayerIDs)
}
