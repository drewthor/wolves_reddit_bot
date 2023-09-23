package nba

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
)

// newer league schedule but not sure if you can find by year https://cdn.nba.com/static/json/staticData/scheduleLeagueV2.json
const leagueScheduleURL = "https://cdn.nba.com/static/json/staticData/scheduleLeagueV2_1.json"

type LeagueSchedule struct {
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
			Games    []Game `json:"games"`
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

func GetLeagueSchedule(ctx context.Context, r2Client cloudflare.Client, bucket string, seasonStartYear int) (LeagueSchedule, error) {
	t := time.Now().UTC().Round(time.Hour).Format(time.RFC3339)
	filename := fmt.Sprintf(os.Getenv("STORAGE_PATH")+"/schedule/%d/%s_v2v1.json", seasonStartYear, t)

	objectKey := fmt.Sprintf("schedule/%d/%s_cdn.json", seasonStartYear, t)

	leagueSchedule, err := fetchObjectAndSaveToFile[LeagueSchedule](ctx, r2Client, leagueScheduleURL, filename, bucket, objectKey)
	if err != nil {
		return LeagueSchedule{}, err
	}

	return leagueSchedule, nil
}
