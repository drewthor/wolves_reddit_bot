package nba

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
)

// newer league schedule but not sure if you can find by year https://cdn.nba.com/static/json/staticData/scheduleLeagueV2.json
const leagueScheduleURL = "https://cdn.nba.com/static/json/staticData/scheduleLeagueV2_1.json"

type LeagueSchedule struct {
	LeagueNode struct {
		Games []Game `json:"standard"`
	} `json:"league"`
}

type LeagueScheduleCDN struct {
	Meta struct {
		Version int       `json:"version"`
		Request string    `json:"request"`
		Time    time.Time `json:"time"`
	} `json:"meta"`
	LeagueSchedule struct {
		SeasonYear string `json:"seasonYear"` // ex. 2022-23
		LeagueID   string `json:"leagueId"`
		GameDates  []struct {
			GameDate string `json:"gameDate"`
			Games    []struct {
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
			} `json:"games"`
		} `json:"gameDates"`
		Weeks []struct {
			WeekNumber int       `json:"weekNumber"`
			WeekName   string    `json:"weekName"`
			StartDate  time.Time `json:"startDate"`
			EndDate    time.Time `json:"endDate"`
		} `json:"weeks"`
		BroadcasterList []struct {
			BroadcasterID           int    `json:"broadcasterId"`
			BroadcasterDisplay      string `json:"broadcasterDisplay"`
			BroadcasterAbbreviation string `json:"broadcasterAbbreviation"`
			RegionID                int    `json:"regionId"`
		} `json:"broadcasterList"`
	} `json:"leagueSchedule"`
}

func GetLeagueSchedule(ctx context.Context, r2Client cloudflare.Client, bucket string) (LeagueScheduleCDN, error) {
	t := time.Now().UTC().Round(time.Hour).Format(time.RFC3339)
	filename := fmt.Sprintf(os.Getenv("STORAGE_PATH")+"/schedule/2022/%s_v2v1.json", t)

	objectKey := fmt.Sprintf("schedule/2022/%s_cdn", t)

	leagueSchedule, err := fetchObjectAndSaveToFile[LeagueScheduleCDN](ctx, r2Client, leagueScheduleURL, filename, bucket, objectKey)
	if err != nil {
		return LeagueScheduleCDN{}, err
	}

	return leagueSchedule, nil
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
