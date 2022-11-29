package nba

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var NBACurrentSeasonStartYear = -1

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
	Boxscore                  string      `json:"boxscore"`                 // e.g. "/prod/v1/{{gameDate}}/{{gameId}}_boxscore.json"
	CurrentDate               currentDate `json:"currentDate"`              // e.g. "20210704"
	Players                   string      `json:"leagueRosterPlayers"`      // e.g. "/prod/v1/2020/players.json"
	Scoreboard                string      `json:"scoreboard"`               // e.g. "/prod/v2/{{gameDate}}/scoreboard.json"
	Teams                     string      `json:"teams"`                    // e.g. "/prod/v2/2020/teams.json"
	TeamSchedule              string      `json:"teamSchedule"`             // e.g. "/prod/v1/2020/teams/{{teamUrlCode}}/schedule.json"
	LeagueSchedule            string      `json:"leagueSchedule"`           // e.g. "/prod/v1/2020/schedule.json"
	Coaches                   string      `json:"leagueRosterCoaches"`      // e.g. "/prod/v1/2020/coaches.json"
	TeamHomeICalendarDownload string      `json:"teamICS"`                  // e.g. "/prod/teams/schedules/2020/{{teamUrlCode}}_home_schedule.ics"
	TeamAllICalendarDownload  string      `json:"teamICS2"`                 // e.g. "/prod/teams/schedules/2020/{{teamUrlCode}}_schedule.ics"
	PeriodPlayByPlay          string      `json:"pbp"`                      // e.g. "/prod/v1/{{gameDate}}/{{gameId}}_pbp_{{periodNum}}.json"
	PlayerGameLog             string      `json:"playerGameLog"`            // e.g. "/prod/v1/2020/players/{{personId}}_gamelog.json"
	PlayerProfile             string      `json:"playerProfile"`            // e.g. "/prod/v1/2020/players/{{personId}}_profile.json" - cumulative player stats by year
	LeagueConferenceStandings string      `json:"leagueConfStandings"`      // e.g. "/prod/v1/current/standings_conference.json"
	LeagueDivisionStandings   string      `json:"leagueDivStandings"`       // e.g. "/prod/v1/current/standings_division.json"
	LeagueStandings           string      `json:"leagueUngroupedStandings"` // e.g. "/prod/v1/current/standings_all.json"
}

type currentDate struct {
	Time time.Time
}

func (c *currentDate) UnmarshalJSON(data []byte) error {
	raw := ""
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&raw)
	if err != nil {
		return fmt.Errorf("could not unmarshal nba current date: %s to time.Time error: %w", string(data), err)
	}
	c.Time, err = time.Parse(TimeDayFormat, raw)
	if err != nil {
		return fmt.Errorf("could not parse nba current date %s using format %s error: %w", string(data), TimeDayFormat, err)
	}

	return nil
}

func GetDailyAPIPaths() (DailyAPI, error) {
	url := nbaAPIBaseURI + NBADailyAPIPath
	response, err := http.Get(url)
	if err != nil {
		return DailyAPI{}, fmt.Errorf("failed to get nba daily api paths from url %s %w", url, err)
	}
	defer response.Body.Close()

	dailyAPIResult, err := unmarshalNBAHttpResponseToJSON[DailyAPI](response.Body)
	if err != nil {
		return DailyAPI{}, err
	}
	if dailyAPIResult.APIPaths.CurrentDate.Time.IsZero() || dailyAPIResult.APIPaths.Teams == "" || dailyAPIResult.APIPaths.TeamSchedule == "" || dailyAPIResult.APIPaths.Scoreboard == "" {
		return DailyAPI{}, fmt.Errorf("could not get nba daily nba API paths, empty json from url %s", url)
	}
	return dailyAPIResult, nil
}

func SetCurrentSeasonStartYear(startYear int) {
	NBACurrentSeasonStartYear = startYear
}

func init() {
	dailyAPIPaths, err := GetDailyAPIPaths()
	if err != nil {
		currentTime := time.Now()
		NBACurrentSeasonStartYear = currentTime.Year()
		if currentTime.Month() < time.July {
			NBACurrentSeasonStartYear = currentTime.Year() - 1
		}
		return
	}
	NBACurrentSeasonStartYear = dailyAPIPaths.APISeasonInfoNode.SeasonYear
}
