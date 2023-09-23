package nba

import (
	"fmt"
	"reflect"
)

type endpointName string

const (
	endpointNameTeamInfoCommon  = "teaminfocommon"
	endpointNameBoxscoreSummary = "boxscoresummary"
)

type statsBaseResponse struct {
	Resource   string `json:"resource"`
	Parameters struct {
		LeagueID   string      `json:"LeagueID"`
		Season     interface{} `json:"Season"`
		SeasonType interface{} `json:"SeasonType"`
		TeamID     int         `json:"TeamID"`
	} `json:"parameters"`
	ResultSets []struct {
		Name    string          `json:"name"`
		Headers []string        `json:"headers"`
		RowSet  [][]interface{} `json:"rowSet"`
	} `json:"resultSets"`
}

func parseRowSetValue[T any](headersMap map[string]int, rowSet []interface{}, header string) (T, error) {
	headerValueRaw := rowSet[headersMap[header]]
	value, ok := headerValueRaw.(T)
	if !ok {
		return *(new(T)), fmt.Errorf("failed to parse %s from row set: expected %v and got %v", header, reflect.TypeOf(*(new(T))), reflect.TypeOf(headerValueRaw))
	}

	return value, nil
}
