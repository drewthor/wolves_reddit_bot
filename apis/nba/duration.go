package nba

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// This duration has come back from the NBA as MM:SS, "", and ISO8601 duration PT03M39.00S, PT240M, PT240M00.00S
type duration struct {
	DurationTenthSeconds int
	boxscoreRawValue     string
}

func (bgc *duration) UnmarshalJSON(data []byte) error {
	dataStr := string(data)
	errorStr := fmt.Sprintf("could not unmarshal nba boxscore game clock: %s to json", dataStr)
	unmarshalError := fmt.Errorf(errorStr)

	err := json.Unmarshal(data, &bgc.boxscoreRawValue)
	if err != nil {
		log.Println(errorStr)
		return unmarshalError
	}

	// handle empty duration ""
	if bgc.boxscoreRawValue == "" {
		bgc.DurationTenthSeconds = 0
		return nil
	}

	// remove extra whitespace
	rawStrFiltered := strings.ReplaceAll(bgc.boxscoreRawValue, " ", "")
	rawStringLower := strings.ToLower(rawStrFiltered)

	minutes := 0
	seconds := 0
	tenthSeconds := 0

	// handle 11:15 case
	if !strings.Contains(rawStringLower, "pt") {
		strs := strings.Split(rawStringLower, ":")
		if len(strs) != 2 {
			return unmarshalError
		}

		minutes, err = strconv.Atoi(strs[0])
		if err != nil {
			return unmarshalError
		}
		seconds, err = strconv.Atoi(strs[1])
		if err != nil {
			return unmarshalError
		}
	}

	// handle ISO-8601 duration
	if strings.Contains(rawStringLower, "pt") {
		durationStr := strings.TrimPrefix(rawStringLower, "pt")

		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return unmarshalError
		}

		tenthSeconds = int(duration.Milliseconds() / 100)
	}

	bgc.DurationTenthSeconds = (minutes * 60 * 100) + (seconds * 100) + tenthSeconds

	return nil
}
