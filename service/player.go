package service

import (
	"log"
	"strconv"
	"time"

	"github.com/drewthor/wolves_reddit_bot/dao"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
)

type PlayerService struct {
	PlayerDAO *dao.PlayerDAO
}

func (ps PlayerService) Get(playerID string) (api.Player, error) {
	player, err := ps.PlayerDAO.Get(playerID)
	if err != nil {
		return player, err
	}
	return player, nil
}

func (ps PlayerService) GetAll() ([]api.Player, error) {
	players, err := ps.PlayerDAO.GetAll()
	if err != nil {
		return nil, err
	}
	return players, nil
}

func (ps PlayerService) UpdatePlayers(seasonStartYear int) ([]api.Player, error) {
	players, err := ps.getSeasonPlayers(seasonStartYear)
	if err != nil {
		return nil, err
	}
	updatedPlayers, err := ps.PlayerDAO.UpdatePlayers(players)
	if err != nil {
		return nil, err
	}

	return updatedPlayers, nil
}

func (ps PlayerService) getSeasonPlayers(seasonStartYear int) ([]api.Player, error) {
	nbaPlayers, err := nba.GetPlayers(seasonStartYear)

	if err != nil {
		return nil, err
	}

	players := []api.Player{}
	for _, nbaPlayer := range nbaPlayers {
		birthdate := new(time.Time)
		*birthdate, err = time.Parse(nba.TimeBirthdateFormat, nbaPlayer.DateOfBirthUTC)
		if err != nil {
		}

		heightFeet := new(int)
		*heightFeet, err = strconv.Atoi(nbaPlayer.HeightFeet)
		if err != nil {
		}

		heightInches := new(int)
		*heightInches, err = strconv.Atoi(nbaPlayer.HeightInches)
		if err != nil {
		}

		heightMeters := new(float64)
		*heightMeters, err = strconv.ParseFloat(nbaPlayer.HeightMeters, 64)
		if err != nil {
		}

		weightPounds := new(int)
		*weightPounds, err = strconv.Atoi(nbaPlayer.WeightPounds)
		if err != nil {
		}

		weightKilograms := new(float64)
		*weightKilograms, err = strconv.ParseFloat(nbaPlayer.WeightKilograms, 64)
		if err != nil {
		}

		jerseyNumber := new(int)
		if len(nbaPlayer.Jersey) > 0 {
			*jerseyNumber, err = strconv.Atoi(nbaPlayer.Jersey)
			if err != nil {
				continue
			}
		}

		position := api.PositionFromNBAPosition(nbaPlayer.Position)

		yearsPro := new(int)
		*yearsPro, err = strconv.Atoi(nbaPlayer.YearsPro)
		if err != nil {
		}

		nbaDebutYear := new(int)
		if len(nbaPlayer.NBADebutYear) > 0 {
			*nbaDebutYear, err = strconv.Atoi(nbaPlayer.NBADebutYear)
			if err != nil {
			}
		}

		country := new(string)
		if nbaPlayer.Country != "" {
			*country = nbaPlayer.Country
		}

		nbaPlayerID, err := strconv.Atoi(nbaPlayer.ID)
		if err != nil {
			log.Println("could not convert player id: ", nbaPlayer.ID, " to int for player with first name: ", nbaPlayer.FirstName, " last name: ", nbaPlayer.LastName, " for season: ", seasonStartYear)
		}

		player := api.Player{
			FirstName:       nbaPlayer.FirstName,
			LastName:        nbaPlayer.LastName,
			Birthdate:       birthdate,
			HeightFeet:      heightFeet,
			HeightInches:    heightInches,
			HeightMeters:    heightMeters,
			WeightPounds:    weightPounds,
			WeightKilograms: weightKilograms,
			JerseyNumber:    jerseyNumber,
			Positions:       position,
			CurrentlyInNBA:  nbaPlayer.CurrentlyInNBA,
			YearsPro:        yearsPro,
			NBADebutYear:    nbaDebutYear,
			NBAPlayerID:     nbaPlayerID,
			Country:         country,
		}
		players = append(players, player)
	}
	return players, nil
}
