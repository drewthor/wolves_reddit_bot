package nba

import (
//"encoding/json"
)

func GetNBAAPIDailyURL() string {
	return "http://data.nba.net/10s/prod/v1/today.json"
}

type Daily struct {
	CurrentDate string `json:"currentDate"`
	TeamsURL    string `json:"teams"`
}
