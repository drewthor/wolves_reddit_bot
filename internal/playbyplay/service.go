package playbyplay

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/util"
	"go.opentelemetry.io/otel"
)

type Service interface {
	FetchPlayByPlayForGame(ctx context.Context, logger *slog.Logger, gameID string) (nba.PlayByPlay, error)
	UpdatePlayByPlayForGames(ctx context.Context, logger *slog.Logger, nbaGameIDs []string) ([]api.PlayByPlay, error)
}

func NewService(nbaClient nba.Client, r2Client cloudflare.Client, playByPlayStore PlayByPlayWriter) Service {
	return &service{nbaClient: nbaClient, r2Client: r2Client, playByPlayStore: playByPlayStore}
}

type service struct {
	playByPlayStore PlayByPlayWriter

	nbaClient nba.Client
	r2Client  cloudflare.Client
}

func (s service) FetchPlayByPlayForGame(ctx context.Context, logger *slog.Logger, gameID string) (nba.PlayByPlay, error) {
	ctx, span := otel.Tracer("playbyplay").Start(ctx, "playbyplay.service.FetchPlayByPlayForGame")
	defer span.End()

	//filepath := fmt.Sprintf(os.Getenv("STORAGE_PATH")+"/playbyplay/%s.json", gameID)
	objectKey := fmt.Sprintf("playbyplay/%s.json", gameID)
	playByPlay, err := s.nbaClient.PlayByPlayForGame(ctx, gameID, util.WithR2OutputWriter(logger, s.r2Client, util.NBAR2Bucket, objectKey))
	if err != nil {
		return nba.PlayByPlay{}, fmt.Errorf("failed to get play by play for game: %w", err)
	}

	return playByPlay, nil
}

func (s service) UpdatePlayByPlayForGames(ctx context.Context, logger *slog.Logger, nbaGameIDs []string) ([]api.PlayByPlay, error) {
	ctx, span := otel.Tracer("playbyplay").Start(ctx, "playbyplay.service.UpdatePlayByPlayForGames")
	defer span.End()

	var playByPlayUpdates []PlayByPlayUpdate

	for _, nbaGameID := range nbaGameIDs {
		_, err := s.FetchPlayByPlayForGame(ctx, logger, nbaGameID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playbyplay for game: %w", err)
		}

		update := PlayByPlayUpdate{}

		playByPlayUpdates = append(playByPlayUpdates, update)
	}

	return s.playByPlayStore.UpdatePlayByPlays(ctx, playByPlayUpdates)
}
