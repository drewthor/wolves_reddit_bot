package nba

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const NBADailyAPIPath = "/prod/v1/today.json"

type DailyAPI struct {
	APIPaths          DailyAPIPaths `json:"links"`
	APISeasonInfoNode struct {
		SeasonStage       SeasonStage `json:"seasonStage"`
		SeasonYear        int         `json:"seasonYear"` // e.g. season start year so 2020 for 2020-2021
		RosterYear        int         `json:"rosterYear"` // e.g. season start year so 2020 for 2020-2021
		StatsStage        SeasonStage `json:"statsStage"`
		StatsYear         int         `json:"statsYear"`      // e.g. season start year so 2020 for 2020-2021
		DisplayYear       string      `json:"displayYear"`    // e.g. "2020-2021
		LastPlayByPlayURL string      `json:"lastPlayByPlay"` // e.g. "/json/cms/noseason/game/{{gameDate}}/{{gameId}}/pbp_last.json"
		AllPlayByPlayURL  string      `json:"allPlayByPlay"`  // e.g. "/data/10s/json/cms/noseason/game/{{gameDate}}/{{gameId}}/pbp_all.json"
		PlayerMatchupURL  string      `json:"playerMatchup"`  // e.g. "/data/10s/json/cms/2020/game/{{gameDate}}/{{gameId}}/playersPerGame.json"
		TeamSeriesURL     string      `json:"series"`         // e.g. "/data/5s/json/cms/2020/regseason/team/{{teamUrlCode}}/series.json"
	} `json:"teamSitesOnly"`
}

type DailyAPIPaths struct {
	Boxscore                  string `json:"boxscore"`                 // e.g. "/prod/v1/{{gameDate}}/{{gameId}}_boxscore.json"
	CurrentDate               string `json:"currentDate"`              // e.g. "20210704"
	Players                   string `json:"leagueRosterPlayers"`      // e.g. "/prod/v1/2020/players.json"
	Scoreboard                string `json:"scoreboard"`               // e.g. "/prod/v2/{{gameDate}}/scoreboard.json"
	Teams                     string `json:"teams"`                    // e.g. "/prod/v2/2020/teams.json"
	TeamSchedule              string `json:"teamSchedule"`             // e.g. "/prod/v1/2020/teams/{{teamUrlCode}}/schedule.json"
	Coaches                   string `json:"leagueRosterCoaches"`      // e.g. "/prod/v1/2020/coaches.json"
	TeamHomeICalendarDownload string `json:"teamICS"`                  // e.g. "/prod/teams/schedules/2020/{{teamUrlCode}}_home_schedule.ics"
	TeamAllICalendarDownload  string `json:"teamICS2"`                 // e.g. "/prod/teams/schedules/2020/{{teamUrlCode}}_schedule.ics"
	PeriodPlayByPlay          string `json:"pbp"`                      // e.g. "/prod/v1/{{gameDate}}/{{gameId}}_pbp_{{periodNum}}.json"
	PlayerGameLog             string `json:"playerGameLog"`            // e.g. "/prod/v1/2020/players/{{personId}}_gamelog.json"
	PlayerProfile             string `json:"playerProfile"`            // e.g. "/prod/v1/2020/players/{{personId}}_profile.json" - cumulative player stats by year
	LeagueConferenceStandings string `json:"leagueConfStandings"`      // e.g. "/prod/v1/current/standings_conference.json"
	LeagueDivisionStandings   string `json:"leagueDivStandings"`       // e.g. "/prod/v1/current/standings_division.json"
	LeagueStandings           string `json:"leagueUngroupedStandings"` // e.g. "/prod/v1/current/standings_all.json"
}

func GetDailyAPIPaths() DailyAPIPaths {
	url := nbaAPIBaseURI + NBADailyAPIPath
	response, httpErr := http.Get(url)

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if httpErr != nil {
		log.Fatal(httpErr)
	}

	dailyAPIResult := DailyAPI{}
	decodeErr := json.NewDecoder(response.Body).Decode(&dailyAPIResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	if dailyAPIResult.APIPaths.CurrentDate == "" || dailyAPIResult.APIPaths.Teams == "" || dailyAPIResult.APIPaths.TeamSchedule == "" || dailyAPIResult.APIPaths.Scoreboard == "" {
		log.Fatal("Could not get daily API paths")
	}
	return dailyAPIResult.APIPaths
}
