package nba

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
)

// play by play https://cdn.nba.com/static/json/liveData/playbyplay/playbyplay_0022000180.json

const BoxscoreURL = "https://cdn.nba.com/static/json/liveData/boxscore/boxscore_%s.json"

type Boxscore struct {
	GameNode struct {
		GameID               string             `json:"gameId"`
		GameTimeLocal        string             `json:"gameTimeLocal"` // ex. 2021-07-14T20:00:00-05:00
		GameTimeUTC          datetime           `json:"gameTimeUTC"`
		GameTimeHome         string             `json:"gameTimeHome"`      // ex. 2021-07-14T20:00:00-05:00
		GameTimeAway         string             `json:"gameTimeAway"`      // ex. 2021-07-14T20:00:00-07:00
		GameET               string             `json:"gameEt"`            // ex. 2021-07-14T20:00:00-04:00
		TotalDurationMinutes int                `json:"duration"`          // duration in minutes (real world time) from tipoff to final buzzer
		GameCode             string             `json:"gameCode"`          // ex. 20210714/PHXMIL
		GameStatusText       string             `json:"gameStatusText"`    // ex. [Q3 03:03, Final]
		GameStatus           int                `json:"gameStatus"`        // ex. [1 - scheduled, 2 - in progress, 3 - final]
		RegulationPeriods    int                `json:"regulationPeriods"` // not sure why this would be anything but 4?
		Period               int                `json:"period"`
		GameClock            duration           `json:"gameClock"` // ex. PT11M34.00S
		Attendance           int                `json:"attendance"`
		Sellout              string             `json:"sellout"` // ex. [0,1]
		Arena                BoxscoreArena      `json:"arena"`
		Officials            []BoxscoreOfficial `json:"officials"`
		HomeTeam             BoxscoreTeam       `json:"homeTeam"`
		AwayTeam             BoxscoreTeam       `json:"awayTeam"`
	} `json:"game"`
}

type BoxscoreGameClockMinutes struct {
	Duration         int
	boxscoreRawValue string
}

func (bgc *BoxscoreGameClockMinutes) UnmarshalJSON(data []byte) error {
	dataStr := string(data)
	errorStr := fmt.Sprintf("could not unmarshal nba boxscore game clock: %s to json", dataStr)
	unmarshalError := fmt.Errorf(errorStr)

	err := json.Unmarshal(data, &bgc.boxscoreRawValue)
	if err != nil {
		return unmarshalError
	}

	if bgc.boxscoreRawValue == "" {
		bgc.Duration = 0
		return nil
	}

	minutesSecondsStr := strings.TrimPrefix(bgc.boxscoreRawValue, "PT")

	minutesStr := strings.TrimSuffix(minutesSecondsStr, "M")

	minutes, err := strconv.Atoi(minutesStr)
	if err != nil {
		return unmarshalError
	}

	bgc.boxscoreRawValue = dataStr
	bgc.Duration = minutes

	return nil
}

type BoxscoreArena struct {
	ID       int     `json:"arenaId"`
	Name     string  `json:"arenaName"`
	City     *string `json:"arenaCity"`
	State    *string `json:"arenaState"`    // ex. MN
	Country  string  `json:"arenaCountry"`  // ex. US
	Timezone string  `json:"arenaTimezone"` // ex. America/Chicago
}

type BoxscoreOfficial struct {
	PersonID     int    `json:"personId"`
	Name         string `json:"name"`  // ex. Tony Brothers
	NameI        string `json:"nameI"` // ex. T. Brothers
	FirstName    string `json:"firstName"`
	LastName     string `json:"familyName"`
	JerseyNumber string `json:"jerseyNum"`
	Assignment   string `json:"assignment"` // ex. [OFFICIAL1, OFFICIAL2, OFFICIAL3]
}

