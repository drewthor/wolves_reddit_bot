package nba

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// TimeDayFormat - Year/month/day format used by the NBA api
const TimeDayFormat = "20060102"

// TimeBirthdateFormat - yyyy-mm-dd format used by players api response
const TimeBirthdateFormat = "2006-01-02"

// UTCFormat - UTC format used by the NBA api
const UTCFormat = "2006-01-02T15:04:00.000Z"

const nbaStatsTimestampFormat = "2006-01-02T15:04:00"

func makeURIFormattable(uri string) string {
	regex := regexp.MustCompile(`{{.+?}}`)
	format := "%s"
	formattedString := regex.ReplaceAllString(uri, format)
	return formattedString
}

// returns a player's string of the form "D. Howard"
func getPlayerString(playerID string, players map[string]Player) string {
	playerString := "%s. %s"
	player := players[playerID]
	firstInitial := ""
	lastName := ""
	if player.FirstName == "" {
		firstInitial = ""
		lastName = ""
	} else {
		firstInitial = player.FirstName[:1]
		lastName = player.LastName
	}
	return fmt.Sprintf(playerString, firstInitial, lastName)
}

func parseSeasonStartEndYears(seasonYear string) (int, int, error) {
	years := strings.Split(seasonYear, "-")
	if len(years) != 2 {
		return -1, -1, fmt.Errorf("invalid season year expected format yyyy-yy and got %s", seasonYear)
	}

	startYear, err := strconv.Atoi(years[0])
	if err != nil {
		return -1, -1, fmt.Errorf("failed to get season start year from %s: %w", seasonYear, err)
	}
	// if start year and end year have the same last 2, then just prepend the first 2 from start to end, otherwise look at the next year's first 2 and prepend that
	// to end 2 to handle cases like 1999-99 and 1999-00 for the two mentioned cases
	endYearLast2, err := strconv.Atoi(years[1])
	if err != nil {
		return -1, -1, fmt.Errorf("failed to get season end year from %s: %w", seasonYear, err)
	}
	endYearFirst2 := (startYear + 1) / 100
	startYearLast2 := startYear % 100
	if startYearLast2 == endYearLast2 {
		endYearFirst2 = (startYear) / 100
	}
	endYear := endYearFirst2*100 + endYearLast2

	return startYear, endYear, nil
}

type seasonStage int

const (
	preSeason     seasonStage = 1
	regularSeason seasonStage = 2
	allStar       seasonStage = 3
	postSeason    seasonStage = 4
	playIn        seasonStage = 5
)

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

func unmarshalNBAHttpResponseToJSON[T any](reader io.Reader) (T, error) {
	t := *new(T)

	raw := json.RawMessage{}
	err := json.NewDecoder(reader).Decode(&raw)
	if err != nil {
		return t, fmt.Errorf("could not decode json body to raw json: %w", err)
	}
	err = json.Unmarshal(raw, &t)
	if err != nil {
		return t, fmt.Errorf("could not decode json to type %T raw: %s: %w", t, raw, err)
	}

	return t, nil
}
