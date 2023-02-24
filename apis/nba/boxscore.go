package nba

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	log "github.com/sirupsen/logrus"
)

// play by play https://cdn.nba.com/static/json/liveData/playbyplay/playbyplay_0022000180.json

const BoxscoreURL = "https://cdn.nba.com/static/json/liveData/boxscore/boxscore_%s.json"
const OldBoxscoreURL = "https://data.nba.net/prod/v1/%s/%s_boxscore.json"

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

type BoxscoreOld struct {
	StatsNode *struct {
		LeadChanges  string        `json:"leadChanges"`
		TimesTied    string        `json:"timesTied"`
		HomeTeamNode TeamStats     `json:"hTeam"`
		AwayTeamNode TeamStats     `json:"vTeam"`
		PlayersStats []PlayerStats `json:"activePlayers"`
	} `json:"stats,omitempty"`
	BasicGameDataNode struct {
		LeagueName           string            `json:"leagueName"` // ex. standard, vegas, sacramento, etc.
		StatusNum            int               `json:"statusNum"`  // 1 - upcoming 2 - started 3 - completed
		SeasonYear           string            `json:"seasonYear"` // ex. 2022
		GameID               string            `json:"gameId"`
		Arena                ArenaInfo         `json:"arena"`
		Attendance           string            `json:"attendance"`
		Clock                string            `json:"clock"`
		GameIsActivated      bool              `json:"isGameActivated"` // see UpdateTeamsRegularSeasonRecords
		GameStartTimeEastern string            `json:"startTimeEastern"`
		GameStartDateEastern string            `json:"startDateEastern"`
		GameStartTimeUTC     datetime          `json:"startTimeUTC"`
		GameEndTimeUTC       *time.Time        `json:"endTimeUTC,omitempty"`
		HomeTeamInfo         TeamBoxscoreInfo  `json:"hTeam"`
		AwayTeamInfo         TeamBoxscoreInfo  `json:"vTeam"`
		PlayoffsNode         *PlayoffsGameInfo `json:"playoffs,omitempty"`
		// season stage IDs
		// 1: preseason
		// 2: regular season
		// 3: all star
		// 4: playoffs
		// 5: play in
		SeasonStage seasonStage `json:"seasonStageId"`

		GameDurationNode *struct {
			Hours   string `json:"hours"`
			Minutes string `json:"minutes"`
		} `json:"gameDuration"`

		PeriodNode struct {
			CurrentPeriod int `json:"current"`
		} `json:"period"`

		RefereeNode struct {
			Referees []RefereeInfo `json:"formatted"`
		} `json:"officials"`

		WatchNode struct {
			BroadcastNode struct {
				BroadcastersNode struct {
					HomeTeamVideoFeeds []VideoBroadcasterInfo `json:"hTeam"`
					AwayTeamVideoFeeds []VideoBroadcasterInfo `json:"vTeam"`
					NationalVideoFeeds []VideoBroadcasterInfo `json:"national"`
				} `json:"broadcasters"`
				VideoBroadcastInfo GameVideoBroadcastInfo `json:"video"`
			} `json:"broadcast"`
		} `json:"watch"`
	} `json:"basicGameData"`
}

func (b *BoxscoreOld) IsPlayoffGame() bool {
	return b.BasicGameDataNode.SeasonStage == postSeason || b.BasicGameDataNode.PlayoffsNode != nil
}

func (b *BoxscoreOld) Final() bool {
	return b.BasicGameDataNode.StatusNum == 3
}

func (b *BoxscoreOld) GameEnded() bool {
	if b.StatsNode == nil {
		// the nba api will post a boxscore without the stats json node for some time before games
		return false
	}
	hasEndTime := b.BasicGameDataNode.GameEndTimeUTC != nil
	if hasEndTime {
		log.Println("endTimeUTC reported")
		return true
	}
	noTimeRemaining := b.BasicGameDataNode.Clock == "0.0" || b.BasicGameDataNode.Clock == ""

	homeTeamPoints, err := strconv.Atoi(b.StatsNode.HomeTeamNode.TeamStatsTotals.Points)
	if err != nil {
		log.Fatal("could not convert home team points to int")
	}
	awayTeamPoints, err := strconv.Atoi(b.StatsNode.AwayTeamNode.TeamStatsTotals.Points)
	if err != nil {
		log.Fatal("could not convert away team points to int")
	}

	gameNotTied := homeTeamPoints != awayTeamPoints
	log.Println(fmt.Sprintf("clock: %s", b.BasicGameDataNode.Clock))
	log.Println(fmt.Sprintf("noTimeRemaining: %t", noTimeRemaining))
	log.Println(fmt.Sprintf("gameNotTied: %t", gameNotTied))
	gameEndingPeriod := b.BasicGameDataNode.PeriodNode.CurrentPeriod >= 4
	log.Println(fmt.Sprintf("currentPeriod: %d", b.BasicGameDataNode.PeriodNode.CurrentPeriod))
	log.Println(fmt.Sprintf("gameEndPeriod: %t", gameEndingPeriod))
	if noTimeRemaining && gameNotTied && gameEndingPeriod {
		return true
	}
	return false
}

func (b *BoxscoreOld) DurationUntilGameStarts() (time.Duration, error) {
	currentTimeUTC := time.Now().UTC()
	// Issues occur when using eastern time for "today's games" as games on the west coast can still be going on
	// when the eastern time rolls over into the next day
	eastCoastLocation, locationError := time.LoadLocation("America/New_York")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeEastern := currentTimeUTC.In(eastCoastLocation)

	gameTime, err := makeGoTimeFromAPIData(b.BasicGameDataNode.GameStartTimeEastern, b.BasicGameDataNode.GameStartDateEastern)
	if err != nil {
		return *new(time.Duration), err
	}

	return gameTime.Sub(currentTimeEastern), nil
}

func (b *BoxscoreOld) GameStarted() (bool, error) {
	currentTimeUTC := time.Now().UTC()
	// Issues occur when using eastern time for "today's games" as games on the west coast can still be going on
	// when the eastern time rolls over into the next day
	eastCoastLocation, locationError := time.LoadLocation("America/New_York")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeEastern := currentTimeUTC.In(eastCoastLocation)

	gameTime, err := makeGoTimeFromAPIData(b.BasicGameDataNode.GameStartTimeEastern, b.BasicGameDataNode.GameStartDateEastern)
	if err != nil {
		return false, err
	}

	if currentTimeEastern.After(gameTime) {
		return true, nil
	}
	return false, nil
}

