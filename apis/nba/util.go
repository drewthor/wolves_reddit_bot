package nba

import (
	"regexp"
)

const NBAAPIBaseURI = "http://data.nba.net"

const TimeDayFormat = "20060102"

const UTCFormat = "2006-01-02T15:04:00.000Z"

func MakeURIFormattable(uri string) string {
	regex := regexp.MustCompile(`{{.+?}}`)
	format := "%s"
	formattedString := regex.ReplaceAllString(uri, format)
	return formattedString
}