type BoxscoreTeam struct {
	ID                int                    `json:"teamId"`
	Name              string                 `json:"teamName"`
	City              string                 `json:"teamCity"`
	Tricode           string                 `json:"teamTricode"`
	Points            int                    `json:"score"`
	InBonus           string                 `json:"inBonus"` // ex. [0,1]
	TimeoutsRemaining int                    `json:"timeoutsRemaining"`
	Periods           []BoxscorePeriod       `json:"periods"`
	Players           []BoxscorePlayer       `json:"players"`
	Statistics        BoxscoreTeamStatistics `json:"statistics"`
}

type BoxscorePeriod struct {
	Period     int    `json:"period"`
	PeriodType string `json:"periodType"` // ex. [REGULAR, OVERTIME?]
	Points     int    `json:"score"`
}

type BoxscorePlayer struct {
	Name                  string                   `json:"name"`  // ex. Jrue Holiday
	NameI                 string                   `json:"nameI"` // ex. J. Holiday
	FirstName             string                   `json:"firstName"`
	LastName              string                   `json:"familyName"`
	Status                string                   `json:"status"`                // ex. [ACTIVE, INACTIVE]
	NotPlayingReason      *string                  `json:"notPlayingReason"`      // ex. INACTIVE_INJURY only set if status is INACTIVE
	NotPlayingDescription *string                  `json:"notPlayingDescription"` // ex. Left Ankle; Surgery only set if status is INACTIVE
	Order                 int                      `json:"order"`                 // I believe this is the order in which they played e.g. 6 is the 6th man or 1st off the bench
	ID                    int                      `json:"personId"`
	JerseyNumber          string                   `json:"jerseyNum"`
	Position              *string                  `json:"position"` // ex. [PG, SG, SF, PF, C] only set for starters?
	Starter               string                   `json:"starter"`  // ex. [0,1]
	OnCourt               string                   `json:"oncourt"`  // ex. [0,1]
	Played                string                   `json:"played"`   // ex. [0,1]
	Statistics            BoxscorePlayerStatistics `json:"statistics"`
}

type BoxscorePlayerStatistics struct {
	Assists                 int                      `json:"assists"`
	Blocks                  int                      `json:"blocks"`
	BlocksReceived          int                      `json:"blocksReceived"` // times the player got blocked?
	FieldGoalsAttempted     int                      `json:"fieldGoalsAttempted"`
	FieldGoalsMade          int                      `json:"fieldGoalsMade"`
	FieldGoalsPercentage    float64                  `json:"fieldGoalsPercentage"` // ex. 0.444444444444444
	FoulsOffensive          int                      `json:"foulsOffensive"`
	FoulsDrawn              int                      `json:"foulsDrawn"`
	FoulsPersonal           int                      `json:"foulsPersonal"`
	FoulsTechnical          int                      `json:"foulsTechnical"`
	FreeThrowsAttempted     int                      `json:"freeThrowsAttempted"`
	FreeThrowsMade          int                      `json:"freeThrowsMade"`
	FreeThrowsPercentage    float64                  `json:"freeThrowsPercentage"` // ex. 1
	Minus                   float64                  `json:"minus"`                // boxscore minus
	Minutes                 duration                 `json:"minutes"`              // ex. PT34M43.00S
	MinutesCalculated       BoxscoreGameClockMinutes `json:"minutesCalculated"`    // ex. PT35M, PT00M
	Plus                    float64                  `json:"plus"`                 // boxscore plus
	PlusMinus               float64                  `json:"plusMinusPoints"`      // boxscore plus minus
	Points                  int                      `json:"points"`
	PointsFastBreak         int                      `json:"pointsFastBreak"`
	PointsInThePaint        int                      `json:"pointsInThePaint"`
	PointsSecondChance      int                      `json:"pointsSecondChance"`
	ReboundsDefensive       int                      `json:"reboundsDefensive"`
	ReboundsOffensive       int                      `json:"reboundsOffensive"`
	ReboundsTotal           int                      `json:"reboundsTotal"`
	Steals                  int                      `json:"steals"`
	ThreePointersAttempted  int                      `json:"threePointersAttempted"`
	ThreePointersMade       int                      `json:"threePointersMade"`
	ThreePointersPercentage float64                  `json:"threePointersPercentage"` // ex. 0.2
	Turnovers               int                      `json:"turnovers"`
	TwoPointersAttempted    int                      `json:"twoPointersAttempted"`
	TwoPointersMade         int                      `json:"twoPointersMade"`
	TwoPointersPercentage   float64                  `json:"twoPointersPercentage"`
}

