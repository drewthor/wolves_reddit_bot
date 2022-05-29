package nba

import (
	"fmt"
	"net/http"
)

// newer league schedule but not sure if you can find by year https://cdn.nba.com/static/json/staticData/scheduleLeagueV2.json

type LeagueSchedule struct {
	LeagueNode struct {
		Games []Game `json:"standard"`
	} `json:"league"`
}

func GetCurrentLeagueSchedule(leageSchedulePath string) ([]Game, error) {
	url := nbaAPIBaseURI + leageSchedulePath
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get current league schedule from nba from %s %w", url, err)
	}
	defer response.Body.Close()

	leagueScheduleResult, err := unmarshalNBAHttpResponseToJSON[LeagueSchedule](response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get current league schedule from nba from %s %w", url, err)
	}

	return leagueScheduleResult.LeagueNode.Games, nil
}

func GetSeasonLeagueSchedule(seasonStartYear int) ([]Game, error) {
	url := nbaAPIBaseURI + fmt.Sprintf("/prod/v1/%d/schedule.json", seasonStartYear)
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get league schedule from nba for year %d from %s %w", seasonStartYear, url, err)
	}

	leagueScheduleResult, err := unmarshalNBAHttpResponseToJSON[LeagueSchedule](response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get league schedule from nba for year %d from %s %w", seasonStartYear, url, err)
	}

	return leagueScheduleResult.LeagueNode.Games, nil

}
