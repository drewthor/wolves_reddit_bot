package nba

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Boxscore struct {
	StatsNode struct {
		HomeTeamNode struct {
			TeamStats TeamStats `json:"totals"`
		} `json:"hTeam"`
		AwayTeamNode struct {
			TeamStats TeamStats `json:"totals"`
		} `json:"vTeam"`
		PlayersStats []PlayerStats `json:"activePlayers"`
	} `json:"stats"`
	HomeTeam TeamBoxscoreInfo `json:"hTeam"`
	AwayTeam TeamBoxscoreInfo `json:"vTeam"`
}

func get_team_stats_table_string(teamBoxscoreInfo TeamBoxscoreInfo, teamStats TeamStats, players map[string]Player, playersStats []PlayerStats) string {
	columnHeader := "**[](/%s) %s**|**Min**|**FG**|**FT**|**3PT**|**+/-**|**OR**|**Reb**|**A**|**Blk**|**Stl**|**TO**|**PF**|**Pts**|\n"
	columnHeaderSeparator := "|:---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n"
	playerStatsString := "%s. %s|%s|%s-%s|%s-%s|%s-%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|\n"
	totalsString := "Totals|%s|%s-%s(%s%%)|%s-%s(%s%%)|%s-%s(%s%%)|-|%s|%s|%s|%s|%s|%s|%s|%s|\n"
	teamStatsTableString := ""
	teamStatsTableString += fmt.Sprintf(columnHeader, teamBoxscoreInfo.AbbreviatedName, teamBoxscoreInfo.AbbreviatedName)
	teamStatsTableString += columnHeaderSeparator
	for _, playerStats := range playersStats {
		if playerStats.TeamID == teamBoxscoreInfo.TeamID {
			player := players[playerStats.ID]
			teamStatsTableString += fmt.Sprintf(playerStatsString, player.FirstName[1], player.LastName, playerStats.Minutes, playerStats.FieldGoalsMade, playerStats.FieldGoalsAttempted, playerStats.FreeThrowsMade, playerStats.FreeThrowsAttempted, playerStats.ThreePointsMade, playerStats.ThreePointsAttempted, playerStats.PlusMinus, playerStats.OffensiveRebounds, playerStats.TotalRebounds, playerStats.Assists, playerStats.Blocks, playerStats.Steals, playerStats.Turnovers, playerStats.PersonalFouls, playerStats.Points)
		}
	}
	teamStatsTableString += fmt.Sprintf(totalsString, teamStats.Minutes, teamStats.FieldGoalsMade, teamStats.FieldGoalsAttempted, teamStats.FieldGoalPercentage, teamStats.FreeThrowsMade, teamStats.FreeThrowsAttempted, teamStats.FreeThrowPercentage, teamStats.ThreePointsMade, teamStats.ThreePointsAttempted, teamStats.ThreePointPercentage, teamStats.OffensiveRebounds, teamStats.TotalRebounds, teamStats.Assists, teamStats.Blocks, teamStats.Steals, teamStats.Turnovers, teamStats.PersonalFouls, teamStats.Points)
	return teamStatsTableString
}

func (b *Boxscore) GetRedditBodyString(players map[string]Player) string {
	bodyString := ""
	bodyString += get_team_stats_table_string(b.HomeTeam, b.StatsNode.HomeTeamNode.TeamStats, players, b.StatsNode.PlayersStats)
	bodyString += "\n"
	bodyString += get_team_stats_table_string(b.AwayTeam, b.StatsNode.AwayTeamNode.TeamStats, players, b.StatsNode.PlayersStats)
	bodyString += "\n"
	bodyString += "||"
	bodyString += "|:-:|"
	return bodyString
}

type TeamBoxscoreInfo struct {
	TeamID          string `json:"teamId"`
	AbbreviatedName string `json:"triCode"`
	Wins            string `json:"win"`
	Loss            string `json:"loss"`
	SeriesWin       string `json:"seriesWin"`
	SeriesLoss      string `json:"seriesLoss"`
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
	return boxscoreResult
}
