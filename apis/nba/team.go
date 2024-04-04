package nba

import (
	"compress/gzip"
	"context"
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
	teamLogoUrl           = "https://cdn.nba.com/logos/nba/%d/primary/L/logo.svg"
	teamCommonInfoBaseURL = "https://stats.nba.com/stats/teaminfocommon?"
	teamStandingsURL      = "https://stats.nba.com/stats/leaguestandingsv3?"
)

type TeamCommonInfo struct {
	TeamID          int
	SeasonStartYear int
	SeasonEndYear   int
	City            string
	Name            string
	Abbreviation    string
	Conference      string
	Division        string
	Code            string
	Slug            string
	SeasonIDs       []int
}

func (c Client) CommonTeamInfo(ctx context.Context, leagueID string, teamID int) (TeamCommonInfo, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "nba.Client.CommonTeamInfo")
	defer span.End()

	urlValues := url.Values{
		"LeagueID": {leagueID},
		"TeamID":   {strconv.Itoa(teamID)},
	}
	teamURL := teamCommonInfoBaseURL + urlValues.Encode()

	req, err := retryablehttp.NewRequest(http.MethodGet, teamURL, nil)
	if err != nil {
		return TeamCommonInfo{}, fmt.Errorf("failed to create request to get common team info: %w", err)
	}
	response, err := c.statsClient.Do(req)
	if err != nil {
		return TeamCommonInfo{}, fmt.Errorf("failed to get current teams from nba from url %s: %w", teamURL, err)
	}
	defer response.Body.Close()

	teamsResult, err := unmarshalNBAHttpResponseToJSON[statsBaseResponse](response.Body)
	if err != nil {
		gzipReader, gzipErr := gzip.NewReader(response.Body)
		if gzipErr != nil {
			return TeamCommonInfo{}, fmt.Errorf("failed to unmarshal json for nba team from url %s: %w: %w", teamURL, gzipErr, err)
		}
		defer gzipReader.Close()
		respBody, _ := io.ReadAll(gzipReader)
		return TeamCommonInfo{}, fmt.Errorf("failed to unmarshal json for nba team from url %s: %s: %w", teamURL, respBody, err)
	}

	teamCommonInfo := TeamCommonInfo{}

	for _, resultSet := range teamsResult.ResultSets {
		headersMap := make(map[string]int, len(resultSet.Headers))
		for i, header := range resultSet.Headers {
			headersMap[header] = i
		}

		switch resultSet.Name {
		case "TeamInfoCommon":
			if len(resultSet.RowSet) != 1 {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse TeamInfoCommon from nba stats TeamInfoCommon endpoint: expected 1 row set and got %d", len(resultSet.RowSet))
			}
			rowSet := resultSet.RowSet[0]

			tID, err := parseRowSetValue[int](headersMap, rowSet, "TEAM_ID")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameTeamInfoCommon, err)
			}

			seasonYear, err := parseRowSetValue[string](headersMap, rowSet, "SEASON_YEAR")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameTeamInfoCommon, err)
			}

			startYear, endYear, err := parseSeasonStartEndYears(seasonYear)
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse start and end years from nba stats TeamInfoCommon endpoint: %w", err)
			}

			city, err := parseRowSetValue[string](headersMap, rowSet, "TEAM_CITY")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse start and end years from nba stats TeamInfoCommon endpoint: %w", err)
			}

			name, err := parseRowSetValue[string](headersMap, rowSet, "TEAM_NAME")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse start and end years from nba stats TeamInfoCommon endpoint: %w", err)
			}

			abbreviation, err := parseRowSetValue[string](headersMap, rowSet, "TEAM_ABBREVIATION")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse start and end years from nba stats TeamInfoCommon endpoint: %w", err)
			}

			conference, err := parseRowSetValue[string](headersMap, rowSet, "TEAM_CONFERENCE")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse start and end years from nba stats TeamInfoCommon endpoint: %w", err)
			}

			division, err := parseRowSetValue[string](headersMap, rowSet, "TEAM_DIVISION")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse start and end years from nba stats TeamInfoCommon endpoint: %w", err)
			}

			code, err := parseRowSetValue[string](headersMap, rowSet, "TEAM_CODE")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse start and end years from nba stats TeamInfoCommon endpoint: %w", err)
			}

			slug, err := parseRowSetValue[string](headersMap, rowSet, "TEAM_SLUG")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse start and end years from nba stats TeamInfoCommon endpoint: %w", err)
			}

			teamCommonInfo.TeamID = tID
			teamCommonInfo.SeasonStartYear = startYear
			teamCommonInfo.SeasonEndYear = endYear
			teamCommonInfo.City = city
			teamCommonInfo.Name = name
			teamCommonInfo.Abbreviation = abbreviation
			teamCommonInfo.Conference = conference
			teamCommonInfo.Division = division
			teamCommonInfo.Code = code
			teamCommonInfo.Slug = slug

			break

		case "AvailableSeasons":
			var seasonIDs []int
			for _, rowSet := range resultSet.RowSet {
				seasonIDStr, err := parseRowSetValue[[]string](headersMap, rowSet, "SEASON_ID")
				if err != nil {
					return TeamCommonInfo{}, fmt.Errorf("failed to parse available season id from nba stats TeamInfoCommon endpoint: %w", err)
				}

				if len(seasonIDStr) != 1 {
					return TeamCommonInfo{}, fmt.Errorf("invalid available season id from nba stats TeamInfoCommon endpoint got %v", seasonIDStr)
				}

				seasonID, err := strconv.Atoi(seasonIDStr[0])
				if err != nil {
					return TeamCommonInfo{}, fmt.Errorf("failed to parse available season id string to int from nba stats TeamInfoCommon endpoint: %w", err)
				}

				seasonIDs = append(seasonIDs, seasonID)
			}

			teamCommonInfo.SeasonIDs = seasonIDs

			break
		}

	}

	return teamCommonInfo, nil
}

