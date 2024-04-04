package nba

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const (
	franchiseHistoryURL = "https://stats.nba.com/stats/franchisehistory?"
)

type Franchise struct {
	LeagueID           string
	TeamID             int
	City               string
	Name               string
	StartYear          int
	EndYear            int
	Years              int
	Games              int
	Wins               int
	Losses             int
	PlayoffAppearances int
	DivisionTitles     int
	ConferenceTitles   int
	LeagueTitles       int
	Active             bool
	TeamSeasons        []TeamSeason
}

type TeamSeason struct {
	Year int // year the season started
	City string
	Name string
}

func (c Client) FranchiseHistory(ctx context.Context, leagueID string) ([]Franchise, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "nba.Client.FranchisesAndTeams")
	defer span.End()

	urlValues := url.Values{
		"LeagueID": {leagueID},
	}
	franchiseURL := franchiseHistoryURL + urlValues.Encode()

	req, err := retryablehttp.NewRequest(http.MethodGet, franchiseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to get franchise history: %w", err)
	}
	response, err := c.statsClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get franchise history from nba from url %s: %w", franchiseURL, err)
	}
	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") == "gzip" {
		response.Body, err = gzip.NewReader(response.Body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to create gzip reader when getting nba franchise history: %w", err)
		}
	}

	franchiseResult, err := unmarshalNBAHttpResponseToJSON[statsBaseResponse](response.Body)
	if err != nil {
		respBody, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("failed to unmarshal json for franchise history from url %s: %s: %w", franchiseURL, respBody, err)
	}

	var franchises []Franchise

	for _, resultSet := range franchiseResult.ResultSets {
		headersMap := make(map[string]int, len(resultSet.Headers))
		for i, header := range resultSet.Headers {
			headersMap[header] = i
		}

		switch resultSet.Name {
		case "FranchiseHistory":
			activeFranchises, err := parseFranchiseHistoryRow(ctx, leagueID, true, headersMap, resultSet.RowSet)
			if err != nil {
				return nil, fmt.Errorf("failed to parse active franchises: %w", err)
			}
			franchises = append(franchises, activeFranchises...)
		case "DefunctTeams":
			inactiveFranchises, err := parseFranchiseHistoryRow(ctx, leagueID, false, headersMap, resultSet.RowSet)
			if err != nil {
				return nil, fmt.Errorf("failed to parse inactive franchises: %w", err)
			}

			for i := range inactiveFranchises {
				for j := 0; j < inactiveFranchises[i].Years; j++ {
					inactiveFranchises[i].TeamSeasons = append(inactiveFranchises[i].TeamSeasons, TeamSeason{
						Year: inactiveFranchises[i].StartYear + j,
						City: inactiveFranchises[i].City,
						Name: inactiveFranchises[i].Name,
					})
				}
			}
			franchises = append(franchises, inactiveFranchises...)
		}
	}

	return franchises, nil
}

func parseFranchiseHistoryRow(ctx context.Context, leagueID string, active bool, headersMap map[string]int, rowSets [][]json.RawMessage) ([]Franchise, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "nba.Client.parseFranchiseHistoryRow")
	defer span.End()

	teamIDfranchiseMap := make(map[int]Franchise)
	for _, rowSet := range rowSets {
		teamID, err := parseRowSetValue[int](headersMap, rowSet, "TEAM_ID")
		if err != nil {
			return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
		}

		city, err := parseRowSetValue[string](headersMap, rowSet, "TEAM_CITY")
		if err != nil {
			return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
		}

		name, err := parseRowSetValue[string](headersMap, rowSet, "TEAM_NAME")
		if err != nil {
			return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
		}

		startYear, err := parseRowSetValue[string](headersMap, rowSet, "START_YEAR")
		if err != nil {
			return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
		}

		startYearInt, err := strconv.Atoi(startYear)
		if err != nil {
			return nil, fmt.Errorf("failed to convert start year from franchise history to int: %w", err)
		}

		endYear, err := parseRowSetValue[string](headersMap, rowSet, "END_YEAR")
		if err != nil {
			return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
		}

		endYearInt, err := strconv.Atoi(endYear)
		if err != nil {
			return nil, fmt.Errorf("failed to convert end year from franchise history to int: %w", err)
		}

		if teamFranchise, ok := teamIDfranchiseMap[teamID]; ok {
			teamSeason := TeamSeason{
				Year: startYearInt,
				City: city,
				Name: name,
			}
			teamFranchise.TeamSeasons = append(teamFranchise.TeamSeasons, teamSeason)
			teamIDfranchiseMap[teamID] = teamFranchise
		} else {
			years, err := parseRowSetValue[int](headersMap, rowSet, "YEARS")
			if err != nil {
				return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
			}

			games, err := parseRowSetValue[int](headersMap, rowSet, "GAMES")
			if err != nil {
				return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
			}

			wins, err := parseRowSetValue[int](headersMap, rowSet, "WINS")
			if err != nil {
				return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
			}

			losses, err := parseRowSetValue[int](headersMap, rowSet, "LOSSES")
			if err != nil {
				return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
			}

			playoffAppearances, err := parseRowSetValue[int](headersMap, rowSet, "PO_APPEARANCES")
			if err != nil {
				return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
			}

			divisionTitles, err := parseRowSetValue[int](headersMap, rowSet, "DIV_TITLES")
			if err != nil {
				return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
			}

			conferenceTitles, err := parseRowSetValue[int](headersMap, rowSet, "CONF_TITLES")
			if err != nil {
				return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
			}

			leagueTitles, err := parseRowSetValue[int](headersMap, rowSet, "LEAGUE_TITLES")
			if err != nil {
				return nil, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameLeagueGameLog, err)
			}

			teamIDfranchiseMap[teamID] = Franchise{
				LeagueID:           leagueID,
				TeamID:             teamID,
				City:               city,
				Name:               name,
				StartYear:          startYearInt,
				EndYear:            endYearInt,
				Years:              years,
				Games:              games,
				Wins:               wins,
				Losses:             losses,
				PlayoffAppearances: playoffAppearances,
				DivisionTitles:     divisionTitles,
				ConferenceTitles:   conferenceTitles,
				LeagueTitles:       leagueTitles,
				Active:             active,
			}
		}
	}
	franchises := make([]Franchise, 0, len(teamIDfranchiseMap))
	for _, franchise := range teamIDfranchiseMap {
		franchises = append(franchises, franchise)
	}

	return franchises, nil
}
