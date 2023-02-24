package nba

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const TeamsURL = "https://data.nba.net/prod/v2/%d/teams.json"

const TeamLogoUrl = "https://cdn.nba.com/logos/nba/%d/primary/L/logo.svg"

const (
	teamCommonInfoBaseURL = "https://stats.nba.com/stats/teaminfocommon?"
)

type TeamsResult struct {
	LeagueNode struct {
		NBA        []Team `json:"standard"`
		Vegas      []Team `json:"vegas,omitempty"`
		Sacramento []Team `json:"sacramento,omitempty"`
		Utah       []Team `json:"utah,omitempty"`
	} `json:"league"`
}

type Team struct {
	IsNBAFranchise bool    `json:"isNBAFranchise"`
	ID             string  `json:"teamId"`
	TriCode        TriCode `json:"tricode"`
	FullName       string  `json:"fullName"`
	ShortName      string  `json:"teamShortName"`
	Nickname       string  `json:"nickname"`
	City           string  `json:"city"`
	AlternateCity  string  `json:"altCityName"`
	UrlName        string  `json:"urlName"`
	Conference     string  `json:"confName"`
	Division       string  `json:"divName"`
	AllStar        bool    `json:"isAllStar"`
}

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

func GetTeamsForSeason(seasonStartYear int) ([]Team, error) {
	url := fmt.Sprintf(TeamsURL, seasonStartYear)
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get current teams from nba from url %s", url)
	}
	defer response.Body.Close()

	teamsResult, err := unmarshalNBAHttpResponseToJSON[TeamsResult](response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get current teams from nba from url %s", url)
	}
	teams := []Team{}
	for _, team := range teamsResult.LeagueNode.NBA {
		teams = append(teams, team)
	}

	return teams, nil
}

func GetTeams(teamsAPIPath string) ([]Team, error) {
	url := nbaAPIBaseURI + teamsAPIPath
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get current teams from nba from url %s", url)
	}
	defer response.Body.Close()

	teamsResult, err := unmarshalNBAHttpResponseToJSON[TeamsResult](response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get current teams from nba from url %s", url)
	}
	teams := []Team{}
	for _, team := range teamsResult.LeagueNode.NBA {
		teams = append(teams, team)
	}

	return teams, nil
}

func (c client) GetCommonTeamInfo(ctx context.Context, leagueID string, teamID int) (TeamCommonInfo, error) {
	urlValues := url.Values{
		"LeagueID": {leagueID},
		"TeamID":   {strconv.Itoa(teamID)},
	}
	teamURL := teamCommonInfoBaseURL + urlValues.Encode()
	response, err := http.Get(teamURL)
	if err != nil {
		return TeamCommonInfo{}, fmt.Errorf("failed to get current teams from nba from url %s", teamURL)
	}
	defer response.Body.Close()

	teamsResult, err := unmarshalNBAHttpResponseToJSON[statsBaseResponse](response.Body)
	if err != nil {
		return TeamCommonInfo{}, fmt.Errorf("failed to get current teams from nba from url %s", teamURL)
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
				return TeamCommonInfo{}, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endPointNameTeamInfoCommon, err)
			}

			seasonYear, err := parseRowSetValue[string](headersMap, rowSet, "SEASON_YEAR")
			if err != nil {
				return TeamCommonInfo{}, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endPointNameTeamInfoCommon, err)
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
