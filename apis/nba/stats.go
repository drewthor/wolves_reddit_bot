package nba

import (
	"encoding/json"
	"fmt"
)

type endpointName string

const (
	endpointNameTeamInfoCommon    endpointName = "teaminfocommon"
	endpointNameBoxscoreSummary   endpointName = "boxscoresummary"
	endpointNamePlayByPlayV3      endpointName = "playbyplayv3"
	endpointNameLeagueGameLog     endpointName = "leaguegamelog"
	endpointNameFranchiseHistory  endpointName = "franchisehistory"
	endpointNameLeagueStandingsV3 endpointName = "leaguestandingsv3"
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
		Name    string              `json:"name"`
		Headers []string            `json:"headers"`
		RowSet  [][]json.RawMessage `json:"rowSet"`
	} `json:"resultSets"`
}

func parseRowSetValue[T any](headersMap map[string]int, rowSet []json.RawMessage, header string) (T, error) {
	rowSetIndex, ok := headersMap[header]
	if !ok {
		return *(new(T)), fmt.Errorf("failed to find header in header map for row set")
	}
	headerValueRaw := rowSet[rowSetIndex]
	var t T
	if err := json.Unmarshal(headerValueRaw, &t); err != nil {
		return *(new(T)), fmt.Errorf("failed to unmarshal json from row set: %w", err)
	}

	return t, nil
}