type BoxscoreTeamStatistics struct {
	Assists                      int      `json:"assists"`
	AssistsToTurnoverRatio       float64  `json:"assistsTurnoverRatio"` // ex. 0.866666666666667
	BenchPoints                  int      `json:"benchPoints"`
	BiggestLead                  int      `json:"biggestLead"`
	BiggestLeadScore             string   `json:"biggestLeadScore"` // ex. 16-29
	BiggestScoringRun            int      `json:"biggestScoringRun"`
	BiggestScoringRunScore       string   `json:"biggestScoringRunScore"` // ex. 35-29
	Blocks                       int      `json:"blocks"`
	BlocksReceived               int      `json:"blocksReceived"` // times the team got blocked?
	FastBreakPointsAttempted     int      `json:"fastBreakPointsAttempted"`
	FastBreakPointsMade          int      `json:"fastBreakPointsMade"`
	FastBreakPointsPercentage    float64  `json:"fastBreakPointsPercentage"` // ex. 0.375
	FieldGoalsAttempted          int      `json:"fieldGoalsAttempted"`
	FieldGoalsEffectiveAdjusted  float64  `json:"fieldGoalsEffectiveAdjusted"` // ex. 0.41538461538461496
	FieldGoalsMade               int      `json:"fieldGoalsMade"`
	FieldGoalsPercentage         float64  `json:"fieldGoalsPercentage"` // ex. 0.444444444444444
	FoulsOffensive               int      `json:"foulsOffensive"`
	FoulsDrawn                   int      `json:"foulsDrawn"`
	FoulsPersonal                int      `json:"foulsPersonal"`
	FoulsTeam                    int      `json:"foulsTeam"`
	FoulsTechnical               int      `json:"foulsTechnical"`
	FoulsTeamTechnical           int      `json:"foulsTeamTechnical"`
	FreeThrowsAttempted          int      `json:"freeThrowsAttempted"`
	FreeThrowsMade               int      `json:"freeThrowsMade"`
	FreeThrowsPercentage         float64  `json:"freeThrowsPercentage"` // ex. 1
	LeadChanges                  int      `json:"leadChanges"`
	Minutes                      duration `json:"minutes"`           // ex. PT34M43.00S
	MinutesCalculated            duration `json:"minutesCalculated"` // ex. PT35M, PT00M
	Points                       int      `json:"points"`
	PointsAgainst                int      `json:"pointsAgainst"`
	PointsFastBreak              int      `json:"pointsFastBreak"`
	PointsOffTurnovers           int      `json:"pointsFromTurnovers"`
	PointsInThePaint             int      `json:"pointsInThePaint"`
	PointsInThePaintAttempted    int      `json:"pointsInThePaintAttempted"`
	PointsInThePaintMade         int      `json:"pointsInThePaintMade"`
	PointsInThePaintPercentage   float64  `json:"pointsInThePaintPercentage"` // ex. 0.5
	PointsSecondChance           int      `json:"pointsSecondChance"`
	PointsSecondChanceAttempted  int      `json:"secondChancePointsAttempted"`
	PointsSecondChanceMade       int      `json:"secondChancePointsMade"`
	PointsSecondChancePercentage float64  `json:"secondChancePointsPercentage"`
	ReboundsDefensive            int      `json:"reboundsDefensive"`
	ReboundsOffensive            int      `json:"reboundsOffensive"`
	ReboundsPersonal             int      `json:"reboundsPersonal"` // rebounds made by one player?
	ReboundsTeam                 int      `json:"reboundsTeam"`     // from nba.com No individual rebound is credited in situations where the whistle stops play before there is player possession following a shot attempt. Instead only a team rebound is credited to the team that gains possession following a stop in play. For example, if the ball goes out of bounds after the field goal attempt, a team rebound is awarded to the team in white since no player secured possession.
	ReboundsTeamDefensive        int      `json:"reboundsTeamDefensive"`
	ReboundsTeamOffensive        int      `json:"reboundsTeamOffensive"`
	ReboundsTotal                int      `json:"reboundsTotal"`
	Steals                       int      `json:"steals"`
	ThreePointersAttempted       int      `json:"threePointersAttempted"`
	ThreePointersMade            int      `json:"threePointersMade"`
	ThreePointersPercentage      float64  `json:"threePointersPercentage"` // ex. 0.2
	TimeLeading                  duration `json:"timeLeading"`             // ex. PT09M26.00S
	TimesTied                    int      `json:"timesTied"`
	TrueShootingAttempts         float64  `json:"trueShootingAttempts"`   // ex. 71.72
	TrueShootingPercentage       float64  `json:"trueShootingPercentage"` // ex. 0.53680981595092
	Turnovers                    int      `json:"turnovers"`
	TurnoversTeam                int      `json:"turnoversTeam"` // 5 second inbound violation, 24 second shot clock violation or others not attributable to a single player
	TurnoversTotal               int      `json:"turnoversTotal"`
	TwoPointersAttempted         int      `json:"twoPointersAttempted"`
	TwoPointersMade              int      `json:"twoPointersMade"`
	TwoPointersPercentage        float64  `json:"twoPointersPercentage"`
}
type TeamBoxscoreInfo struct {
	TeamID          string  `json:"teamId"`
	TriCode         TriCode `json:"triCode"`
	Wins            string  `json:"win"`
	Losses          string  `json:"loss"`
	SeriesWins      string  `json:"seriesWin"`
	SeriesLosses    string  `json:"seriesLoss"`
	Points          string  `json:"score"`
	PointsByQuarter []struct {
		Points string `json:"score"`
	} `json:"linescore"`
}

