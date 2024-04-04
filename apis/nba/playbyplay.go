package nba

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/otel"
)

const playByPlayURL = "https://cdn.nba.com/static/json/liveData/playbyplay/playbyplay_%s.json" // %s is the gameID
const playByPlayV3URL = "https://stats.nba.com/stats/playbyplayv3?StartPeriod=%d&EndPeriod=%d&GameID=%s"

func (c Client) PlayByPlayForGame(ctx context.Context, gameID string, outputWriters ...OutputWriter) (PlayByPlay, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(playByPlayURL, gameID), nil)
	if err != nil {
		return PlayByPlay{}, fmt.Errorf("failed to create request to get play by play for game: %w", err)
	}

	response, err := c.client.Do(req)
	if err != nil {
		return PlayByPlay{}, fmt.Errorf("failed to make call to get playbyplay: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusForbidden {
		return PlayByPlay{}, ErrNotFound
	}

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
			JumpBallWonPlayerName    *string   `json:"jumpBallWonPlayerName,omitempty"`
			JumpBallWonPersonID      *int      `json:"jumpBallWonPersonId,omitempty"`
			JumpBallLostPlayerName   *string   `json:"jumpBallLostPlayerName,omitempty"`
			JumpBallLostPersonID     *int      `json:"jumpBallLostPersonId,omitempty"`
			ShotDistance             float64   `json:"shotDistance,omitempty"`
			ShotResult               string    `json:"shotResult,omitempty"`
			PointsTotal              int       `json:"pointsTotal,omitempty"`
			AssistPlayerNameInitial  *string   `json:"assistPlayerNameInitial,omitempty"`
			AssistPersonID           *int      `json:"assistPersonId,omitempty"`
			AssistTotal              *int      `json:"assistTotal,omitempty"`
			OfficialID               int       `json:"officialId,omitempty"`
			ShotActionNumber         int       `json:"shotActionNumber,omitempty"`
			ReboundTotal             int       `json:"reboundTotal,omitempty"`
			ReboundDefensiveTotal    int       `json:"reboundDefensiveTotal,omitempty"`
			ReboundOffensiveTotal    int       `json:"reboundOffensiveTotal,omitempty"`
			FoulPersonalTotal        int       `json:"foulPersonalTotal,omitempty"`
			FoulTechnicalTotal       int       `json:"foulTechnicalTotal,omitempty"`
			FoulDrawnPlayerName      *string   `json:"foulDrawnPlayerName,omitempty"`
			FoulDrawnPersonID        *int      `json:"foulDrawnPersonId,omitempty"`
			TurnoverTotal            int       `json:"turnoverTotal,omitempty"`
			StealPlayerName          *string   `json:"stealPlayerName,omitempty"`
			StealPersonID            *int      `json:"stealPersonId,omitempty"`
			Value                    string    `json:"value,omitempty"`
			BlockPlayerName          *string   `json:"blockPlayerName,omitempty"`
			BlockPersonID            *int      `json:"blockPersonId,omitempty"`
		} `json:"actions"`
	} `json:"game"`
}

type PlayByPlayV3 struct {
	Meta struct {
		Version int       `json:"version"`
		Request string    `json:"request"`
		Time    time.Time `json:"time"`
	} `json:"meta"`
	Game struct {
		GameID         string `json:"gameId"`
		VideoAvailable int    `json:"videoAvailable"`
		Actions        []struct {
			ActionNumber   int    `json:"actionNumber"`
			Clock          string `json:"clock"`
			Period         int    `json:"period"`
			TeamID         int    `json:"teamId"`
			TeamTricode    string `json:"teamTricode"`
			PersonID       int    `json:"personId"`
			PlayerName     string `json:"playerName"`
			PlayerNameI    string `json:"playerNameI"`
			XLegacy        int    `json:"xLegacy"`
			YLegacy        int    `json:"yLegacy"`
			ShotDistance   int    `json:"shotDistance"`
			ShotResult     string `json:"shotResult"`
			IsFieldGoal    int    `json:"isFieldGoal"`
			ScoreHome      string `json:"scoreHome"`
			ScoreAway      string `json:"scoreAway"`
			PointsTotal    int    `json:"pointsTotal"`
			Location       string `json:"location"`
			Description    string `json:"description"`
			ActionType     string `json:"actionType"`
			SubType        string `json:"subType"`
			VideoAvailable int    `json:"videoAvailable"`
			ActionID       int    `json:"actionId"`
		} `json:"actions"`
	} `json:"game"`
}

func (c Client) PlayByPlayV3ForGame(ctx context.Context, gameID string, outputWriters ...OutputWriter) (PlayByPlayV3, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "nba.Client.PlayByPlayV3ForGame")
	defer span.End()

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(playByPlayV3URL, 0, 0, gameID), nil)
	if err != nil {
		return PlayByPlayV3{}, fmt.Errorf("failed to get play by play v3 for game: %w", err)
	}

	response, err := c.statsClient.Do(req)
	if err != nil {
		return PlayByPlayV3{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return PlayByPlayV3{}, fmt.Errorf("failed to get %s object", endpointNamePlayByPlayV3)
	}

	if response.Header.Get("Content-Encoding") == "gzip" {
		response.Body, err = gzip.NewReader(response.Body)
		if err != nil {
			return PlayByPlayV3{}, fmt.Errorf("failed to create gzip reader when getting nba %s: %w", endpointNamePlayByPlayV3, err)
		}
	}

	var respBody []byte
	respBody, err = io.ReadAll(response.Body)

	for _, outputWriter := range outputWriters {
		if err := outputWriter.Put(ctx, respBody); err != nil {
			return PlayByPlayV3{}, fmt.Errorf("failed to write output for %s for game: %w", endpointNamePlayByPlayV3, err)
		}
	}

	playByPlayResult, err := unmarshalNBAHttpResponseToJSON[PlayByPlayV3](bytes.NewReader(respBody))
	if err != nil {
		return PlayByPlayV3{}, fmt.Errorf("failed to get %s response: %w", endpointNamePlayByPlayV3, err)
	}

	return playByPlayResult, nil
}
