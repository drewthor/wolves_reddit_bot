package nba

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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

func (c Client) CurrentLeagueSchedule(ctx context.Context, objectKey string) (LeagueSchedule, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "nba.Client.LeagueSchedule")
	defer span.End()

	req, err := retryablehttp.NewRequest(http.MethodGet, leagueScheduleURL, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return LeagueSchedule{}, fmt.Errorf("failed to create request to get league schedule: %w", err)
	}

	response, err := c.client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return LeagueSchedule{}, fmt.Errorf("failed to get LeagueSchedule object: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = fmt.Errorf("failed to successfully get LeagueSchedule object")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if slices.Contains([]int{http.StatusNotFound, http.StatusForbidden}, response.StatusCode) {
			return LeagueSchedule{}, ErrNotFound
		}
		return LeagueSchedule{}, err
	}

	var respBody []byte
	respBody, err = io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
	}

	if c.Cache != nil {
		if err := c.Cache.PutObject(ctx, objectKey, bytes.NewReader(respBody)); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return LeagueSchedule{}, fmt.Errorf("failed to cache league schedule object: %w", err)
		}
	}

	var leagueSchedule LeagueSchedule
	if err := json.Unmarshal(respBody, &leagueSchedule); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return LeagueSchedule{}, fmt.Errorf("failed to unmarshal league schedule json: %w", err)
	}

	return leagueSchedule, nil
}
