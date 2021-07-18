package nba

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

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
