package player

import (
	"context"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
)

type Service interface {
	Get(ctx context.Context, playerID string) (api.Player, error)
	ListPlayers(ctx context.Context) ([]api.Player, error)
	UpdatePlayers(ctx context.Context, seasonStartYear int) ([]api.Player, error)
}

func NewService(playerStore Store) Service {
	return &service{PlayerStore: playerStore}
}

type service struct {
	PlayerStore Store
}

func (s *service) Get(ctx context.Context, playerID string) (api.Player, error) {
	ctx, span := otel.Tracer("player").Start(ctx, "player.service.Get")
	defer span.End()

	player, err := s.PlayerStore.GetPlayerWithID(ctx, playerID)
	if err != nil {
		return player, err
	}
	return player, nil
}

func (s *service) ListPlayers(ctx context.Context) ([]api.Player, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "player.service.ListPlayers")
	defer span.End()

	players, err := s.PlayerStore.ListPlayers(ctx)
	if err != nil {
		return nil, err
	}
	return players, nil
}

func (s *service) UpdatePlayers(ctx context.Context, seasonStartYear int) ([]api.Player, error) {
	ctx, span := otel.Tracer("player").Start(ctx, "player.service.UpdatePlayers")
	defer span.End()

	players, err := s.getSeasonPlayers(ctx, seasonStartYear)
	if err != nil {
		return nil, err
	}
	updatedPlayers, err := s.PlayerStore.UpdatePlayers(ctx, players)
	if err != nil {
		return nil, err
	}

	return updatedPlayers, nil
}

func (s *service) getSeasonPlayers(ctx context.Context, seasonStartYear int) ([]api.Player, error) {
	ctx, span := otel.Tracer("player").Start(ctx, "player.service.getSeasonPlayers")
	defer span.End()

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
			Active:          nbaPlayer.CurrentlyInNBA,
			YearsPro:        yearsPro,
			NBADebutYear:    nbaDebutYear,
			NBAPlayerID:     nbaPlayerID,
			Country:         country,
		}
		players = append(players, player)
	}
	return players, nil
}