type TeamStats struct {
	BiggestLead        string          `json:"biggestLead"`
	LongestRun         string          `json:"longestRun"`
	PointsInPaint      string          `json:"pointsInPaint"`
	PointsOffTurnovers string          `json:"pointsOffTurnovers"`
	SecondChancePoints string          `json:"secondChancePoints"`
	TeamStatsTotals    TeamStatsTotals `json:"totals"`
}

type TeamStatsTotals struct {
	Points               string      `json:"points"`
	Minutes              duration    `json:"min"`
	FieldGoalsMade       string      `json:"fgm"`
	FieldGoalsAttempted  string      `json:"fga"`
	FieldGoalPercentage  string      `json:"fgp"`
	FreeThrowsMade       string      `json:"ftm"`
	FreeThrowsAttempted  string      `json:"fta"`
	FreeThrowPercentage  string      `json:"ftp"`
	ThreePointsMade      string      `json:"tpm"`
	ThreePointsAttempted string      `json:"tpa"`
	ThreePointPercentage string      `json:"tpp"`
	OffensiveRebounds    string      `json:"offReb"`
	DefensiveRebounds    string      `json:"defReb"`
	TotalRebounds        string      `json:"totReb"`
	Assists              string      `json:"assists"`
	PersonalFouls        string      `json:"pfouls"`
	Steals               string      `json:"steals"`
	Turnovers            string      `json:"turnovers"`
	Blocks               string      `json:"blocks"`
	PlusMinus            string      `json:"plusMinus"`
	TeamLeaders          TeamLeaders `json:"leaders"`
}

type TeamLeaders struct {
	PointsNode struct {
		Points      string `json:"value"`
		PlayersNode []struct {
			PlayerID string `json:"personId"`
		} `json:"players"`
	} `json:"points"`
	ReboundsNode struct {
		Rebounds    string `json:"value"`
		PlayersNode []struct {
			PlayerID string `json:"personId"`
		} `json:"players"`
	} `json:"points"`
	AssistsNode struct {
		Assists     string `json:"value"`
		AssistsNode []struct {
			PlayerID string `json:"personId"`
		} `json:"players"`
	} `json:"points"`

	teamID          int
	PointsLeaders   []string
	Points          int
	ReboundsLeaders []string
	Rebounds        int
	AssistsLeaders  []string
	Assists         int
	BlocksLeaders   []string
	Blocks          int
	StealsLeaders   []string
	Steals          int
}