func (b *BoxscoreOld) GetOpponent(team TriCode) TriCode {
	if b.BasicGameDataNode.HomeTeamInfo.TriCode == team {
		return b.BasicGameDataNode.AwayTeamInfo.TriCode
	} else {
		return b.BasicGameDataNode.HomeTeamInfo.TriCode
	}
}

func incrementString(str string) string {
	// convert string to a number
	i, err := strconv.Atoi(str)
	if err != nil {
		log.Fatal("could not convert string: " + str + "to int")
	}

	// add one to the number
	i = i + 1

	// convert number back to string
	str = strconv.FormatInt(int64(i), 10)

	return str
}

func (b *BoxscoreOld) UpdateTeamsRegularSeasonRecords(currentGameNumber int) {
	if b.IsPlayoffGame() {
		return
	}

	log.Println(fmt.Sprintf("GameIsActivated: %t", b.BasicGameDataNode.GameIsActivated))
	// the nba does not appear to update the series wins and losses right after the game for either team for regular reason series records; update them based on the result of the game
	// they do eventually update the series wins and losses, but by then we should have already posted the thread
	// isGameActivated might be the trigger/think to look at for if the series has been updated see https://github.com/f1uk3r/Some-Python-Scripts/blob/master/reddit-nba-bot/reddit-boxscore-bot.py
	// update: this does not appear to be reliable either
	// update 10/30/2019: the nba appears to be updating records in time
	// update 12/11/2019: the nba appears to be updating some records in time but it's not consistent; removing this check so that we rely on updating vs what game of the season it is always
	/*if !b.BasicGameDataNode.GameIsActivated {
		return
	}*/

	homeTeamWins, err := strconv.Atoi(b.BasicGameDataNode.HomeTeamInfo.Wins)
	if err != nil {
		log.Fatal("could not convert home wins to int")
	}
	homeTeamLosses, err := strconv.Atoi(b.BasicGameDataNode.HomeTeamInfo.Losses)
	if err != nil {
		log.Fatal("could not convert home losses to int")
	}

	if homeTeamWins+homeTeamLosses >= currentGameNumber {
		log.Println("regular season records already updated")
		return
	}

	log.Println("Updating team regular season records")

	homeTeamPoints, err := strconv.Atoi(b.StatsNode.HomeTeamNode.TeamStatsTotals.Points)
	if err != nil {
		log.Fatal("could not convert home regular season points to int")
	}
	awayTeamPoints, err := strconv.Atoi(b.StatsNode.AwayTeamNode.TeamStatsTotals.Points)
	if err != nil {
		log.Fatal("could not convert away regular season points to int")
	}

	homeTeamWon := homeTeamPoints > awayTeamPoints
	if homeTeamWon {
		b.BasicGameDataNode.HomeTeamInfo.SeriesWins = incrementString(b.BasicGameDataNode.HomeTeamInfo.SeriesWins)
		b.BasicGameDataNode.AwayTeamInfo.SeriesLosses = incrementString(b.BasicGameDataNode.AwayTeamInfo.SeriesLosses)
		b.BasicGameDataNode.HomeTeamInfo.Wins = incrementString(b.BasicGameDataNode.HomeTeamInfo.Wins)
		b.BasicGameDataNode.AwayTeamInfo.Losses = incrementString(b.BasicGameDataNode.AwayTeamInfo.Losses)
	} else {
		b.BasicGameDataNode.HomeTeamInfo.SeriesLosses = incrementString(b.BasicGameDataNode.HomeTeamInfo.SeriesLosses)
		b.BasicGameDataNode.AwayTeamInfo.SeriesWins = incrementString(b.BasicGameDataNode.AwayTeamInfo.SeriesWins)
		b.BasicGameDataNode.HomeTeamInfo.Losses = incrementString(b.BasicGameDataNode.HomeTeamInfo.Losses)
		b.BasicGameDataNode.AwayTeamInfo.Wins = incrementString(b.BasicGameDataNode.AwayTeamInfo.Wins)
	}
}

func (b *BoxscoreOld) UpdateTeamsPlayoffsSeriesRecords() {
	if !b.IsPlayoffGame() {
		return
	}

	log.Println("Updating team playoff records")

	homeWins, err := strconv.Atoi(b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.SeriesWins)
	if err != nil {
		log.Fatal("could not convert home playoff series wins to int")
	}
	awayWins, err := strconv.Atoi(b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.SeriesWins)
	if err != nil {
		log.Fatal("could not convert away playoff series wins to int")
	}
	gameInSeries, err := strconv.Atoi(b.BasicGameDataNode.PlayoffsNode.GameInSeries)
	if err != nil {
		log.Fatal("could not convert away playoff series wins to int")
	}
	log.Println(fmt.Sprintf("gameInSeries: %d", gameInSeries))
	if (homeWins + awayWins) != gameInSeries {
		log.Println("updating playoff series records")

		homeTeamPoints, err := strconv.Atoi(b.StatsNode.HomeTeamNode.TeamStatsTotals.Points)
		if err != nil {
			log.Fatal("could not convert home playoff points to int")
		}
		awayTeamPoints, err := strconv.Atoi(b.StatsNode.AwayTeamNode.TeamStatsTotals.Points)
		if err != nil {
			log.Fatal("could not convert away playoff points to int")
		}

		homeTeamWon := homeTeamPoints > awayTeamPoints
		if homeTeamWon {
			b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.SeriesWins = incrementString(b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.SeriesWins)
			b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.WonSeries = b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.SeriesWins == "4"
		} else {
			b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.SeriesWins = incrementString(b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.SeriesWins)
			b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.WonSeries = b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.SeriesWins == "4"
		}
	}
}

