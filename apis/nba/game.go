package nba

import (
	"time"
)

type Game struct {
	GameID           string    `json:"gameId"`
	GameCode         string    `json:"gameCode"`
	GameStatus       int       `json:"gameStatus"`
	GameStatusText   string    `json:"gameStatusText"`
	GameSequence     int       `json:"gameSequence"`
	GameDateEst      time.Time `json:"gameDateEst"`
	GameTimeEst      time.Time `json:"gameTimeEst"`
	GameDateTimeEst  time.Time `json:"gameDateTimeEst"`
	GameDateUTC      time.Time `json:"gameDateUTC"`
	GameTimeUTC      time.Time `json:"gameTimeUTC"`
	GameDateTimeUTC  time.Time `json:"gameDateTimeUTC"`
	AwayTeamTime     time.Time `json:"awayTeamTime"`
	HomeTeamTime     time.Time `json:"homeTeamTime"`
	Day              string    `json:"day"`
	MonthNum         int       `json:"monthNum"`
	WeekNumber       int       `json:"weekNumber"`
	WeekName         string    `json:"weekName"`
	IfNecessary      bool      `json:"ifNecessary"`
	SeriesGameNumber string    `json:"seriesGameNumber"`
	SeriesText       string    `json:"seriesText"`
	ArenaName        string    `json:"arenaName"`
	ArenaState       string    `json:"arenaState"`
	ArenaCity        string    `json:"arenaCity"`
	PostponedStatus  string    `json:"postponedStatus"`
	BranchLink       string    `json:"branchLink"`
	Broadcasters     struct {
		NationalTvBroadcasters []struct {
			BroadcasterScope        string `json:"broadcasterScope"`
			BroadcasterMedia        string `json:"broadcasterMedia"`
			BroadcasterID           int    `json:"broadcasterId"`
			BroadcasterDisplay      string `json:"broadcasterDisplay"`
			BroadcasterAbbreviation string `json:"broadcasterAbbreviation"`
			TapeDelayComments       string `json:"tapeDelayComments"`
			RegionID                int    `json:"regionId"`
		} `json:"nationalTvBroadcasters"`
		NationalRadioBroadcasters []interface{} `json:"nationalRadioBroadcasters"`
		HomeTvBroadcasters        []struct {
			BroadcasterScope        string `json:"broadcasterScope"`
			BroadcasterMedia        string `json:"broadcasterMedia"`
			BroadcasterID           int    `json:"broadcasterId"`
			BroadcasterDisplay      string `json:"broadcasterDisplay"`
			BroadcasterAbbreviation string `json:"broadcasterAbbreviation"`
			TapeDelayComments       string `json:"tapeDelayComments"`
			RegionID                int    `json:"regionId"`
		} `json:"homeTvBroadcasters"`
		HomeRadioBroadcasters []struct {
			BroadcasterScope        string `json:"broadcasterScope"`
			BroadcasterMedia        string `json:"broadcasterMedia"`
			BroadcasterID           int    `json:"broadcasterId"`
			BroadcasterDisplay      string `json:"broadcasterDisplay"`
			BroadcasterAbbreviation string `json:"broadcasterAbbreviation"`
			TapeDelayComments       string `json:"tapeDelayComments"`
			RegionID                int    `json:"regionId"`
		} `json:"homeRadioBroadcasters"`
		AwayTvBroadcasters    []interface{} `json:"awayTvBroadcasters"`
		AwayRadioBroadcasters []struct {
			BroadcasterScope        string `json:"broadcasterScope"`
			BroadcasterMedia        string `json:"broadcasterMedia"`
			BroadcasterID           int    `json:"broadcasterId"`
			BroadcasterDisplay      string `json:"broadcasterDisplay"`
			BroadcasterAbbreviation string `json:"broadcasterAbbreviation"`
			TapeDelayComments       string `json:"tapeDelayComments"`
			RegionID                int    `json:"regionId"`
		} `json:"awayRadioBroadcasters"`
		IntlRadioBroadcasters []interface{} `json:"intlRadioBroadcasters"`
		IntlTvBroadcasters    []interface{} `json:"intlTvBroadcasters"`
	} `json:"broadcasters"`
	HomeTeam struct {
		TeamID      int    `json:"teamId"`
		TeamName    string `json:"teamName"`
		TeamCity    string `json:"teamCity"`
		TeamTricode string `json:"teamTricode"`
		TeamSlug    string `json:"teamSlug"`
		Wins        int    `json:"wins"`
		Losses      int    `json:"losses"`
		Score       int    `json:"score"`
		Seed        int    `json:"seed"`
	} `json:"homeTeam"`
	AwayTeam struct {
		TeamID      int    `json:"teamId"`
		TeamName    string `json:"teamName"`
		TeamCity    string `json:"teamCity"`
		TeamTricode string `json:"teamTricode"`
		TeamSlug    string `json:"teamSlug"`
		Wins        int    `json:"wins"`
		Losses      int    `json:"losses"`
		Score       int    `json:"score"`
		Seed        int    `json:"seed"`
	} `json:"awayTeam"`
	PointsLeaders []struct {
		PersonID    int     `json:"personId"`
		FirstName   string  `json:"firstName"`
		LastName    string  `json:"lastName"`
		TeamID      int     `json:"teamId"`
		TeamCity    string  `json:"teamCity"`
		TeamName    string  `json:"teamName"`
		TeamTricode string  `json:"teamTricode"`
		Points      float64 `json:"points"`
	} `json:"pointsLeaders"`
}
