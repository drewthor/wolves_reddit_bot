package franchise

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/internal/team"
	"github.com/drewthor/wolves_reddit_bot/internal/team_season"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type Service interface {
	UpdateFranchises(ctx context.Context, logger *slog.Logger) ([]api.Franchise, error)
}

func NewService(franchiseStore Store, teamService team.Service, teamSeasonService team_season.Service, nbaClient nba.Client) Service {
	return &service{
		franchiseStore:    franchiseStore,
		teamService:       teamService,
		teamSeasonService: teamSeasonService,
		nbaClient:         nbaClient,
	}
}

type service struct {
	franchiseStore    Store
	teamService       team.Service
	teamSeasonService team_season.Service

	nbaClient nba.Client
}

func (s service) UpdateFranchises(ctx context.Context, logger *slog.Logger) ([]api.Franchise, error) {
	ctx, span := otel.Tracer("team").Start(ctx, "franchise.service.UpdateFranchises")
	defer span.End()

	franchises, err := s.nbaClient.FranchiseHistory(ctx, "00")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to get franchise history when updating teams: %w", err)
	}

	franchiseUpdates := make([]FranchiseUpdate, len(franchises))
	for i, franchise := range franchises {
		franchiseUpdates[i] = FranchiseUpdate{
			NBALeagueID:        franchise.LeagueID,
			NBATeamID:          franchise.TeamID,
			Name:               franchise.Name,
			City:               franchise.City,
			StartYear:          franchise.StartYear,
			EndYear:            franchise.EndYear,
			Years:              franchise.Years,
			Games:              franchise.Games,
			Wins:               franchise.Wins,
			Losses:             franchise.Losses,
			PlayoffAppearances: franchise.PlayoffAppearances,
			DivisionTitles:     franchise.DivisionTitles,
			ConferenceTitles:   franchise.ConferenceTitles,
			LeagueTitles:       franchise.LeagueTitles,
			Active:             franchise.Active,
		}
	}

	updatedFranchises, err := s.franchiseStore.UpdateFranchises(ctx, franchiseUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to store updated franchises: %w", err)
	}

	if _, err = s.teamService.UpdateFranchiseTeams(ctx, franchises); err != nil {
		return nil, fmt.Errorf("failed to update franchise teams when updating franchises: %w", err)
	}

	if _, err = s.teamSeasonService.UpdateFranchiseTeamSeasons(ctx, "00", franchises); err != nil {
		return nil, fmt.Errorf("failed to update franchise team seasons when updating franchises: %w", err)
	}

	return updatedFranchises, nil
}