func getTeamLeaders(playersStats []PlayerStats, homeTeamID string, awayTeamID string) (TeamLeaders, TeamLeaders) {
	var homeTeamLeaders, awayTeamLeaders TeamLeaders
	for _, playerStats := range playersStats {
		if len(playerStats.DidNotPlayStatus) > 0 {
			// player did not play and instead of populating stats with 0's the nba api sends us "" which can break with the logic below
			continue
		}

		var teamLeaders *TeamLeaders
		if playerStats.TeamID == homeTeamID {
			teamLeaders = &homeTeamLeaders
		} else if playerStats.TeamID == awayTeamID {
			teamLeaders = &awayTeamLeaders
		}

		playerPoints, err := strconv.Atoi(playerStats.Points)
		if err != nil {
			log.Fatal("failed to convert points to string")
		}

		if playerPoints == teamLeaders.Points {
			teamLeaders.PointsLeaders = append(teamLeaders.PointsLeaders, playerStats.ID)
		} else if playerPoints > teamLeaders.Points {
			teamLeaders.PointsLeaders = nil
			teamLeaders.Points = playerPoints
			teamLeaders.PointsLeaders = append(teamLeaders.PointsLeaders, playerStats.ID)
		}

		playerRebounds, err := strconv.Atoi(playerStats.TotalRebounds)
		if err != nil {
			log.Fatal("failed to convert rebounds to string")
		}

		if playerRebounds == teamLeaders.Rebounds {
			teamLeaders.ReboundsLeaders = append(teamLeaders.ReboundsLeaders, playerStats.ID)
		} else if playerRebounds > teamLeaders.Rebounds {
			teamLeaders.ReboundsLeaders = nil
			teamLeaders.Rebounds = playerRebounds
			teamLeaders.ReboundsLeaders = append(teamLeaders.ReboundsLeaders, playerStats.ID)
		}

		playerAssists, err := strconv.Atoi(playerStats.Assists)
		if err != nil {
			log.Fatal("failed to convert assists to string")
		}

		if playerAssists == teamLeaders.Assists {
			teamLeaders.AssistsLeaders = append(teamLeaders.AssistsLeaders, playerStats.ID)
		} else if playerAssists > teamLeaders.Assists {
			teamLeaders.AssistsLeaders = nil
			teamLeaders.Assists = playerAssists
			teamLeaders.AssistsLeaders = append(teamLeaders.AssistsLeaders, playerStats.ID)
		}

		playerBlocks, err := strconv.Atoi(playerStats.Blocks)
		if err != nil {
			log.Fatal("failed to convert blocks to string")
		}

		if playerBlocks == teamLeaders.Blocks {
			teamLeaders.BlocksLeaders = append(teamLeaders.BlocksLeaders, playerStats.ID)
		} else if playerBlocks > teamLeaders.Blocks {
			teamLeaders.BlocksLeaders = nil
			teamLeaders.Blocks = playerBlocks
			teamLeaders.BlocksLeaders = append(teamLeaders.BlocksLeaders, playerStats.ID)
		}

		playerSteals, err := strconv.Atoi(playerStats.Steals)
		if err != nil {
			log.Fatal("failed to convert steals to string")
		}

		if playerSteals == teamLeaders.Steals {
			teamLeaders.StealsLeaders = append(teamLeaders.StealsLeaders, playerStats.ID)
		} else if playerSteals > teamLeaders.Steals {
			teamLeaders.StealsLeaders = nil
			teamLeaders.Steals = playerSteals
			teamLeaders.StealsLeaders = append(teamLeaders.StealsLeaders, playerStats.ID)
		}
	}
	return homeTeamLeaders, awayTeamLeaders
}

func getGameInfoTableString(arenaName, city, country, startTimeEastern, startDateEastern, attendance string, gameThread bool, otherThreadURL string) string {
	gameInfoTableString := "|**Game Info**||\n"
	gameInfoTableString += "|:-:|:-:|\n"
	gameInfoTableString += fmt.Sprintf("|**Arena**|%s|\n", arenaName)
	gameInfoTableString += fmt.Sprintf("|**Location**|%s, %s|\n", city, country)
	gameInfoTableString += fmt.Sprintf("|**Attendance**|%s|\n", attendance)

	gameTimeEastern, err := makeGoTimeFromAPIData(startTimeEastern, startDateEastern)
	if err != nil {
		log.Fatal(err)
	}
	centralLocation, locationErr := time.LoadLocation("America/Chicago")
	if locationErr != nil {
		log.Fatal("Failed to load Minneapolis location")
	}
	gameTimeCentral := gameTimeEastern.In(centralLocation)
	gameTimeCentralString := gameTimeCentral.Format("3:04 PM MST")

	gameTimeHour := gameTimeCentral.Hour()
	timePM := false
	if gameTimeHour > 12 {
		gameTimeHour = gameTimeHour - 12
		timePM = true
	}
	gameTimeHourMinute := fmt.Sprintf("%02d%02d", gameTimeHour, gameTimeCentral.Minute())
	if timePM {
		gameTimeHourMinute += "PM"
	} else {
		gameTimeHourMinute += "AM"
	}
	timeStringURL := fmt.Sprintf("https://time.is/compare/%s_%v_%s_%v__in_Minneapolis", gameTimeHourMinute, gameTimeCentral.Day(), gameTimeCentral.Month().String(), gameTimeCentral.Year())

	timeString := fmt.Sprintf("%s [other time zones](%s)", gameTimeCentralString, timeStringURL)

	gameInfoTableString += fmt.Sprintf("|**Time**|%s|\n", timeString)

	if gameThread {
		gameInfoTableString += fmt.Sprintf("|**Post Game Thread**|[link](%s)|\n", otherThreadURL)
	} else {
		gameInfoTableString += fmt.Sprintf("|**Game Thread**|[link](%s)|\n", otherThreadURL)
	}

	return gameInfoTableString
}

func getTeamQuarterScoreTableString(homeTeamBoxscoreInfo TeamBoxscoreInfo, homeTeamStats TeamStatsTotals, awayTeamBoxscoreInfo TeamBoxscoreInfo, awayTeamStats TeamStatsTotals) string {
	if len(homeTeamBoxscoreInfo.PointsByQuarter) != len(awayTeamBoxscoreInfo.PointsByQuarter) {
		log.Fatal("Home team and away team line scores are different lengths")
	}

	quarterScoreTableString := "||"
	// format for team tricode
	quarterScoreTableFormatString := "|:-:|"
	homeTeamLinescoreString := fmt.Sprintf("|%s|", homeTeamBoxscoreInfo.TriCode)
	awayTeamLinescoreString := fmt.Sprintf("|%s|", awayTeamBoxscoreInfo.TriCode)
	for i := 1; i <= len(homeTeamBoxscoreInfo.PointsByQuarter); i++ {
		if i < 5 {
			// regular time quarter
			quarterScoreTableString += fmt.Sprintf("**Q%d**|", i)
		} else {
			// overtime quarter
			quarterScoreTableString += fmt.Sprintf("**%dOT**|", i-4)
		}

		// quarter score format
		quarterScoreTableFormatString += ":-:|"

		homeTeamLinescoreString += fmt.Sprintf("%s|", homeTeamBoxscoreInfo.PointsByQuarter[i-1].Points)
		awayTeamLinescoreString += fmt.Sprintf("%s|", awayTeamBoxscoreInfo.PointsByQuarter[i-1].Points)
	}
	homeTeamLinescoreString += fmt.Sprintf("%s|", homeTeamStats.Points)
	homeTeamLinescoreString += "\n"
	awayTeamLinescoreString += fmt.Sprintf("%s|", awayTeamStats.Points)
	awayTeamLinescoreString += "\n"

	// total score format
	quarterScoreTableFormatString += ":-:|"
	quarterScoreTableFormatString += "\n"

	quarterScoreTableString += "**Total**|\n"
	quarterScoreTableString += quarterScoreTableFormatString
	quarterScoreTableString += homeTeamLinescoreString
	quarterScoreTableString += awayTeamLinescoreString

	return quarterScoreTableString
}

