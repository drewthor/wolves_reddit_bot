package nba

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
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
		Clock           string            `json:"clock"`
		GameIsActivated bool              `json:"isGameActivated"` // if this is true, it might be able to be used to determine if things like record and series record is updated
		GameEndTimeUTC  string            `json:"endTimeUTC,omitempty"`
		HomeTeamInfo    TeamBoxscoreInfo  `json:"hTeam"`
		AwayTeamInfo    TeamBoxscoreInfo  `json:"vTeam"`
		PlayoffsNode    *PlayoffsGameInfo `json:"playoffs"`

		PeriodNode struct {
			CurrentPeriod int `json:"current"`
		} `json:"period"`
	} `json:"basicGameData"`
}

func (b *Boxscore) IsPlayoffGame() bool {
	return b.BasicGameDataNode.PlayoffsNode != nil
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

func (b *Boxscore) UpdateSeriesRecord() {
	// the nba does not appear to update the series wins and losses right after the game for either team; update them based on the result of the game
	// they do eventually update the series wins and losses, but by then we should have already posted the thread
	// isGameActivated might be the trigger/think to look at for if the series has been updated see https://github.com/f1uk3r/Some-Python-Scripts/blob/master/reddit-nba-bot/reddit-boxscore-bot.py
	if !b.BasicGameDataNode.GameIsActivated {
		return
	}

	homeTeamWon := b.StatsNode.HomeTeamNode.TeamStats.Points > b.StatsNode.AwayTeamNode.TeamStats.Points
	if homeTeamWon {
		b.BasicGameDataNode.HomeTeamInfo.SeriesWins = incrementString(b.BasicGameDataNode.HomeTeamInfo.SeriesWins)
		b.BasicGameDataNode.AwayTeamInfo.SeriesLosses = incrementString(b.BasicGameDataNode.AwayTeamInfo.SeriesLosses)

		if b.IsPlayoffGame() {
			b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.SeriesWins = incrementString(b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.SeriesWins)
			if b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.SeriesWins == "4" {
				b.BasicGameDataNode.PlayoffsNode.HomeTeamInfo.WonSeries = true
			}
		}
	} else {
		b.BasicGameDataNode.HomeTeamInfo.SeriesLosses = incrementString(b.BasicGameDataNode.HomeTeamInfo.SeriesLosses)
		b.BasicGameDataNode.AwayTeamInfo.SeriesWins = incrementString(b.BasicGameDataNode.AwayTeamInfo.SeriesWins)

		if b.IsPlayoffGame() {
			b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.SeriesWins = incrementString(b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.SeriesWins)
			if b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.SeriesWins == "4" {
				b.BasicGameDataNode.PlayoffsNode.AwayTeamInfo.WonSeries = true
			}
		}
	}
}

func get_team_stats_table_string(teamBoxscoreInfo TeamBoxscoreInfo, teamStats TeamStats, players map[string]Player, playersStats []PlayerStats) string {
	columnHeader := "**[](/%s) %s**|**Min**|**FG**|**FT**|**3PT**|**+/-**|**OR**|**Reb**|**A**|**Blk**|**Stl**|**TO**|**PF**|**Pts**|\n"
	columnHeaderSeparator := "|:---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n"
	playerStatsString := "%s. %s|%s|%s-%s|%s-%s|%s-%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|\n"
	totalsString := "Totals|%s|%s-%s(%s%%)|%s-%s(%s%%)|%s-%s(%s%%)|-|%s|%s|%s|%s|%s|%s|%s|%s|\n"
	teamStatsTableString := ""
	teamStatsTableString += fmt.Sprintf(columnHeader, teamBoxscoreInfo.TriCode, teamBoxscoreInfo.TriCode)
	teamStatsTableString += columnHeaderSeparator
	for _, playerStats := range playersStats {
		if playerStats.TeamID == teamBoxscoreInfo.TeamID {
			player := players[playerStats.ID]
			firstInitial := ""
			lastName := ""
			if player.FirstName == "" {
				log.Println(player)
				log.Println(playerStats.ID)
				log.Println(playerStats)
				firstInitial = ""
				lastName = ""
			} else {
				firstInitial = player.FirstName[:1]
				lastName = player.LastName
			}
			teamStatsTableString += fmt.Sprintf(playerStatsString, firstInitial, lastName, playerStats.Minutes, playerStats.FieldGoalsMade, playerStats.FieldGoalsAttempted, playerStats.FreeThrowsMade, playerStats.FreeThrowsAttempted, playerStats.ThreePointsMade, playerStats.ThreePointsAttempted, playerStats.PlusMinus, playerStats.OffensiveRebounds, playerStats.TotalRebounds, playerStats.Assists, playerStats.Blocks, playerStats.Steals, playerStats.Turnovers, playerStats.PersonalFouls, playerStats.Points)
		}
	}
	teamStatsTableString += fmt.Sprintf(totalsString, teamStats.Minutes, teamStats.FieldGoalsMade, teamStats.FieldGoalsAttempted, teamStats.FieldGoalPercentage, teamStats.FreeThrowsMade, teamStats.FreeThrowsAttempted, teamStats.FreeThrowPercentage, teamStats.ThreePointsMade, teamStats.ThreePointsAttempted, teamStats.ThreePointPercentage, teamStats.OffensiveRebounds, teamStats.TotalRebounds, teamStats.Assists, teamStats.Blocks, teamStats.Steals, teamStats.Turnovers, teamStats.PersonalFouls, teamStats.Points)
	return teamStatsTableString
}

func (b *Boxscore) GetRedditBodyString(players map[string]Player) string {
	body := ""
	body += get_team_stats_table_string(b.BasicGameDataNode.HomeTeamInfo, b.StatsNode.HomeTeamNode.TeamStats, players, b.StatsNode.PlayersStats)
	body += "\n"
	body += get_team_stats_table_string(b.BasicGameDataNode.AwayTeamInfo, b.StatsNode.AwayTeamNode.TeamStats, players, b.StatsNode.PlayersStats)
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
			title += fmt.Sprintf("%sERN CONF SEMIS", b.BasicGameDataNode.PlayoffsNode.Conference)
		} else if playoffsRoundInt == 3 {
			title += fmt.Sprintf("%sERN CONF FINALS", b.BasicGameDataNode.PlayoffsNode.Conference)
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
	title += ","
	title += " "

	if b.IsPlayoffGame() {
		// Playoff series info
		log.Println(fmt.Sprintf("%s: %s", firstTeamInfo.TriCode, firstTeamPlayoffsGameTeamInfo.SeriesWins))
		log.Println(fmt.Sprintf("%s: %s", secondTeamInfo.TriCode, secondTeamPlayoffsGameTeamInfo.SeriesWins))
		if firstTeamPlayoffsGameTeamInfo.SeriesWins == secondTeamPlayoffsGameTeamInfo.SeriesWins {
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
		title += "(" + firstTeamInfo.SeriesWins + "-" + firstTeamInfo.SeriesLosses + ")"
	}

	return title
}

type TeamBoxscoreInfo struct {
	TeamID       string  `json:"teamId"`
	TriCode      TriCode `json:"triCode"`
	Wins         string  `json:"win"`
	Losses       string  `json:"loss"`
	SeriesWins   string  `json:"seriesWin"`
	SeriesLosses string  `json:"seriesLoss"`
}

type TeamStats struct {
	Points               string `json:"points"`
	Minutes              string `json:"min"`
	FieldGoalsMade       string `json:"fgm"`
	FieldGoalsAttempted  string `json:"fga"`
	FieldGoalPercentage  string `json:"fgp"`
	FreeThrowsMade       string `json:"ftm"`
	FreeThrowsAttempted  string `json:"fta"`
	FreeThrowPercentage  string `json:"ftp"`
	ThreePointsMade      string `json:"tpm"`
	ThreePointsAttempted string `json:"tpa"`
	ThreePointPercentage string `json:"tpp"`
	OffensiveRebounds    string `json:"offReb"`
	DefensiveRebounds    string `json:"defReb"`
	TotalRebounds        string `json:"totReb"`
	Assists              string `json:"assists"`
	PersonalFouls        string `json:"pfouls"`
	Steals               string `json:"steals"`
	Turnovers            string `json:"turnovers"`
	Blocks               string `json:"blocks"`
	PlusMinus            string `json:"plusMinus"`
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
}

func GetBoxscore(boxscoreAPIPath, todaysDate string, gameID string) Boxscore {
	templateURI := makeURIFormattable(nbaAPIBaseURI + boxscoreAPIPath)
	url := fmt.Sprintf(templateURI, todaysDate, gameID)
	log.Println(url)
	response, httpErr := http.Get(url)
	if httpErr != nil {
		log.Fatal(httpErr)
	}
	defer response.Body.Close()

	boxscoreResult := Boxscore{}
	decodeErr := json.NewDecoder(response.Body).Decode(&boxscoreResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	if boxscoreResult.GameEnded() {
		boxscoreResult.UpdateSeriesRecord()
	}
	return boxscoreResult
}
