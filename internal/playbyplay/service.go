package playbyplay

import (
	"context"
	"fmt"
	"os"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/util"
)

type Service interface {
	FetchPlayByPlayForGame(ctx context.Context, gameID string) (nba.PlayByPlay, error)
	UpdatePlayByPlayForGames(ctx context.Context, nbaGameIDs []string) ([]api.PlayByPlay, error)
}

func NewService(nbaClient nba.Client, r2Client cloudflare.Client) Service {
	return &service{nbaClient: nbaClient, r2Client: r2Client}
}

type service struct {
	playByPlayStore PlayByPlayWriter

	nbaClient nba.Client
	r2Client  cloudflare.Client
}

func (s service) FetchPlayByPlayForGame(ctx context.Context, gameID string) (nba.PlayByPlay, error) {
	filepath := fmt.Sprintf(os.Getenv("STORAGE_PATH")+"/playbyplay/%s.json", gameID)
	playByPlay, err := s.nbaClient.PlayByPlayForGame(ctx, gameID, util.WithFileOutputWriter(filepath), util.WithR2OutputWriter(s.r2Client, util.NBAR2Bucket, "playbyplay"))
	if err != nil {
		return nba.PlayByPlay{}, fmt.Errorf("failed to get play by play for game: %w", err)
	}

	return playByPlay, nil
}

func (s service) UpdatePlayByPlayForGames(ctx context.Context, nbaGameIDs []string) ([]api.PlayByPlay, error) {
	var playByPlayUpdates []PlayByPlayUpdate

	for _, nbaGameID := range nbaGameIDs {
		_, err := s.FetchPlayByPlayForGame(ctx, nbaGameID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playbyplay for game: %w", err)
		}

		update := PlayByPlayUpdate{}

		playByPlayUpdates = append(playByPlayUpdates, update)
	}

	return s.playByPlayStore.UpdatePlayByPlays(ctx, playByPlayUpdates)
}
