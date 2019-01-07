package nba

import (
	"regexp"
)

const NBAAPIBaseURI = "http://data.nba.net/10s"

func MakeURIFormattable(uri string) string {
	regex := regexp.MustCompile(`{{.+?}}`)
	format := "%s"
	formattedString := regex.ReplaceAllString(uri, format)
	return formattedString
}