func getTeamStatsCategoryLeadersString(stat int, leaders []string, players map[string]Player) string {
	statLeadersString := ""
	if stat == 0 {
		return statLeadersString
	}

	numStatLeaders := len(leaders)
	for index, player := range leaders {
		statLeadersString += getPlayerString(player, players)
		if index != numStatLeaders-1 {
			statLeadersString += ", "
		}
	}
	return statLeadersString
}

func getTeamLeadersTableString(homeTeamBoxscoreInfo TeamBoxscoreInfo, homeTeamLeaders TeamLeaders, awayTeamBoxscoreInfo TeamBoxscoreInfo, awayTeamLeaders TeamLeaders, players map[string]Player) string {
	teamLeadersTableString := ""
	teamLeadersTableString += "|**Team Leaders**|**PTS**|**REB**|**AST**|**STL**|**BLK**|\n"
	teamLeadersTableString += "|:-:|:-:|:-:|:-:|:-:|:-:|\n"

	teamLeadersString := "|%s|%s (%d)|%s (%d)|%s (%d)|%s (%d)|%s (%d)|\n"

	homeTeamPointsLeadersString := getTeamStatsCategoryLeadersString(homeTeamLeaders.Points, homeTeamLeaders.PointsLeaders, players)
	homeTeamReboundsLeadersString := getTeamStatsCategoryLeadersString(homeTeamLeaders.Rebounds, homeTeamLeaders.ReboundsLeaders, players)
	homeTeamAssistsLeadersString := getTeamStatsCategoryLeadersString(homeTeamLeaders.Assists, homeTeamLeaders.AssistsLeaders, players)
	homeTeamStealsLeadersString := getTeamStatsCategoryLeadersString(homeTeamLeaders.Steals, homeTeamLeaders.StealsLeaders, players)
	homeTeamBlocksLeadersString := getTeamStatsCategoryLeadersString(homeTeamLeaders.Blocks, homeTeamLeaders.BlocksLeaders, players)

	awayTeamPointsLeadersString := getTeamStatsCategoryLeadersString(awayTeamLeaders.Points, awayTeamLeaders.PointsLeaders, players)
	awayTeamReboundsLeadersString := getTeamStatsCategoryLeadersString(awayTeamLeaders.Rebounds, awayTeamLeaders.ReboundsLeaders, players)
	awayTeamAssistsLeadersString := getTeamStatsCategoryLeadersString(awayTeamLeaders.Assists, awayTeamLeaders.AssistsLeaders, players)
	awayTeamStealsLeadersString := getTeamStatsCategoryLeadersString(awayTeamLeaders.Steals, awayTeamLeaders.StealsLeaders, players)
	awayTeamBlocksLeadersString := getTeamStatsCategoryLeadersString(awayTeamLeaders.Blocks, awayTeamLeaders.BlocksLeaders, players)

	teamLeadersTableString += fmt.Sprintf(teamLeadersString, homeTeamBoxscoreInfo.TriCode, homeTeamPointsLeadersString, homeTeamLeaders.Points, homeTeamReboundsLeadersString, homeTeamLeaders.Rebounds, homeTeamAssistsLeadersString, homeTeamLeaders.Assists, homeTeamStealsLeadersString, homeTeamLeaders.Steals, homeTeamBlocksLeadersString, homeTeamLeaders.Blocks)
	teamLeadersTableString += fmt.Sprintf(teamLeadersString, awayTeamBoxscoreInfo.TriCode, awayTeamPointsLeadersString, awayTeamLeaders.Points, awayTeamReboundsLeadersString, awayTeamLeaders.Rebounds, awayTeamAssistsLeadersString, awayTeamLeaders.Assists, awayTeamStealsLeadersString, awayTeamLeaders.Steals, awayTeamBlocksLeadersString, awayTeamLeaders.Blocks)

	return teamLeadersTableString
}

func getTeamStatsTableString(teamBoxscoreInfo TeamBoxscoreInfo, teamStats TeamStatsTotals, players map[string]Player, playersStats []PlayerStats) string {
	columnHeader := "|**[](/%s) %s**|**MIN**|**PTS**|**FG**|**3PT**|**FT**|**OREB**|**REB**|**AST**|**TOV**|**STL**|**BLK**|**PF**|**+/-**|\n"
	columnHeaderSeparator := "|:---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n"
	playerStatsString := "|%s|%s|%s|%s-%s|%s-%s|%s-%s|%s|%s|%s|%s|%s|%s|%s|%s|\n"
	totalsString := "|Totals|%s|%s|%s-%s (%s%%)|%s-%s (%s%%)|%s-%s (%s%%)|%s|%s|%s|%s|%s|%s|%s|-|\n"
	teamStatsTableString := ""
	teamStatsTableString += fmt.Sprintf(columnHeader, teamBoxscoreInfo.TriCode, teamBoxscoreInfo.TriCode)
	teamStatsTableString += columnHeaderSeparator
	for _, playerStats := range playersStats {
		if playerStats.TeamID == teamBoxscoreInfo.TeamID {
			playerString := getPlayerString(playerStats.ID, players)
			playerMinutesString := playerStats.Minutes
			if len(playerStats.DidNotPlayStatus) > 0 && len(playerStats.Minutes) == 0 {
				playerMinutesString = playerStats.DidNotPlayStatus
			}

			teamStatsTableString += fmt.Sprintf(playerStatsString, playerString, playerMinutesString, playerStats.Points, playerStats.FieldGoalsMade, playerStats.FieldGoalsAttempted, playerStats.ThreePointsMade, playerStats.ThreePointsAttempted, playerStats.FreeThrowsMade, playerStats.FreeThrowsAttempted, playerStats.OffensiveRebounds, playerStats.TotalRebounds, playerStats.Assists, playerStats.Turnovers, playerStats.Steals, playerStats.Blocks, playerStats.PersonalFouls, playerStats.PlusMinus)
		}
	}
	teamStatsTableString += fmt.Sprintf(totalsString, teamStats.Minutes, teamStats.Points, teamStats.FieldGoalsMade, teamStats.FieldGoalsAttempted, teamStats.FieldGoalPercentage, teamStats.ThreePointsMade, teamStats.ThreePointsAttempted, teamStats.ThreePointPercentage, teamStats.FreeThrowsMade, teamStats.FreeThrowsAttempted, teamStats.FreeThrowPercentage, teamStats.OffensiveRebounds, teamStats.TotalRebounds, teamStats.Assists, teamStats.Turnovers, teamStats.Steals, teamStats.Blocks, teamStats.PersonalFouls)
	return teamStatsTableString
}

