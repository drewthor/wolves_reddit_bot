package nba

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/util"
	log "github.com/sirupsen/logrus"
)

const nbaAPIBaseURI = "http://data.nba.net"

// TimeDayFormat - Year/month/day format used by the NBA api
const TimeDayFormat = "20060102"

// TimeBirthdateFormat - yyyy-mm-dd format used by players api response
const TimeBirthdateFormat = "2006-01-02"

// UTCFormat - UTC format used by the NBA api
const UTCFormat = "2006-01-02T15:04:00.000Z"

func makeURIFormattable(uri string) string {
	regex := regexp.MustCompile(`{{.+?}}`)
	format := "%s"
	formattedString := regex.ReplaceAllString(uri, format)
	return formattedString
}

func NormalizeGameDate(gameDate string) string {
	return strings.ReplaceAll(gameDate, "-", "")
}

func makeGoTimeFromAPIData(startTimeEastern, startDateEastern string) (time.Time, error) {
	eastCoastLocation, err := time.LoadLocation("America/New_York")
	if err != nil {
		errorMessage := "failed to load new york time location"
		log.WithError(err).Error(errorMessage)
		return time.Time{}, fmt.Errorf(errorMessage+" %w", err)
	}

	// add space between time zone and year to help parser
	APIFormat := "3:04 PM 20060102"

	// strip out time zone from start time as the time zone is eastern US
	// and doesn't match golang's time package time zones (e.g. ET vs golang expects EST)
	re := regexp.MustCompile(`([^ET])*`)
	matches := re.FindStringSubmatch(startTimeEastern)

	// grab the first match since the NBA time string puts the time zone last
	combinedAPIData := matches[0] + startDateEastern
	parsedTime, err := time.ParseInLocation(APIFormat, combinedAPIData, eastCoastLocation)
	if err != nil {
		log.Infof("combined API game time: %s", combinedAPIData)
		errorMessage := fmt.Sprintf("failed to parse combined API game time: %s in new york time location", combinedAPIData)
		log.WithError(err).Error(errorMessage)
		return time.Time{}, fmt.Errorf(errorMessage+" %w", err)
	}

	return parsedTime, nil
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
		errorMessage := "could not decode json body to raw json"
		log.WithError(err).Error(errorMessage)
		return t, fmt.Errorf(errorMessage+": %w", err)
	}
	err = json.Unmarshal(raw, &t)
	if err != nil {
		errorMessage := fmt.Sprintf("could not decode json to type %T raw: %s", t, raw)
		log.WithError(err).Error(errorMessage)
		return t, fmt.Errorf(errorMessage+": %w", err)
	}

	return t, nil
}

func fetchObjectAndSaveToFile[T any](ctx context.Context, r2Client cloudflare.Client, url string, filePath string, bucket string, objectKey string) (T, error) {
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, 0770); err != nil {
		return *new(T), err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return *new(T), err
	}

	defer file.Close()

	obj := *new(T)

	response, err := http.Get(url)

	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return *new(T), err
	}

	if response.StatusCode != 200 {
		return *new(T), fmt.Errorf("failed to get object of type %T from url %s with status code %d", *new(T), url, response.StatusCode)
	}

	var respBody []byte
	respBody, err = io.ReadAll(response.Body)
	obj, err = unmarshalNBAHttpResponseToJSON[T](bytes.NewReader(respBody))
	if err != nil {
		return *new(T), fmt.Errorf("failed to get object of type %T from url %s with status code %d", *new(T), url, response.StatusCode)
	}

	if err := r2Client.CreateObject(ctx, bucket, objectKey, util.ContentTypeJSON, bytes.NewReader(respBody)); err != nil {
		log.WithError(err).WithFields(log.Fields{"bucket": bucket, "object_key": objectKey}).Error("failed to write object to r2 bucket")
	}

	n, err := file.Write(respBody)
	if err != nil {
		return *new(T), err
	}
	if n == 0 {
		return *new(T), fmt.Errorf("wrote nothing to file %s", filePath)
	}

	return obj, nil
}

func fetchObjectFromFileOrURL[T any](ctx context.Context, r2Client cloudflare.Client, url string, filePath string, bucket string, objectKey string, loadFromFileFunc func(T) bool) (T, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0770); err != nil {
		return *new(T), err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return *new(T), err
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return *new(T), err
	}

	reqBody := []byte{}
	obj := *new(T)

	loadFromFile := false
	if stat.Size() > 0 {
		reqBody, err = io.ReadAll(file)

		err = json.NewDecoder(bytes.NewReader(reqBody)).Decode(&obj)
		if err == nil {
			loadFromFile = loadFromFileFunc(obj)
		}

		if err := r2Client.CreateObject(ctx, bucket, objectKey, util.ContentTypeJSON, bytes.NewReader(reqBody)); err != nil {
			log.WithError(err).WithFields(log.Fields{"bucket": bucket, "object_key": objectKey}).Error("failed to write object to r2 bucket")
		}
	}

	if stat.Size() <= 0 || loadFromFile {
		response, err := http.Get(url)

		if response != nil {
			defer response.Body.Close()
		}

		if err != nil {
			return *new(T), err
		}

		if response.StatusCode != 200 {
			return *new(T), fmt.Errorf("failed to get object of type %T from url %s with status code %d", *new(T), url, response.StatusCode)
		}

		reqBody, err = io.ReadAll(response.Body)
		obj, err = unmarshalNBAHttpResponseToJSON[T](bytes.NewReader(reqBody))
		if err != nil {
			return *new(T), fmt.Errorf("failed to get object of type %T from url %s with status code %d", *new(T), url, response.StatusCode)
		}

		n, err := file.Write(reqBody)
		if err != nil {
			return *new(T), err
		}
		if n == 0 {
			return *new(T), fmt.Errorf("wrote nothing to file %s", filePath)
		}

		if err := r2Client.CreateObject(ctx, bucket, objectKey, util.ContentTypeJSON, bytes.NewReader(reqBody)); err != nil {
			log.WithError(err).WithFields(log.Fields{"bucket": bucket, "object_key": objectKey}).Error("failed to write object to r2 bucket")
		}
	}

	return obj, nil
}
