package nba

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type datetime struct {
	Time time.Time
}

func (d *datetime) UnmarshalJSON(data []byte) error {
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&d.Time)
	if err != nil {
		// for some dumb reason the nba api sometimes has fields like utc times that get returned in either rfc3339 format
		// or in the format yyyy-mm-dd
		raw := ""
		err = json.NewDecoder(bytes.NewReader(data)).Decode(&raw)
		if err != nil {
			return fmt.Errorf("could not unmarshal nba time: %s to time.Time error: %w", string(data), err)
		}
		d.Time, err = time.Parse("2006-01-02", raw)
		if err != nil {
			return fmt.Errorf("could not unmarshal nba time: %s to time.Time error: %w", raw, err)
		}
	}

	return nil
}
