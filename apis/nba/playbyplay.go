package nba

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"
)

const playByPlayURL = "https://cdn.nba.com/static/json/liveData/playbyplay/playbyplay_%s.json" // %s is the gameID

func (c client) PlayByPlayForGame(ctx context.Context, gameID string, outputWriters ...OutputWriter) (PlayByPlay, error) {
	response, err := c.client.Get(fmt.Sprintf(playByPlayURL, gameID))
	if err != nil {
		return PlayByPlay{}, fmt.Errorf("failed to make call to get playbyplay: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return PlayByPlay{}, fmt.Errorf("failed to get PlayByPlay object from url with status code: %d", response.StatusCode)
	}

	var respBody []byte
	respBody, err = io.ReadAll(response.Body)

	pbp, err := unmarshalNBAHttpResponseToJSON[PlayByPlay](bytes.NewReader(respBody))
	if err != nil {
		return PlayByPlay{}, fmt.Errorf("failed to get PlayByPlay object")
	}

	for _, outputWriter := range outputWriters {
		if err := outputWriter.Put(ctx, respBody); err != nil {
			return PlayByPlay{}, fmt.Errorf("failed to write output for play by play for game: %w", err)
		}
	}

	return pbp, nil
}

type PlayByPlay struct {
	Meta struct {
		Version int    `json:"version"`
		Code    int    `json:"code"`
		Request string `json:"request"`
		Time    string `json:"time"`
	} `json:"meta"`
	Game struct {
		GameID  string `json:"gameId"`
		Actions []struct {
			ActionNumber             int       `json:"actionNumber"`
			Clock                    duration  `json:"clock"`
			TimeActual               time.Time `json:"timeActual"`
			Period                   int       `json:"period"`
			PeriodType               string    `json:"periodType"`
			ActionType               string    `json:"actionType"`
			SubType                  string    `json:"subType,omitempty"`
			Qualifiers               []string  `json:"qualifiers"` // ex. ["2ndchance"], ["pointsinthepaint"], ["pointsinthepaint", "2ndchance"], ["fromturnover"]
			PersonID                 int       `json:"personId"`
			X                        *float64  `json:"x"`
			Y                        *float64  `json:"y"`
			Possession               int       `json:"possession"`
			ScoreHome                string    `json:"scoreHome"`
			ScoreAway                string    `json:"scoreAway"`
			Edited                   time.Time `json:"edited"`
			OrderNumber              int       `json:"orderNumber"`
			XLegacy                  *int      `json:"xLegacy"`
			YLegacy                  *int      `json:"yLegacy"`
			IsFieldGoal              int       `json:"isFieldGoal"`
			Side                     *string   `json:"side"` // ex. "left", "right"
			Description              string    `json:"description,omitempty"`
			PersonIdsFilter          []int     `json:"personIdsFilter"`
			TeamID                   int       `json:"teamId,omitempty"`
			TeamTricode              string    `json:"teamTricode,omitempty"`
			Descriptor               string    `json:"descriptor,omitempty"`
			JumpBallRecoveredName    string    `json:"jumpBallRecoveredName,omitempty"`
			JumpBallRecoverdPersonID int       `json:"jumpBallRecoverdPersonId,omitempty"`
			PlayerName               string    `json:"playerName,omitempty"`
			PlayerNameI              string    `json:"playerNameI,omitempty"`
			JumpBallWonPlayerName    string    `json:"jumpBallWonPlayerName,omitempty"`
			JumpBallWonPersonID      int       `json:"jumpBallWonPersonId,omitempty"`
			JumpBallLostPlayerName   string    `json:"jumpBallLostPlayerName,omitempty"`
			JumpBallLostPersonID     int       `json:"jumpBallLostPersonId,omitempty"`
			ShotDistance             float64   `json:"shotDistance,omitempty"`
			ShotResult               string    `json:"shotResult,omitempty"`
			PointsTotal              int       `json:"pointsTotal,omitempty"`
			AssistPlayerNameInitial  string    `json:"assistPlayerNameInitial,omitempty"`
			AssistPersonID           int       `json:"assistPersonId,omitempty"`
			AssistTotal              int       `json:"assistTotal,omitempty"`
			OfficialID               int       `json:"officialId,omitempty"`
			ShotActionNumber         int       `json:"shotActionNumber,omitempty"`
			ReboundTotal             int       `json:"reboundTotal,omitempty"`
			ReboundDefensiveTotal    int       `json:"reboundDefensiveTotal,omitempty"`
			ReboundOffensiveTotal    int       `json:"reboundOffensiveTotal,omitempty"`
			FoulPersonalTotal        int       `json:"foulPersonalTotal,omitempty"`
			FoulTechnicalTotal       int       `json:"foulTechnicalTotal,omitempty"`
			FoulDrawnPlayerName      string    `json:"foulDrawnPlayerName,omitempty"`
			FoulDrawnPersonID        int       `json:"foulDrawnPersonId,omitempty"`
			TurnoverTotal            int       `json:"turnoverTotal,omitempty"`
			StealPlayerName          string    `json:"stealPlayerName,omitempty"`
			StealPersonID            int       `json:"stealPersonId,omitempty"`
			Value                    string    `json:"value,omitempty"`
			BlockPlayerName          string    `json:"blockPlayerName,omitempty"`
			BlockPersonID            int       `json:"blockPersonId,omitempty"`
		} `json:"actions"`
	} `json:"game"`
}
