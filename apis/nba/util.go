package nba

import (
	"fmt"
	"log"
	"regexp"
	"time"
)

const nbaAPIBaseURI = "http://data.nba.net"

// TimeDayFormat - Year/month/day format used by the NBA api
const TimeDayFormat = "20060102"

// UTCFormat - UTC format used by the NBA api
const UTCFormat = "2006-01-02T15:04:00.000Z"

func makeURIFormattable(uri string) string {
	regex := regexp.MustCompile(`{{.+?}}`)
	format := "%s"
	formattedString := regex.ReplaceAllString(uri, format)
	return formattedString
}

func makeGoTimeFromAPIData(startTimeEastern, startDateEastern string) time.Time {
	eastCoastLocation, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Println(err)
	}

	// add space between time zone and year to help parser
	APIFormat := "3:04 PM 20060102"

	// strip out time zone from start time as the time zone is eastern US
	// and doesn't match golang's time package time zones (e.g. ET vs golang expects EST)
	re := regexp.MustCompile(`([^ET])*`)
	matches := re.FindStringSubmatch(startTimeEastern)

	// grab the first match since the NBA time string puts the time zone last
	combinedAPIData := matches[0] + startDateEastern
	time, err := time.ParseInLocation(APIFormat, combinedAPIData, eastCoastLocation)
	if err != nil {
		log.Println(fmt.Sprintf("combined API game time: %s", combinedAPIData))
		log.Println(err)
	}

	return time
}

// returns a player's string of the form "D. Howard"
func getPlayerString(playerID string, players map[string]Player) string {
	playerString := "%s. %s"
	player := players[playerID]
	firstInitial := ""
	lastName := ""
	if player.FirstName == "" {
		log.Println(player)
		log.Println(playerID)
		firstInitial = ""
		lastName = ""
	} else {
		firstInitial = player.FirstName[:1]
		lastName = player.LastName
	}
	return fmt.Sprintf(playerString, firstInitial, lastName)
}

type TriCode string

const (
	AtlantaHawks          TriCode = "ATL"
	BostonCeltics         TriCode = "BOS"
	BrooklynNets          TriCode = "BKN"
	CharlotteHornets      TriCode = "CHA"
	ChicagoBulls          TriCode = "CHI"
	ClevelandCavaliers    TriCode = "CLE"
	DallasMavericks       TriCode = "DAL"
	DenverNuggets         TriCode = "DEN"
	DetroitPistons        TriCode = "DET"
	GoldenStateWarriors   TriCode = "GSW"
	HoustonRockets        TriCode = "HOU"
	IndianaPacers         TriCode = "IND"
	LosAngelesClippers    TriCode = "LAC"
	LosAngelesLakers      TriCode = "LAL"
	MemphisGrizzlies      TriCode = "MEM"
	MiamiHeat             TriCode = "MIA"
	MilwaukeeBucks        TriCode = "MIL"
	MinnesotaTimberwolves TriCode = "MIN"
	NewOrleansPelicans    TriCode = "NOP"
	NewYorkKnicks         TriCode = "NYK"
	OklahomaCityThunder   TriCode = "OKC"
	OrlandoMagic          TriCode = "ORL"
	Philadelphia76ers     TriCode = "PHI"
	PhoenixSuns           TriCode = "PHX"
	PortlandTrailblazers  TriCode = "POR"
	SacramentoKings       TriCode = "SAC"
	SanAntonioSpurs       TriCode = "SAS"
	TorontoRaptors        TriCode = "TOR"
	UtahJazz              TriCode = "UTA"
	WashingtonWizards     TriCode = "WAS"
)
