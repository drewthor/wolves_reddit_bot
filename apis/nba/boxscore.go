package nba

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Boxscore struct {
	StatsNode *struct {
		HomeTeamNode struct {
			TeamStats TeamStats `json:"totals"`
		} `json:"hTeam"`
		AwayTeamNode struct {
			TeamStats TeamStats `json:"totals"`
		} `json:"vTeam"`
		PlayersStats []PlayerStats `json:"activePlayers"`
	} `json:"stats,omitempty"`
	BasicGameDataNode struct {
		Arena                ArenaInfo         `json:"arena"`
		Attendance           string            `json:"attendance"`
		Clock                string            `json:"clock"`
		GameIsActivated      bool              `json:"isGameActivated"` // see UpdateTeamsRegularSeasonRecords
		GameStartTimeEastern string            `json:"startTimeEastern"`
		GameStartDateEastern string            `json:"startDateEastern"`
		GameEndTimeUTC       string            `json:"endTimeUTC,omitempty"`
		HomeTeamInfo         TeamBoxscoreInfo  `json:"hTeam"`
		AwayTeamInfo         TeamBoxscoreInfo  `json:"vTeam"`
		PlayoffsNode         *PlayoffsGameInfo `json:"playoffs,omitempty"`
		// season stage IDs
		// 1: preseason
		// 2: regular season
		// 3: playoffs
		SeasonStageID int `json:"seasonStageId"`

		PeriodNode struct {
			CurrentPeriod int `json:"current"`
		} `json:"period"`

		RefereeNode struct {
			Referees []RefereeInfo `json:"formatted"`
		} `json:"officials"`
	} `json:"basicGameData"`
}

func (b *Boxscore) IsPlayoffGame() bool {
	return b.BasicGameDataNode.SeasonStageID == 3 || b.BasicGameDataNode.PlayoffsNode != nil
}

