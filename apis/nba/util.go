package nba

import (
	"regexp"
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
