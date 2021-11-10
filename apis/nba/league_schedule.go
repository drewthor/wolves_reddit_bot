package nba

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if err != nil {
		log.Println(err)
		return nil, err
	}

	leagueScheduleResult := LeagueSchedule{}
	err = json.NewDecoder(response.Body).Decode(&leagueScheduleResult)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return leagueScheduleResult.LeagueNode.Games, nil
}

func GetSeasonLeagueSchedule(seasonStartYear int) ([]Game, error) {
	url := nbaAPIBaseURI + fmt.Sprintf("/prod/v1/%d/schedule.json", seasonStartYear)
	response, err := http.Get(url)

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if err != nil {
		log.Println(err)
		return nil, err
	}

	leagueScheduleResult := LeagueSchedule{}
	err = json.NewDecoder(response.Body).Decode(&leagueScheduleResult)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return leagueScheduleResult.LeagueNode.Games, nil

}