func getRefereeTableString(refereesInfo []RefereeInfo) string {
	refereeTableString := ""

	if len(refereesInfo) <= 0 {
		return refereeTableString
	}

	refereeTableString += "|**Referees**|"

	refereeInfosString := "|"
	refereeFormatString := "|"

	for i := 0; i < len(refereesInfo)-1; i++ {
		refereeTableString += "|"
	}
	for _, referee := range refereesInfo {
		refereeFormatString += ":-:|"
		refereeInfosString += fmt.Sprintf("%s|", referee.FullName)
	}

	refereeTableString += "\n"
	refereeTableString += refereeFormatString
	refereeTableString += "\n"
	refereeTableString += refereeInfosString
	refereeTableString += "\n"

	return refereeTableString
}

func getBroadcastInfoTable(gameVideoBroadcastInfo GameVideoBroadcastInfo, nationalVideoFeeds []VideoBroadcasterInfo, homeTeamTriCode TriCode, homeTeamVideoFeeds []VideoBroadcasterInfo, awayTeamTriCode TriCode, awayTeamVideoFeeds []VideoBroadcasterInfo) string {
	broadcastInfoTable := ""

	broadcastInfoTable += "||**Feeds**|"
	broadcastInfoTable += "\n"

	broadcastInfoTable += "|**National**| "

	nationalFeedStrs := []string{}

	if gameVideoBroadcastInfo.LeaguePass {
		nationalFeedStrs = append(nationalFeedStrs, "League Pass")
	}

	for _, nationalVideoFeed := range nationalVideoFeeds {
		nationalFeedStrs = append(nationalFeedStrs, nationalVideoFeed.LongName)
	}

	broadcastInfoTable += strings.Join(nationalFeedStrs, " ") + " |"
	broadcastInfoTable += "\n"

	broadcastInfoTable += fmt.Sprintf("|**%s**| ", homeTeamTriCode)

	homeTeamFeedsStrs := []string{}

	for _, homeTeamVideoFeed := range homeTeamVideoFeeds {
		homeTeamFeedsStrs = append(homeTeamFeedsStrs, homeTeamVideoFeed.LongName)
	}

	broadcastInfoTable += strings.Join(homeTeamFeedsStrs, " ") + " |"
	broadcastInfoTable += "\n"

	broadcastInfoTable += fmt.Sprintf("|**%s**| ", awayTeamTriCode)

	awayTeamFeedsStrs := []string{}

	for _, awayTeamVideoFeed := range awayTeamVideoFeeds {
		awayTeamFeedsStrs = append(awayTeamFeedsStrs, awayTeamVideoFeed.LongName)
	}

	broadcastInfoTable += strings.Join(awayTeamFeedsStrs, " ") + " |"

	return broadcastInfoTable
}

func (b *BoxscoreOld) GetRedditPostGameThreadBodyString(players map[string]Player, gameThreadURL string) string {
	body := ""
	body += getGameInfoTableString(b.BasicGameDataNode.Arena.Name, b.BasicGameDataNode.Arena.City, b.BasicGameDataNode.Arena.Country, b.BasicGameDataNode.GameStartTimeEastern, b.BasicGameDataNode.GameStartDateEastern, b.BasicGameDataNode.Attendance, false /*gameThread*/, gameThreadURL)
	body += "\n"
	body += getBroadcastInfoTable(b.BasicGameDataNode.WatchNode.BroadcastNode.VideoBroadcastInfo, b.BasicGameDataNode.WatchNode.BroadcastNode.BroadcastersNode.NationalVideoFeeds, b.BasicGameDataNode.HomeTeamInfo.TriCode, b.BasicGameDataNode.WatchNode.BroadcastNode.BroadcastersNode.HomeTeamVideoFeeds, b.BasicGameDataNode.AwayTeamInfo.TriCode, b.BasicGameDataNode.WatchNode.BroadcastNode.BroadcastersNode.AwayTeamVideoFeeds)
	body += "\n"
	body += getTeamQuarterScoreTableString(b.BasicGameDataNode.HomeTeamInfo, b.StatsNode.HomeTeamNode.TeamStatsTotals, b.BasicGameDataNode.AwayTeamInfo, b.StatsNode.AwayTeamNode.TeamStatsTotals)
	body += "\n"

	homeTeamLeaders, awayTeamLeaders := getTeamLeaders(b.StatsNode.PlayersStats, b.BasicGameDataNode.HomeTeamInfo.TeamID, b.BasicGameDataNode.AwayTeamInfo.TeamID)
	body += getTeamLeadersTableString(b.BasicGameDataNode.HomeTeamInfo, homeTeamLeaders, b.BasicGameDataNode.AwayTeamInfo, awayTeamLeaders, players)
	body += "\n"

	body += getTeamStatsTableString(b.BasicGameDataNode.HomeTeamInfo, b.StatsNode.HomeTeamNode.TeamStatsTotals, players, b.StatsNode.PlayersStats)
	body += "\n"
	body += getTeamStatsTableString(b.BasicGameDataNode.AwayTeamInfo, b.StatsNode.AwayTeamNode.TeamStatsTotals, players, b.StatsNode.PlayersStats)
	body += "\n"
	body += getRefereeTableString(b.BasicGameDataNode.RefereeNode.Referees)
	body += "\n"
	return body
}

