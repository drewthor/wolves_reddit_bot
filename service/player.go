package service

import (
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

func (ps PlayerService) getAllPlayersFromNBAApi() ([]api.Player, error) {
	nbaPlayers, err := nba.GetPlayers(nba.GetDailyAPIPaths().Players)

	if err != nil {
		return nil, err
	}

	players := []api.Player{}
	for _, nbaPlayer := range nbaPlayers {
		birthdate, err := time.Parse(nba.TimeBirthdateFormat, nbaPlayer.DateOfBirthUTC)
		if err != nil {
			continue
			// return nil, err
		}

		heightFeet, err := strconv.Atoi(nbaPlayer.HeightFeet)
		if err != nil {
			continue
			// return nil, err
		}

		heightInches, err := strconv.Atoi(nbaPlayer.HeightInches)
		if err != nil {
			continue
			// return nil, err
		}

		heightMeters, err := strconv.ParseFloat(nbaPlayer.HeightMeters, 64)
		if err != nil {
			continue
			// return nil, err
		}

		weightPounds, err := strconv.Atoi(nbaPlayer.WeightPounds)
		if err != nil {
			continue
			// return nil, err
		}

		weightKilograms, err := strconv.ParseFloat(nbaPlayer.WeightKilograms, 64)
		if err != nil {
			continue
			// return nil, err
		}

		jerseyNumber := new(int)
		if len(nbaPlayer.Jersey) > 0 {
			*jerseyNumber, err = strconv.Atoi(nbaPlayer.Jersey)
			if err != nil {
				continue
				// return nil, err
			}
		}

		position := api.PositionFromNBAPosition(nbaPlayer.Position)

		yearsPro, err := strconv.Atoi(nbaPlayer.YearsPro)
		if err != nil {
			continue
			// return nil, err
		}

		nbaDebutYear := new(int)
		if len(nbaPlayer.NBADebutYear) > 0 {
			*nbaDebutYear, err = strconv.Atoi(nbaPlayer.NBADebutYear)
			if err != nil {
				continue
				// return nil, err
			}
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
			NBAPlayerID:     nbaPlayer.ID,
			Country:         nbaPlayer.Country,
		}
		players = append(players, player)
	}
	return players, nil
}

func (ps PlayerService) UpdatePlayers() ([]api.Player, error) {
	players, err := ps.getAllPlayersFromNBAApi()
	if err != nil {
		return nil, err
	}
	updatedPlayers, err := ps.PlayerDAO.UpdatePlayers(players)
	return updatedPlayers, nil
}