type TeamStanding struct {
	LeagueID        string
	SeasonStartYear int
	SeasonType      SeasonType
	TeamID          int
	Conference      string
	Division        string
}

func (c Client) TeamStandings(ctx context.Context, leagueID string, seasonStartYear int, seasonType SeasonType) ([]TeamStanding, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "nba.Client.TeamStandings")
	defer span.End()

	urlValues := url.Values{
		"LeagueID":   {leagueID},
		"Season":     {strconv.Itoa(seasonStartYear)},
		"SeasonType": {string(seasonType)},
	}
	u := teamStandingsURL + urlValues.Encode()

	req, err := retryablehttp.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to get team standings: %w", err)
	}
	response, err := c.statsClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get team standings from nba from url %s: %w", u, err)
	}
	defer response.Body.Close()

	if response.Header.Get("Content-Encoding") == "gzip" {
		response.Body, err = gzip.NewReader(response.Body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to create gzip reader when getting nba team standings: %w", err)
		}
	}

	teamsResult, err := unmarshalNBAHttpResponseToJSON[statsBaseResponse](response.Body)
	if err != nil {
		respBody, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("failed to unmarshal json for team standings from url %s: %s: %w", u, respBody, err)
	}

	var teamStandings []TeamStanding
	for _, resultSet := range teamsResult.ResultSets {
		headersMap := make(map[string]int, len(resultSet.Headers))
		for i, header := range resultSet.Headers {
			headersMap[header] = i
		}

		switch resultSet.Name {
		case "Standings":
			for _, rowSet := range resultSet.RowSet {
				teamID, err := parseRowSetValue[int](headersMap, rowSet, "TeamID")
				if err != nil {
					return nil, fmt.Errorf("failed to parse teamID from nba stats TeamInfoCommon endpoint: %w", err)
				}

				conference, err := parseRowSetValue[string](headersMap, rowSet, "Conference")
				if err != nil {
					return nil, fmt.Errorf("failed to parse conference from nba stats TeamInfoCommon endpoint: %w", err)
				}

				division, err := parseRowSetValue[string](headersMap, rowSet, "Division")
				if err != nil {
					return nil, fmt.Errorf("failed to parse division from nba stats TeamInfoCommon endpoint: %w", err)
				}

				teamStandings = append(teamStandings, TeamStanding{
					LeagueID:        leagueID,
					SeasonStartYear: seasonStartYear,
					SeasonType:      seasonType,
					TeamID:          teamID,
					Conference:      conference,
					Division:        division,
				})
			}

			break
		}
	}

	return teamStandings, nil
}