func (b *BoxscoreOld) GetRedditPostGameThreadTitle(teamTriCode TriCode, teams []Team) string {
	title := ""
	firstTeam := Team{}
	firstTeamStats := TeamStatsTotals{}
	firstTeamInfo := TeamBoxscoreInfo{}
	firstTeamPlayoffsGameTeamInfo := PlayoffsGameTeamInfo{}
	secondTeam := Team{}
	secondTeamStats := TeamStatsTotals{}
	secondTeamInfo := TeamBoxscoreInfo{}
	secondTeamPlayoffsGameTeamInfo := PlayoffsGameTeamInfo{}
	if b.BasicGameDataNode.HomeTeamInfo.TriCode == teamTriCode {
		firstTeam = findTeamWithTricode(b.BasicGameDataNode.HomeTeamInfo.TriCode, teams)
		firstTeamStats = b.StatsNode.HomeTeamNode.TeamStatsTotals
		firstTeamInfo = b.BasicGameDataNode.HomeTeamInfo
		secondTeam = findTeamWithTricode(b.BasicGameDataNode.AwayTeamInfo.TriCode, teams)
		secondTeamStats = b.StatsNode.AwayTeamNode.TeamStatsTotals
		secondTeamInfo = b.BasicGameDataNode.AwayTeamInfo

		if b.IsPlayoffGame() {
			firstTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo
			secondTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo
		}
	} else {
		firstTeam = findTeamWithTricode(b.BasicGameDataNode.AwayTeamInfo.TriCode, teams)
		firstTeamStats = b.StatsNode.AwayTeamNode.TeamStatsTotals
		firstTeamInfo = b.BasicGameDataNode.AwayTeamInfo
		secondTeam = findTeamWithTricode(b.BasicGameDataNode.HomeTeamInfo.TriCode, teams)
		secondTeamStats = b.StatsNode.HomeTeamNode.TeamStatsTotals
		secondTeamInfo = b.BasicGameDataNode.HomeTeamInfo

		if b.IsPlayoffGame() {
			firstTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo
			secondTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo
		}
	}

	firstTeamPointsInt, err := strconv.Atoi(firstTeamStats.Points)
	if err != nil {
		log.Fatal("could not convert team's points string to int")
	}
	secondTeamPointsInt, err := strconv.Atoi(secondTeamStats.Points)
	if err != nil {
		log.Fatal("could not convert team's points string to int")
	}

	firstTeamWon := firstTeamPointsInt > secondTeamPointsInt

	// Specific game details; blowout, overtime, etc.
	pointDifferential := math.Abs(float64(firstTeamPointsInt) - float64(secondTeamPointsInt))
	blowoutDifferential := 20
	blowout := pointDifferential >= float64(blowoutDifferential)
	overtime := b.BasicGameDataNode.PeriodNode.CurrentPeriod > 4

	teamRecordString := "(%s-%s)"

	title += "[POST GAME THREAD]"
	title += " "

	if b.IsPlayoffGame() {
		playoffsRoundInt, err := strconv.Atoi(b.BasicGameDataNode.PlayoffsNode.Round)
		if err != nil {
			log.Fatal(fmt.Sprintf("could not convert playoff round %s to int", b.BasicGameDataNode.PlayoffsNode.Round))
		}

		if playoffsRoundInt == 2 {
			title += fmt.Sprintf("%sERN CONF SEMIS", strings.ToUpper(b.BasicGameDataNode.PlayoffsNode.Conference))
		} else if playoffsRoundInt == 3 {
			title += fmt.Sprintf("%sERN CONF FINALS", strings.ToUpper(b.BasicGameDataNode.PlayoffsNode.Conference))
		} else if playoffsRoundInt == 4 {
			title += "NBA FINALS"
		} else {
			title += fmt.Sprintf("Playoffs Round %d", playoffsRoundInt)
		}
		title += ":"
		title += " "
		title += fmt.Sprintf("(%s)", firstTeamPlayoffsGameTeamInfo.Seed)
		title += " "
	}

	title += strings.ToUpper(firstTeam.Nickname)
	title += " "

	if !b.IsPlayoffGame() {
		title += fmt.Sprintf(teamRecordString, firstTeamInfo.Wins, firstTeamInfo.Losses)
		title += " "
	}

	if b.IsPlayoffGame() && b.BasicGameDataNode.PlayoffsNode.GameInSeries == "6" && firstTeamWon {
		title += "FORCE GAME 7 AGAINST THE"
	} else if b.IsPlayoffGame() && secondTeamPlayoffsGameTeamInfo.SeriesWins == "3" && firstTeamWon {
		title += "SURVIVE AGAINST THE"
	} else if b.IsPlayoffGame() && firstTeamWon && secondTeamPlayoffsGameTeamInfo.SeriesWins == "0" && firstTeamPlayoffsGameTeamInfo.WonSeries {
		title += "SWEEP THE"
	} else if b.IsPlayoffGame() && !firstTeamWon && firstTeamPlayoffsGameTeamInfo.SeriesWins == "0" && secondTeamPlayoffsGameTeamInfo.WonSeries {
		title += "GET SWEPT BY THE"
	} else if firstTeamWon && blowout {
		title += "BLOWOUT THE"
	} else if firstTeamWon {
		title += "BEAT THE"
	} else if !firstTeamWon && blowout {
		title += "GET BLOWN OUT BY THE"
	} else {
		// first team lost but did not get blown out
		title += "LOSE TO THE"
	}
	title += " "

	if b.IsPlayoffGame() {
		title += fmt.Sprintf("(%s)", secondTeamPlayoffsGameTeamInfo.Seed)
		title += " "
	}

	title += strings.ToUpper(secondTeam.Nickname)
	title += " "

	if !b.IsPlayoffGame() {
		title += fmt.Sprintf(teamRecordString, secondTeamInfo.Wins, secondTeamInfo.Losses)
		title += " "
	}

	if overtime {
		title += fmt.Sprintf("IN %dOT", b.BasicGameDataNode.PeriodNode.CurrentPeriod-4)
		title += " "
	}

	title += firstTeamStats.Points + "-" + secondTeamStats.Points

	if b.IsPlayoffGame() {
		title += ","
		title += " "

		// Playoff series info
		log.Println(fmt.Sprintf("%s: %s", firstTeamInfo.TriCode, firstTeamPlayoffsGameTeamInfo.SeriesWins))
		log.Println(fmt.Sprintf("%s: %s", secondTeamInfo.TriCode, secondTeamPlayoffsGameTeamInfo.SeriesWins))

		firstTeamWins, err := strconv.Atoi(firstTeamPlayoffsGameTeamInfo.SeriesWins)
		if err != nil {
			log.Fatal("failed to convert first team's series wins to int")
		}

		secondTeamWins, err := strconv.Atoi(secondTeamPlayoffsGameTeamInfo.SeriesWins)
		if err != nil {
			log.Fatal("failed to convert second team's series wins to int")
		}

		if firstTeamWins == secondTeamWins {
			if !firstTeamPlayoffsGameTeamInfo.WonSeries && !secondTeamPlayoffsGameTeamInfo.WonSeries {
				if firstTeamWon {
					title += "TIE SERIES"
				} else {
					title += "SERIES TIED"
				}
			} else {
				log.Fatal("a playoff series can't end tied")
			}
		} else if firstTeamPlayoffsGameTeamInfo.SeriesWins < secondTeamPlayoffsGameTeamInfo.SeriesWins {
			if !secondTeamPlayoffsGameTeamInfo.WonSeries {
				title += "TRAIL SERIES"
			} else {
				title += "LOSE SERIES"
			}
		} else {
			// first team leading series
			if !firstTeamPlayoffsGameTeamInfo.WonSeries {
				title += "LEAD SERIES"
			} else {
				title += "WIN SERIES"
			}
		}
		title += " "
		title += "(" + firstTeamPlayoffsGameTeamInfo.SeriesWins + "-" + secondTeamPlayoffsGameTeamInfo.SeriesWins + ")"
	}

	return title
}

