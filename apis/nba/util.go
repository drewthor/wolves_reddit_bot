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