func (b *Boxscore) GameEnded() bool {
	if b.StatsNode == nil {
		// the nba api will post a boxscore without the stats json node for some time before games
		return false
	}
	hasEndTime := b.BasicGameDataNode.GameEndTimeUTC != ""
	if hasEndTime {
		log.Println("endTimeUTC reported")
		return true
	}
	noTimeRemaining := b.BasicGameDataNode.Clock == "0.0" || b.BasicGameDataNode.Clock == ""
	gameNotTied := b.StatsNode.HomeTeamNode.TeamStats.Points != b.StatsNode.AwayTeamNode.TeamStats.Points
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

func (b *Boxscore) DurationUntilGameStarts() time.Duration {
	currentTimeUTC := time.Now().UTC()
	// Issues occur when using eastern time for "today's games" as games on the west coast can still be going on
	// when the eastern time rolls over into the next day
	eastCoastLocation, locationError := time.LoadLocation("America/New_York")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeEastern := currentTimeUTC.In(eastCoastLocation)

	gameTime := makeGoTimeFromAPIData(b.BasicGameDataNode.GameStartTimeEastern, b.BasicGameDataNode.GameStartDateEastern)

	return gameTime.Sub(currentTimeEastern)
}

func (b *Boxscore) GameStarted() bool {
	currentTimeUTC := time.Now().UTC()
	// Issues occur when using eastern time for "today's games" as games on the west coast can still be going on
	// when the eastern time rolls over into the next day
	eastCoastLocation, locationError := time.LoadLocation("America/New_York")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeEastern := currentTimeUTC.In(eastCoastLocation)

	gameTime := makeGoTimeFromAPIData(b.BasicGameDataNode.GameStartTimeEastern, b.BasicGameDataNode.GameStartDateEastern)

	if currentTimeEastern.After(gameTime) {
		return true
	}
	return false
}

func (b *Boxscore) GetOpponent(team TriCode) TriCode {
	if b.BasicGameDataNode.HomeTeamInfo.TriCode == team {
		return b.BasicGameDataNode.AwayTeamInfo.TriCode
	} else {
		return b.BasicGameDataNode.HomeTeamInfo.TriCode
	}
}

func incrementString(str string) string {
	// convert string to a number
	i, _ := strconv.Atoi(str)

	// add one to the number
	i = i + 1

	// convert number back to string
	str = strconv.FormatInt(int64(i), 10)

	return str
}

func (b *Boxscore) UpdateTeamsRegularSeasonRecords() {
	if b.IsPlayoffGame() {
		return
	}

	log.Println(fmt.Sprintf("GameIsActivated: %t", b.BasicGameDataNode.GameIsActivated))
	// the nba does not appear to update the series wins and losses right after the game for either team for regular reason series records; update them based on the result of the game
	// they do eventually update the series wins and losses, but by then we should have already posted the thread
	// isGameActivated might be the trigger/think to look at for if the series has been updated see https://github.com/f1uk3r/Some-Python-Scripts/blob/master/reddit-nba-bot/reddit-boxscore-bot.py
	// update: this does not appear to be reliable either
	if !b.BasicGameDataNode.GameIsActivated {
		return
	}

	log.Println("Updating team regular season records")

	homeTeamPoints, err := strconv.Atoi(b.StatsNode.HomeTeamNode.TeamStats.Points)
	if err != nil {
		log.Println("could not convert home regular season points to int")
	}
	awayTeamPoints, err := strconv.Atoi(b.StatsNode.AwayTeamNode.TeamStats.Points)
	if err != nil {
		log.Println("could not convert away regular season points to int")
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

func (b *Boxscore) UpdateTeamsPlayoffsSeriesRecords() {
	if !b.IsPlayoffGame() {
		return
	}

	log.Println("Updating team playoff records")

	homeWins, err := strconv.Atoi(b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.SeriesWins)
	if err != nil {
		log.Println("could not convert home playoff series wins to int")
	}
	awayWins, err := strconv.Atoi(b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.SeriesWins)
	if err != nil {
		log.Println("could not convert away playoff series wins to int")
	}
	gameInSeries, err := strconv.Atoi(b.BasicGameDataNode.PlayoffsNode.GameInSeries)
	if err != nil {
		log.Println("could not convert away playoff series wins to int")
	}
	log.Println(fmt.Sprintf("gameInSeries: %d", gameInSeries))
	if (homeWins + awayWins) != gameInSeries {
		log.Println("updating playoff series records")

		homeTeamPoints, err := strconv.Atoi(b.StatsNode.HomeTeamNode.TeamStats.Points)
		if err != nil {
			log.Println("could not convert home playoff points to int")
		}
		awayTeamPoints, err := strconv.Atoi(b.StatsNode.AwayTeamNode.TeamStats.Points)
		if err != nil {
			log.Println("could not convert away playoff points to int")
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
			log.Println("failed to convert points to string")
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
			log.Println("failed to convert rebounds to string")
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
			log.Println("failed to convert assists to string")
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
			log.Println("failed to convert blocks to string")
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
			log.Println("failed to convert steals to string")
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

	gameTimeEastern := makeGoTimeFromAPIData(startTimeEastern, startDateEastern)
	centralLocation, locationErr := time.LoadLocation("America/Chicago")
	if locationErr != nil {
		log.Println("Failed to load Minneapolis location")
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

func getTeamQuarterScoreTableString(homeTeamBoxscoreInfo TeamBoxscoreInfo, homeTeamStats TeamStats, awayTeamBoxscoreInfo TeamBoxscoreInfo, awayTeamStats TeamStats) string {
	quarterScoreTableString := ""
	quarterScoreTableString += "||**Q1**|**Q2**|**Q3**|**Q4**|**Total**|\n"
	quarterScoreTableString += "|:-:|:-:|:-:|:-:|:-:|:-:|\n"
	quarterScoreTableString += fmt.Sprintf("|%s|%s|%s|%s|%s|%s|\n", homeTeamBoxscoreInfo.TriCode, homeTeamBoxscoreInfo.PointsByQuarter[0].Points, homeTeamBoxscoreInfo.PointsByQuarter[1].Points, homeTeamBoxscoreInfo.PointsByQuarter[2].Points, homeTeamBoxscoreInfo.PointsByQuarter[3].Points, homeTeamStats.Points)
	quarterScoreTableString += fmt.Sprintf("|%s|%s|%s|%s|%s|%s|\n", awayTeamBoxscoreInfo.TriCode, awayTeamBoxscoreInfo.PointsByQuarter[0].Points, awayTeamBoxscoreInfo.PointsByQuarter[1].Points, awayTeamBoxscoreInfo.PointsByQuarter[2].Points, awayTeamBoxscoreInfo.PointsByQuarter[3].Points, awayTeamStats.Points)
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

func getTeamStatsTableString(teamBoxscoreInfo TeamBoxscoreInfo, teamStats TeamStats, players map[string]Player, playersStats []PlayerStats) string {
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

func (b *Boxscore) GetRedditPostGameThreadBodyString(players map[string]Player, gameThreadURL string) string {
	body := ""
	body += getGameInfoTableString(b.BasicGameDataNode.Arena.Name, b.BasicGameDataNode.Arena.City, b.BasicGameDataNode.Arena.Country, b.BasicGameDataNode.GameStartTimeEastern, b.BasicGameDataNode.GameStartDateEastern, b.BasicGameDataNode.Attendance, false /*gameThread*/, gameThreadURL)
	body += "\n"
	body += getTeamQuarterScoreTableString(b.BasicGameDataNode.HomeTeamInfo, b.StatsNode.HomeTeamNode.TeamStats, b.BasicGameDataNode.AwayTeamInfo, b.StatsNode.AwayTeamNode.TeamStats)
	body += "\n"

	homeTeamLeaders, awayTeamLeaders := getTeamLeaders(b.StatsNode.PlayersStats, b.BasicGameDataNode.HomeTeamInfo.TeamID, b.BasicGameDataNode.AwayTeamInfo.TeamID)
	body += getTeamLeadersTableString(b.BasicGameDataNode.HomeTeamInfo, homeTeamLeaders, b.BasicGameDataNode.AwayTeamInfo, awayTeamLeaders, players)
	body += "\n"

	body += getTeamStatsTableString(b.BasicGameDataNode.HomeTeamInfo, b.StatsNode.HomeTeamNode.TeamStats, players, b.StatsNode.PlayersStats)
	body += "\n"
	body += getTeamStatsTableString(b.BasicGameDataNode.AwayTeamInfo, b.StatsNode.AwayTeamNode.TeamStats, players, b.StatsNode.PlayersStats)
	body += "\n"
	body += getRefereeTableString(b.BasicGameDataNode.RefereeNode.Referees)
	body += "\n"
	return body
}

func (b *Boxscore) GetRedditPostGameThreadTitle(teamTriCode TriCode, teams map[TriCode]Team) string {
	title := ""
	firstTeam := Team{}
	firstTeamStats := TeamStats{}
	firstTeamInfo := TeamBoxscoreInfo{}
	firstTeamPlayoffsGameTeamInfo := PlayoffsGameTeamInfo{}
	secondTeam := Team{}
	secondTeamStats := TeamStats{}
	secondTeamInfo := TeamBoxscoreInfo{}
	secondTeamPlayoffsGameTeamInfo := PlayoffsGameTeamInfo{}
	if b.BasicGameDataNode.HomeTeamInfo.TriCode == teamTriCode {
		firstTeam = teams[b.BasicGameDataNode.HomeTeamInfo.TriCode]
		firstTeamStats = b.StatsNode.HomeTeamNode.TeamStats
		firstTeamInfo = b.BasicGameDataNode.HomeTeamInfo
		secondTeam = teams[b.BasicGameDataNode.AwayTeamInfo.TriCode]
		secondTeamStats = b.StatsNode.AwayTeamNode.TeamStats
		secondTeamInfo = b.BasicGameDataNode.AwayTeamInfo

		if b.IsPlayoffGame() {
			firstTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo
			secondTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo
		}
	} else {
		firstTeam = teams[b.BasicGameDataNode.AwayTeamInfo.TriCode]
		firstTeamStats = b.StatsNode.AwayTeamNode.TeamStats
		firstTeamInfo = b.BasicGameDataNode.AwayTeamInfo
		secondTeam = teams[b.BasicGameDataNode.HomeTeamInfo.TriCode]
		secondTeamStats = b.StatsNode.HomeTeamNode.TeamStats
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
			log.Println("failed to convert first team's series wins to int")
		}

		secondTeamWins, err := strconv.Atoi(secondTeamPlayoffsGameTeamInfo.SeriesWins)
		if err != nil {
			log.Println("failed to convert second team's series wins to int")
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

func (b *Boxscore) GetRedditGameThreadBodyString(players map[string]Player, postGameThreadURL string) string {
	body := ""
	body += getGameInfoTableString(b.BasicGameDataNode.Arena.Name, b.BasicGameDataNode.Arena.City, b.BasicGameDataNode.Arena.Country, b.BasicGameDataNode.GameStartTimeEastern, b.BasicGameDataNode.GameStartDateEastern, b.BasicGameDataNode.Attendance, true /*gameThread*/, postGameThreadURL)
	body += "\n"

	if b.StatsNode != nil {
		if len(b.BasicGameDataNode.HomeTeamInfo.PointsByQuarter) > 0 && len(b.BasicGameDataNode.AwayTeamInfo.PointsByQuarter) > 0 {
			body += getTeamQuarterScoreTableString(b.BasicGameDataNode.HomeTeamInfo, b.StatsNode.HomeTeamNode.TeamStats, b.BasicGameDataNode.AwayTeamInfo, b.StatsNode.AwayTeamNode.TeamStats)
			body += "\n"
		}

		homeTeamLeaders, awayTeamLeaders := getTeamLeaders(b.StatsNode.PlayersStats, b.BasicGameDataNode.HomeTeamInfo.TeamID, b.BasicGameDataNode.AwayTeamInfo.TeamID)
		body += getTeamLeadersTableString(b.BasicGameDataNode.HomeTeamInfo, homeTeamLeaders, b.BasicGameDataNode.AwayTeamInfo, awayTeamLeaders, players)
		body += "\n"

		body += getTeamStatsTableString(b.BasicGameDataNode.HomeTeamInfo, b.StatsNode.HomeTeamNode.TeamStats, players, b.StatsNode.PlayersStats)
		body += "\n"
		body += getTeamStatsTableString(b.BasicGameDataNode.AwayTeamInfo, b.StatsNode.AwayTeamNode.TeamStats, players, b.StatsNode.PlayersStats)
		body += "\n"
	}

	body += getRefereeTableString(b.BasicGameDataNode.RefereeNode.Referees)
	body += "\n"
	return body
}

func (b *Boxscore) GetRedditGameThreadTitle(teamTriCode TriCode, teams map[TriCode]Team) string {
	title := ""
	firstTeam := Team{}
	firstTeamInfo := TeamBoxscoreInfo{}
	firstTeamPlayoffsGameTeamInfo := PlayoffsGameTeamInfo{}
	secondTeam := Team{}
	secondTeamInfo := TeamBoxscoreInfo{}
	secondTeamPlayoffsGameTeamInfo := PlayoffsGameTeamInfo{}
	if b.BasicGameDataNode.HomeTeamInfo.TriCode == teamTriCode {
		firstTeam = teams[b.BasicGameDataNode.HomeTeamInfo.TriCode]
		firstTeamInfo = b.BasicGameDataNode.HomeTeamInfo
		secondTeam = teams[b.BasicGameDataNode.AwayTeamInfo.TriCode]
		secondTeamInfo = b.BasicGameDataNode.AwayTeamInfo

		if b.IsPlayoffGame() {
			firstTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo
			secondTeamPlayoffsGameTeamInfo = b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo
		}
	} else {
		firstTeam = teams[b.BasicGameDataNode.AwayTeamInfo.TriCode]
		firstTeamInfo = b.BasicGameDataNode.AwayTeamInfo
		secondTeam = teams[b.BasicGameDataNode.HomeTeamInfo.TriCode]
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
			log.Println("failed to convert first team's series wins to int")
		}

		secondTeamWins, err := strconv.Atoi(secondTeamPlayoffsGameTeamInfo.SeriesWins)
		if err != nil {
			log.Println("failed to convert second team's series wins to int")
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
	PointsByQuarter []struct {
		Points string `json:"score"`
	} `json:"linescore"`
}

type TeamStats struct {
	Points               string      `json:"points"`
	Minutes              string      `json:"min"`
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

func GetBoxscore(boxscoreAPIPath, todaysDate string, gameID string) Boxscore {
	templateURI := makeURIFormattable(nbaAPIBaseURI + boxscoreAPIPath)
	url := fmt.Sprintf(templateURI, todaysDate, gameID)
	log.Println(url)
	response, httpErr := http.Get(url)

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if httpErr != nil {
		log.Fatal(httpErr)
	}

	boxscoreResult := Boxscore{}
	decodeErr := json.NewDecoder(response.Body).Decode(&boxscoreResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	if boxscoreResult.GameEnded() {
		log.Println("updating series record")
		boxscoreResult.UpdateTeamsPlayoffsSeriesRecords()
		boxscoreResult.UpdateTeamsRegularSeasonRecords()
	}
	return boxscoreResult
}
