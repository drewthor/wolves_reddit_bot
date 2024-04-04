package nba

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const leagueGameLogURL = "https://stats.nba.com/stats/leaguegamelog?"

type SeasonType string

const (
	SeasonTypeRegular       SeasonType = "Regular Season"
	SeasonTypePre           SeasonType = "Pre Season"
	SeasonTypePlayoffs      SeasonType = "Playoffs"
	SeasonTypeAllStar       SeasonType = "All Star"
	SeasonTypeAllStarHyphen SeasonType = "All-Star"
)

type GameLog struct {
	GameID               string
	HomeTeam             TeamGameLog
	AwayTeamID           TeamGameLog
	TotalDurationMinutes int
}

type TeamGameLog struct {
	TeamID                 int
	FieldGoalsMade         int
	FieldGoalsAttempted    *int
	ThreePointersMade      *int
	ThreePointersAttempted *int
	FreeThrowsMade         int
	FreeThrowsAttempted    int
	OffensiveRebounds      *int
	DefensiveRebounds      *int
	TotalRebounds          *int
	Assists                *int
	Steals                 *int
	Blocks                 *int
	Turnovers              *int
	PersonalFouls          *int
	Points                 int
	PlusMinus              int
}

func (c *Client) LeagueGameLog(ctx context.Context, seasonStartYear int, seasonType SeasonType) ([]GameLog, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "nba.Client.GameLog")
	defer span.End()

	//if c.Cache != nil {
	//	obj, err := c.Cache.GetObject(ctx, objectKey)
	//	if err != nil && !errors.Is(err, ErrNotFound) {
	//		return nil, fmt.Errorf("failed to get boxscore summary from cache: %w", err)
	//	}
	//	data = obj
	//	slog.Info("found boxscore summary in cache")
	//}
	urlValues := url.Values{
		"Counter":      {"0"},
		"Direction":    {"ASC"},
		"LeagueID":     {"00"},
		"PlayerOrTeam": {"T"},
		"Season":       {strconv.Itoa(seasonStartYear)},
		"SeasonType":   {string(seasonType)},
		"Sorter":       {"DATE"},
	}

	u := leagueGameLogURL + urlValues.Encode()
	req, err := retryablehttp.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to create request to get league game log: %w", err)
	}

	response, err := c.statsClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to get GameLog object: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = fmt.Errorf("failed to successfully get league game log: status %d: url: %s", response.StatusCode, req.URL.String())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if response.Header.Get("Content-Encoding") == "gzip" {
		response.Body, err = gzip.NewReader(response.Body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to create gzip reader when getting nba league game log: %w", err)
		}
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to read all response data when getting nba league game log: %w", err)
	}

	gameLogsData, err := unmarshalNBAHttpResponseToJSON[statsBaseResponse](bytes.NewReader(data))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to get league game log response: %w", err)
	}

	gameLogs := make(map[string]GameLog)
	for _, resultSet := range gameLogsData.ResultSets {
		headersMap := make(map[string]int, len(resultSet.Headers))
		for i, header := range resultSet.Headers {
			headersMap[header] = i
		}

		switch resultSet.Name {
		case "LeagueGameLog":
			for _, rowSet := range resultSet.RowSet {

				teamID, err := parseRowSetValue[int](headersMap, rowSet, "TEAM_ID")
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				gameID, err := parseRowSetValue[string](headersMap, rowSet, "GAME_ID")
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return nil, fmt.Errorf("failed to parse gameID from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				matchup, err := parseRowSetValue[string](headersMap, rowSet, "MATCHUP")
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return nil, fmt.Errorf("failed to parse matchup from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				totalDurationMinutes, err := parseRowSetValue[int](headersMap, rowSet, "MIN")
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return nil, fmt.Errorf("failed to parse total duration minutes from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				fieldGoalsMade, err := parseRowSetValue[int](headersMap, rowSet, "FGM")
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					return nil, fmt.Errorf("failed to parse field goals made from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				fieldGoalsAttempted, err := parseRowSetValue[*int](headersMap, rowSet, "FGA")
				if err != nil {
					return nil, fmt.Errorf("failed to parse game code from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				threePointersMade, err := parseRowSetValue[*int](headersMap, rowSet, "FG3M")
				if err != nil {
					return nil, fmt.Errorf("failed to parse three pointers made from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				threePointersAttempted, err := parseRowSetValue[*int](headersMap, rowSet, "FG3A")
				if err != nil {
					return nil, fmt.Errorf("failed to parse three pointers attempted from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				freeThrowsMade, err := parseRowSetValue[int](headersMap, rowSet, "FTM")
				if err != nil {
					return nil, fmt.Errorf("failed to parse free throws made from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				freeThrowsAttempted, err := parseRowSetValue[int](headersMap, rowSet, "FTA")
				if err != nil {
					return nil, fmt.Errorf("failed to parse free throws attempted from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				offensiveRebounds, err := parseRowSetValue[*int](headersMap, rowSet, "OREB")
				if err != nil {
					return nil, fmt.Errorf("failed to parse offensive rebounds from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				defensiveRebounds, err := parseRowSetValue[*int](headersMap, rowSet, "DREB")
				if err != nil {
					return nil, fmt.Errorf("failed to parse defensive rebounds from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				totalRebounds, err := parseRowSetValue[*int](headersMap, rowSet, "REB")
				if err != nil {
					return nil, fmt.Errorf("failed to parse total rebounds from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				assists, err := parseRowSetValue[*int](headersMap, rowSet, "AST")
				if err != nil {
					return nil, fmt.Errorf("failed to parse assists from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				steals, err := parseRowSetValue[*int](headersMap, rowSet, "STL")
				if err != nil {
					return nil, fmt.Errorf("failed to parse steals rebounds from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				blocks, err := parseRowSetValue[*int](headersMap, rowSet, "BLK")
				if err != nil {
					return nil, fmt.Errorf("failed to parse blocks from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				turnovers, err := parseRowSetValue[*int](headersMap, rowSet, "TOV")
				if err != nil {
					return nil, fmt.Errorf("failed to parse turnovers from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				personalFouls, err := parseRowSetValue[*int](headersMap, rowSet, "PF")
				if err != nil {
					return nil, fmt.Errorf("failed to parse personal fouls from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				points, err := parseRowSetValue[int](headersMap, rowSet, "PTS")
				if err != nil {
					return nil, fmt.Errorf("failed to parse points from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				plusMinus, err := parseRowSetValue[int](headersMap, rowSet, "PLUS_MINUS")
				if err != nil {
					return nil, fmt.Errorf("failed to parse plus minus from nba stats %s endpoint: %w", endpointNameLeagueGameLog, err)
				}

				gameLog, ok := gameLogs[gameID]
				if !ok {
					gameLog = GameLog{
						GameID:               gameID,
						TotalDurationMinutes: totalDurationMinutes,
					}
				}

				teamGameLog := TeamGameLog{
					TeamID:                 teamID,
					FieldGoalsMade:         fieldGoalsMade,
					FieldGoalsAttempted:    fieldGoalsAttempted,
					ThreePointersMade:      threePointersMade,
					ThreePointersAttempted: threePointersAttempted,
					FreeThrowsMade:         freeThrowsMade,
					FreeThrowsAttempted:    freeThrowsAttempted,
					OffensiveRebounds:      offensiveRebounds,
					DefensiveRebounds:      defensiveRebounds,
					TotalRebounds:          totalRebounds,
					Assists:                assists,
					Steals:                 steals,
					Blocks:                 blocks,
					Turnovers:              turnovers,
					PersonalFouls:          personalFouls,
					Points:                 points,
					PlusMinus:              plusMinus,
				}

				awayTeam := strings.Contains(matchup, "@")

				if awayTeam {
					gameLog.AwayTeamID = teamGameLog
				} else {
					gameLog.HomeTeam = teamGameLog
				}

				gameLogs[gameID] = gameLog
			}
		}
	}

	leagueGameLogs := make([]GameLog, 0, len(gameLogs))
	for _, gameLog := range gameLogs {
		leagueGameLogs = append(leagueGameLogs, gameLog)
	}

	return leagueGameLogs, nil
}