func (b *BoxscoreOld) GetRedditGameThreadBodyString(players map[string]Player, postGameThreadURL string) string {
	body := ""
	body += getGameInfoTableString(b.BasicGameDataNode.Arena.Name, b.BasicGameDataNode.Arena.City, b.BasicGameDataNode.Arena.Country, b.BasicGameDataNode.GameStartTimeEastern, b.BasicGameDataNode.GameStartDateEastern, b.BasicGameDataNode.Attendance, true /*gameThread*/, postGameThreadURL)
	body += "\n"
	body += getBroadcastInfoTable(b.BasicGameDataNode.WatchNode.BroadcastNode.VideoBroadcastInfo, b.BasicGameDataNode.WatchNode.BroadcastNode.BroadcastersNode.NationalVideoFeeds, b.BasicGameDataNode.HomeTeamInfo.TriCode, b.BasicGameDataNode.WatchNode.BroadcastNode.BroadcastersNode.HomeTeamVideoFeeds, b.BasicGameDataNode.AwayTeamInfo.TriCode, b.BasicGameDataNode.WatchNode.BroadcastNode.BroadcastersNode.AwayTeamVideoFeeds)
	body += "\n"

	if b.StatsNode != nil {
		if len(b.BasicGameDataNode.HomeTeamInfo.PointsByQuarter) > 0 && len(b.BasicGameDataNode.AwayTeamInfo.PointsByQuarter) > 0 {
			body += getTeamQuarterScoreTableString(b.BasicGameDataNode.HomeTeamInfo, b.StatsNode.HomeTeamNode.TeamStatsTotals, b.BasicGameDataNode.AwayTeamInfo, b.StatsNode.AwayTeamNode.TeamStatsTotals)
			body += "\n"
		}

		homeTeamLeaders, awayTeamLeaders := getTeamLeaders(b.StatsNode.PlayersStats, b.BasicGameDataNode.HomeTeamInfo.TeamID, b.BasicGameDataNode.AwayTeamInfo.TeamID)
		body += getTeamLeadersTableString(b.BasicGameDataNode.HomeTeamInfo, homeTeamLeaders, b.BasicGameDataNode.AwayTeamInfo, awayTeamLeaders, players)
		body += "\n"

		body += getTeamStatsTableString(b.BasicGameDataNode.HomeTeamInfo, b.StatsNode.HomeTeamNode.TeamStatsTotals, players, b.StatsNode.PlayersStats)
		body += "\n"
		body += getTeamStatsTableString(b.BasicGameDataNode.AwayTeamInfo, b.StatsNode.AwayTeamNode.TeamStatsTotals, players, b.StatsNode.PlayersStats)
		body += "\n"
	}

	body += getRefereeTableString(b.BasicGameDataNode.RefereeNode.Referees)
	body += "\n"
	return body
}

func findTeamWithTricode(tricode TriCode, teams []Team) Team {
	for _, t := range teams {
		if t.TriCode == tricode {
			return t
		}
	}
	log.Println("cannot find team in teams slice with tricode: " + tricode)
	return Team{}
}