type PlayerStats struct {
	ID                    string `json:"personId"`
	TeamID                string `json:"teamId"`
	Points                string `json:"points"`
	Minutes               string `json:"min"`
	FieldGoalsMade        string `json:"fgm"`
	FieldGoalsAttempted   string `json:"fga"`
	FieldGoalPercentage   string `json:"fgp"`
	FreeThrowsMade        string `json:"ftm"`
	FreeThrowsAttempted   string `json:"fta"`
	FreeThrowsPercentage  string `json:"ftp"`
	ThreePointsMade       string `json:"tpm"`
	ThreePointsAttempted  string `json:"tpa"`
	ThreePointsPercentage string `json:"tpp"`
	OffensiveRebounds     string `json:"offReb"`
	DefensiveRebounds     string `json:"defReb"`
	TotalRebounds         string `json:"totReb"`
	Assists               string `json:"assists"`
	PersonalFouls         string `json:"pfouls"`
	Steals                string `json:"steals"`
	Turnovers             string `json:"turnovers"`
	Blocks                string `json:"blocks"`
	PlusMinus             string `json:"plusMinus"`
	DidNotPlayStatus      string `json:"dnp"`
}

type ArenaInfo struct {
	Name    string `json:"name"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type RefereeInfo struct {
	FullName string `json:"firstNameLastName"`
}

type VideoBroadcasterInfo struct {
	ShortName string `json:"shortName"`
	LongName  string `json:"longName"`
}

type GameVideoBroadcastInfo struct {
	LeaguePass bool `json:"isLeaguePass"`
}

func (b Boxscore) Final() bool {
	return b.GameNode.GameStatus == 3
}

type BoxscoreSummary struct {
	GameDate                          time.Time
	GameInSequence                    int // ex. game in season series between opponents or playoff round game number
	GameID                            string
	GameStatusID                      int    // ex. [1 - scheduled, 2 - in progress, 3 - final]
	GameStatusText                    string // ex. Final
	GameCode                          string // ex. 20180207/MINCLE
	HomeTeam                          BoxscoreSummaryTeam
	AwayTeam                          BoxscoreSummaryTeam
	SeasonStartYear                   int
	Period                            int
	PCTime                            string // no idea what this is
	NationalTVBroadcasterAbbreviation string // ex. ESPN
	PeriodTimeBroadcast               string // ex. Q5  - ESPN
	WHStatus                          int    // no idea what this is
	Attendance                        int
	GameDurationSeconds               time.Duration
	//"GAME_DATE_EST",
	//"GAME_SEQUENCE",
	//"GAME_ID",
	//"GAME_STATUS_ID",
	//"GAME_STATUS_TEXT",
	//"GAMECODE",
	//"HOME_TEAM_ID",
	//"VISITOR_TEAM_ID",
	//"SEASON",
	//"LIVE_PERIOD",
	//"LIVE_PC_TIME",
	//"NATL_TV_BROADCASTER_ABBREVIATION",
	//"LIVE_PERIOD_TIME_BCAST",
	//"WH_STATUS"
	Teams      []BoxscoreSummaryTeam
	Officials  []BoxscoreSummaryOfficial
	LineScores []BoxscoreSummaryLineScore
	GameInfo   BoxscoreSummaryGameInfo
}

type BoxscoreSummaryGameInfo struct {
	//"GAME_DATE",
	//"ATTENDANCE",
	//"GAME_TIME",
}

type BoxscoreSummaryTeam struct {
	LeagueID           string
	TeamID             int
	TeamAbbreviation   string
	TeamCity           string
	PointsInPaint      int
	PointsSecondChance int
	PointsFastBreak    int
	LargestLead        int

	//"LEAGUE_ID",
	//"TEAM_ID",
	//"TEAM_ABBREVIATION",
	//"TEAM_CITY",
	//"PTS_PAINT",
	//"PTS_2ND_CHANCE",
	//"PTS_FB",
	//"LARGEST_LEAD",
	//"LEAD_CHANGES",
	//"TIMES_TIED",
	//"TEAM_TURNOVERS",
	//"TOTAL_TURNOVERS",
	//"TEAM_REBOUNDS",
	//"PTS_OFF_TO"
	LineScores []boxscoreSummaryLineScore
}

type boxscoreSummaryLineScore struct {
	Period int
	Score  int
}

type BoxscoreSummaryOfficial struct {
	OfficialID   int
	FirstName    string
	LastName     string
	JerseyNumber string
	//"OFFICIAL_ID",
	//"FIRST_NAME",
	//"LAST_NAME",
	//"JERSEY_NUM"
}

type BoxscoreSummaryLineScore struct {
	//"GAME_DATE_EST",
	//"GAME_SEQUENCE",
	//"GAME_ID",
	//"TEAM_ID",
	//"TEAM_ABBREVIATION",
	//"TEAM_CITY_NAME",
	//"TEAM_NICKNAME",
	//"TEAM_WINS_LOSSES",
	//"PTS_QTR1",
	//"PTS_QTR2",
	//"PTS_QTR3",
	//"PTS_QTR4",
	//"PTS_OT1",
	//"PTS_OT2",
	//"PTS_OT3",
	//"PTS_OT4",
	//"PTS_OT5",
	//"PTS_OT6",
	//"PTS_OT7",
	//"PTS_OT8",
	//"PTS_OT9",
	//"PTS_OT10",
	//"PTS"
}

// preferred: new boxscore with more stats, details, ids. however, returns error if game is in the future
func GetBoxscoreDetailed(ctx context.Context, r2Client cloudflare.Client, bucket string, gameID string, seasonStartYear int) (Boxscore, error) {
	filename := fmt.Sprintf(os.Getenv("STORAGE_PATH")+"/boxscore/%d/%s.json", seasonStartYear, gameID)

	url := fmt.Sprintf(BoxscoreURL, gameID)

	objectKey := fmt.Sprintf("boxscore/%d/%s_cdn.json", seasonStartYear, gameID)

	boxscore, err := fetchObjectFromFileOrURL[Boxscore](ctx, r2Client, url, filename, bucket, objectKey, func(b Boxscore) bool { return b.Final() })
	if err != nil {
		return Boxscore{}, err
	}

	return boxscore, nil
}

const boxscoreSummaryV2URL = "https://stats.nba.com/stats/boxscoresummaryv2?GameID=%s"

func (c client) GetBoxscoreSummary(ctx context.Context, gameID string, outputWriters ...OutputWriter) (BoxscoreSummary, error) {
	response, err := c.statsClient.Get(fmt.Sprintf(boxscoreSummaryV2URL, gameID))
	if err != nil {
		return BoxscoreSummary{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return BoxscoreSummary{}, fmt.Errorf("failed to get BoxscoreSummary object")
	}

	if response.Header.Get("Content-Encoding") == "gzip" {
		response.Body, err = gzip.NewReader(response.Body)
		if err != nil {
			return BoxscoreSummary{}, fmt.Errorf("failed to create gzip reader when getting nba boxscore summary: %w", err)
		}
	}

	var respBody []byte
	respBody, err = io.ReadAll(response.Body)

	boxscoreResult, err := unmarshalNBAHttpResponseToJSON[statsBaseResponse](bytes.NewReader(respBody))
	if err != nil {
		return BoxscoreSummary{}, fmt.Errorf("failed to get boxscore summary response: %w", err)
	}

	boxscoreSummary := BoxscoreSummary{}

	for _, resultSet := range boxscoreResult.ResultSets {
		headersMap := make(map[string]int, len(resultSet.Headers))
		for i, header := range resultSet.Headers {
			headersMap[header] = i
		}

		switch resultSet.Name {
		case "GameSummary":
			if len(resultSet.RowSet) != 1 {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse BoxscoreSummary from nba stats BoxscoreSummaryV2 endpoint: expected 1 row set and got %d", len(resultSet.RowSet))
			}
			rowSet := resultSet.RowSet[0]

			gameDateESTStr, err := parseRowSetValue[string](headersMap, rowSet, "GAME_DATE_EST")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse response from stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}
			eastCoastLoc, err := time.LoadLocation("America/New_York")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to load east coast location for parsing game date est from stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}
			gameDateEST, err := time.ParseInLocation(nbaStatsTimestampFormat, gameDateESTStr, eastCoastLoc)
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse game date est from stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}
			gameDateUTC := gameDateEST.UTC()

			gameSequence, err := parseRowSetValue[float64](headersMap, rowSet, "GAME_SEQUENCE")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse game sequence from stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			gID, err := parseRowSetValue[string](headersMap, rowSet, "GAME_ID")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse gameID from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			gameStatusID, err := parseRowSetValue[float64](headersMap, rowSet, "GAME_STATUS_ID")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse game status ID from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			gameStatusText, err := parseRowSetValue[string](headersMap, rowSet, "GAME_STATUS_TEXT")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse game status text from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			gameCode, err := parseRowSetValue[string](headersMap, rowSet, "GAMECODE")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse game code from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			homeTeamID, err := parseRowSetValue[float64](headersMap, rowSet, "HOME_TEAM_ID")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse home team ID from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			visitorTeamID, err := parseRowSetValue[float64](headersMap, rowSet, "VISITOR_TEAM_ID")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse visitor team ID from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			seasonStartYearStr, err := parseRowSetValue[string](headersMap, rowSet, "SEASON")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse season start year from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			seasonStartYear, err := strconv.Atoi(seasonStartYearStr)
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to convert season start year string to int from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			livePeriod, err := parseRowSetValue[float64](headersMap, rowSet, "LIVE_PERIOD")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse live period from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			livePCTime, err := parseRowSetValue[string](headersMap, rowSet, "LIVE_PC_TIME")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse live pc time from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			natlTVBroadcasterAbbr, err := parseRowSetValue[string](headersMap, rowSet, "NATL_TV_BROADCASTER_ABBREVIATION")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse natl tv broadcaster abbreviation from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			livePeriodTimeBroadcast, err := parseRowSetValue[string](headersMap, rowSet, "LIVE_PERIOD_TIME_BCAST")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse live period time broadcast from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			whStatus, err := parseRowSetValue[float64](headersMap, rowSet, "WH_STATUS")
			if err != nil {
				return BoxscoreSummary{}, fmt.Errorf("failed to parse wh status from nba stats %s endpoint: %w", endpointNameBoxscoreSummary, err)
			}

			boxscoreSummary.GameDate = gameDateUTC
			boxscoreSummary.GameInSequence = int(gameSequence)
			boxscoreSummary.GameID = gID
			boxscoreSummary.GameStatusID = int(gameStatusID)
			boxscoreSummary.GameStatusText = gameStatusText
			boxscoreSummary.GameCode = gameCode
			boxscoreSummary.HomeTeam.TeamID = int(homeTeamID)
			boxscoreSummary.AwayTeam.TeamID = int(visitorTeamID)
			boxscoreSummary.SeasonStartYear = seasonStartYear
			boxscoreSummary.Period = int(livePeriod)
			boxscoreSummary.PCTime = livePCTime
			boxscoreSummary.NationalTVBroadcasterAbbreviation = natlTVBroadcasterAbbr
			boxscoreSummary.PeriodTimeBroadcast = livePeriodTimeBroadcast
			boxscoreSummary.WHStatus = int(whStatus)

			break
		case "OtherStats":

			break
		case "Officials":
			break
		case "InactivePlayers":
			break
		case "GameInfo":
			break
		case "LineScore":
			break
		}

	}
	for _, outputWriter := range outputWriters {
		if err := outputWriter.Put(ctx, respBody); err != nil {
			return BoxscoreSummary{}, fmt.Errorf("failed to write output for boxscore summary for game: %w", err)
		}
	}

	return boxscoreSummary, nil
}