func (b *BoxscoreOld) GetRedditGameThreadTitle(teamTriCode TriCode, teams []Team) string {
	title := ""
	firstTeam := Team{}
	firstTeamInfo := TeamBoxscoreInfo{}
	firstTeamPlayoffsGameTeamInfo := PlayoffsGameTeamInfo{}
	secondTeam := Team{}
	secondTeamInfo := TeamBoxscoreInfo{}
	secondTeamPlayoffsGameTeamInfo := PlayoffsGameTeamInfo{}
	if b.BasicGameDataNode.HomeTeamInfo.TriCode == teamTriCode {
		firstTeam = findTeamWithTricode(b.BasicGameDataNode.HomeTeamInfo.TriCode, teams)
		firstTeamInfo = b.BasicGameDataNode.HomeTeamInfo
		secondTeam = findTeamWithTricode(b.BasicGameDataNode.AwayTeamInfo.TriCode, teams)
		secondTeamInfo = b.BasicGameDataNode.AwayTeamInfo

		if b.IsPlayoffGame() {
			firstTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo
			secondTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo
		}
	} else {
		firstTeam = findTeamWithTricode(b.BasicGameDataNode.AwayTeamInfo.TriCode, teams)
		firstTeamInfo = b.BasicGameDataNode.AwayTeamInfo
		secondTeam = findTeamWithTricode(b.BasicGameDataNode.HomeTeamInfo.TriCode, teams)
		secondTeamInfo = b.BasicGameDataNode.HomeTeamInfo

		if b.IsPlayoffGame() {
			firstTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo
			secondTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo
		}
	}

	teamRecordString := "(%s-%s)"

	title += "[GAME THREAD]"
	title += " "

	if b.IsPlayoffGame() {
		playoffsRoundInt, err := strconv.Atoi(b.BasicGameDataNode.PlayoffsNode.Round)
		if err != nil {
			log.Fatal(fmt.Sprintf("could not convert playoff round %s to int", b.BasicGameDataNode.PlayoffsNode.Round))
		}

		if playoffsRoundInt == 2 {
			title += fmt.Sprintf("%sERN CONF SEMIS", strings.ToUpper(b.BasicGameDataNode.PlayoffsNode.Conference))
		} else if playoffsRoundInt == 3 {
			title += fmt.Sprintf("%sERN CONF FINALS", strings.ToUpper(b.BasicGameDataNode.PlayoffsNode.Conference))
		} else if playoffsRoundInt == 4 {
			title += "NBA FINALS"
		} else {
			title += fmt.Sprintf("Playoffs Round %d", playoffsRoundInt)
		}
		title += ":"
		title += " "
		title += fmt.Sprintf("(%s)", firstTeamPlayoffsGameTeamInfo.Seed)
		title += " "
	}

	title += strings.ToUpper(firstTeam.Nickname)
	title += " "

	if !b.IsPlayoffGame() {
		title += fmt.Sprintf(teamRecordString, firstTeamInfo.Wins, firstTeamInfo.Losses)
		title += " "
	}

	title += "TAKE ON THE"
	title += " "

	if b.IsPlayoffGame() {
		title += fmt.Sprintf("(%s)", secondTeamPlayoffsGameTeamInfo.Seed)
		title += " "
	}

	title += strings.ToUpper(secondTeam.Nickname)

	if !b.IsPlayoffGame() {
		title += " "
		title += fmt.Sprintf(teamRecordString, secondTeamInfo.Wins, secondTeamInfo.Losses)
	}

	if b.IsPlayoffGame() {
		title += ","
		title += " "

		// Playoff series info
		log.Println(fmt.Sprintf("%s: %s", firstTeamInfo.TriCode, firstTeamPlayoffsGameTeamInfo.SeriesWins))
		log.Println(fmt.Sprintf("%s: %s", secondTeamInfo.TriCode, secondTeamPlayoffsGameTeamInfo.SeriesWins))

		firstTeamWins, err := strconv.Atoi(firstTeamPlayoffsGameTeamInfo.SeriesWins)
		if err != nil {
			log.Fatal("failed to convert first team's series wins to int")
		}

		secondTeamWins, err := strconv.Atoi(secondTeamPlayoffsGameTeamInfo.SeriesWins)
		if err != nil {
			log.Fatal("failed to convert second team's series wins to int")
		}

		if firstTeamWins == secondTeamWins {
			title += "SERIES TIED"
		} else if firstTeamWins < secondTeamWins {
			title += "TRAIL SERIES"
		} else {
			// first team leading series
			title += "LEAD SERIES"
		}
		title += " "
		title += "(" + firstTeamPlayoffsGameTeamInfo.SeriesWins + "-" + secondTeamPlayoffsGameTeamInfo.SeriesWins + ")"
	}

	return title
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

func (b *Boxscore) Final() bool {
	return b.GameNode.GameStatus == 3
}

// old boxscore exists for every game current and scheduled with a gameId
func GetCurrentSeasonBoxscore(ctx context.Context, r2Client cloudflare.Client, bucket string, gameID, gameDate string) (BoxscoreOld, error) {
	dailyAPIPaths, err := GetDailyAPIPaths()
	if err != nil {
		return BoxscoreOld{}, err
	}
	seasonStartYear := dailyAPIPaths.APISeasonInfoNode.SeasonYear
	return GetOldBoxscore(ctx, r2Client, bucket, gameID, gameDate, seasonStartYear)
}

func GetOldBoxscore(ctx context.Context, r2Client cloudflare.Client, bucket string, gameID, gameDate string, seasonStartYear int) (BoxscoreOld, error) {
	filename := fmt.Sprintf(os.Getenv("STORAGE_PATH")+"/boxscoreold/%d/%s.json", seasonStartYear, gameID)

	// remove all -'s in the game date. YYYY-mm-dd is a proper format but the old boxscore url uses no -'s like YYYYmmdd
	url := fmt.Sprintf(OldBoxscoreURL, strings.ReplaceAll(gameDate, "-", ""), gameID)

	objectKey := fmt.Sprintf("boxscore/%d/%s_data", seasonStartYear, gameID)

	boxscore, err := fetchObjectFromFileOrURL[BoxscoreOld](ctx, r2Client, url, filename, bucket, objectKey, func(b BoxscoreOld) bool { return b.Final() })
	if err != nil {
		return BoxscoreOld{}, err
	}

	return boxscore, nil
}

// preferred: new boxscore with more stats, details, ids. however, returns error if game is in the future
func GetBoxscoreDetailed(ctx context.Context, r2Client cloudflare.Client, bucket string, gameID string, seasonStartYear int) (Boxscore, error) {
	filename := fmt.Sprintf(os.Getenv("STORAGE_PATH")+"/boxscore/%d/%s.json", seasonStartYear, gameID)

	url := fmt.Sprintf(BoxscoreURL, gameID)

	objectKey := fmt.Sprintf("boxscore/%d/%s_cdn", seasonStartYear, gameID)

	boxscore, err := fetchObjectFromFileOrURL[Boxscore](ctx, r2Client, url, filename, bucket, objectKey, func(b Boxscore) bool { return b.Final() })
	if err != nil {
		return Boxscore{}, err
	}

	return boxscore, nil
}

const boxscoreSummaryV2URL = "https://stats.nba.com/stats/boxscoresummaryv2?GameID=%d"

//func (c client) GetBoxscoreSummary(gameID int) (Boxscore, error) {
//	response, err := c.client.Get(fmt.Sprintf(boxscoreSummaryV2URL, gameID))
//
//	if response != nil {
//		defer response.Body.Close()
//	}
//
//	if err != nil {
//		return Boxscore{}, err
//	}
//
//	if response.StatusCode != 200 {
//		return Boxscore{}, fmt.Errorf("failed to get BoxscoreSummary object")
//	}
//
//	var respBody []byte
//	respBody, err = io.ReadAll(response.Body)
//
//	pbp, err := unmarshalNBAHttpResponseToJSON[PlayByPlay](bytes.NewReader(respBody))
//	if err != nil {
//		return {}, fmt.Errorf("failed to get PlayByPlay object")
//	}
//
//	for _, outputWriter := range outputWriters {
//		if err := outputWriter.Put(ctx, respBody); err != nil {
//			return PlayByPlay{}, fmt.Errorf("failed to write output for play by play for game: %w", err)
//		}
//	}
//
//	return pbp, nil
//}
