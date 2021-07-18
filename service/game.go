package service

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/dao"
)

type GameService struct {
	GameDAO *dao.GameDAO

	ArenaService       *ArenaService
	GameRefereeService *GameRefereeService
	RefereeService     *RefereeService
	SeasonStageService *SeasonStageService
	TeamService        *TeamService
}

func (gs GameService) Get(gameDate string) (api.Games, error) {
	gameScoreboards := nba.GetGameScoreboards(gameDate)
	return api.Games{Games: gameScoreboards.Games}, nil
}

func (gs GameService) UpdateGames() ([]api.Game, error) {
	nbaGames, err := gs.getCurrentLeagueScheduleFromNBAAPI()
	if err != nil {
		return nil, err
	}

	results := make(chan nba.Boxscore, len(nbaGames))
	// boxscores := []nba.Boxscore{}

	wg := sync.WaitGroup{}

	worker := func(nbaGame nba.Game, results chan<- nba.Boxscore) {
		defer wg.Done()

		boxscore, err := nba.GetBoxscore(nba.GetDailyAPIPaths().Boxscore, nbaGame.StartDateEastern, nbaGame.GameID)
		if err != nil {
			log.Printf("could not retrieve boxscore for gameID %v on date %s\n", nbaGame.GameID, nbaGame.StartDateEastern)
			// continue
			return
		}

		// boxscores = append(boxscores, boxscore)
		results <- boxscore
	}

	for _, nbaGame := range nbaGames {
		wg.Add(1)
		go worker(nbaGame, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	arenaUpdatesMap := map[string]dao.ArenaUpdate{}
	refereeFullNameUpdateMap := map[string]dao.RefereeUpdate{}
	gameRefereeUpdates := []dao.GameRefereeUpdate{}

	gameStatusNameMappings := gs.nbaGameStatusNameMappings()
	seasonStageNameMappings := gs.SeasonStageService.NBASeasonStageNameMappings()

	gameUpdates := []dao.GameUpdate{}
	for boxscore := range results {
		periodTimeRemainingFloat, err := strconv.ParseFloat(boxscore.BasicGameDataNode.Clock, 64)
		if err != nil {
			// not really an error, the nba api returns an empty string if the clock has expired
			periodTimeRemainingFloat = 0
		}
		periodTimeRemainingTenthSeconds := int(periodTimeRemainingFloat * 10)

		attendance, err := strconv.Atoi(boxscore.BasicGameDataNode.Attendance)
		if err != nil {
			// nba api treats "" as 0
			attendance = 0
		}

		durationSeconds := new(int)

		if boxscore.BasicGameDataNode.GameDurationNode != nil {
			if hours, err := strconv.Atoi(boxscore.BasicGameDataNode.GameDurationNode.Hours); err == nil {
				if minutes, err := strconv.Atoi(boxscore.BasicGameDataNode.GameDurationNode.Minutes); err == nil {
					*durationSeconds = (hours * 60 * 60) + (minutes * 60)
				}
			}
		}

		// nba api treats "" as 0
		homeTeamPoints, err := strconv.Atoi(boxscore.BasicGameDataNode.HomeTeamInfo.Points)
		homeTeamPointsValid := true
		if err != nil {
			homeTeamPointsValid = false
		}

		// nba api treats "" as 0
		awayTeamPoints, err := strconv.Atoi(boxscore.BasicGameDataNode.AwayTeamInfo.Points)
		awayTeamPointsValid := true
		if err != nil {
			awayTeamPointsValid = false
		}

		for _, nbaReferee := range boxscore.BasicGameDataNode.RefereeNode.Referees {
			// nba uses a ' ' to separate the first and last name in the referees fullname
			splitName := strings.SplitN(nbaReferee.FullName, " ", 2)
			if splitName == nil || len(splitName) != 2 {
				log.Println("could not parse official: " + nbaReferee.FullName + " for nba game id: " + boxscore.BasicGameDataNode.GameID)
				continue
			}

			firstName := splitName[0]
			// nba has a ref fullname that puts two ' ' between first name and last name
			lastName := strings.TrimPrefix(splitName[1], " ")

			refereeUpdate := dao.RefereeUpdate{
				FirstName: firstName,
				LastName:  lastName,
			}
			refereeFullNameUpdateMap[nbaReferee.FullName] = refereeUpdate

			gameRefereeUpdate := dao.GameRefereeUpdate{
				NBAGameID: boxscore.BasicGameDataNode.GameID,
				FirstName: firstName,
				LastName:  lastName,
			}

			gameRefereeUpdates = append(gameRefereeUpdates, gameRefereeUpdate)
		}

		arenaUpdate := dao.ArenaUpdate{
			Name: boxscore.BasicGameDataNode.Arena.Name,
			City: sql.NullString{
				String: boxscore.BasicGameDataNode.Arena.City,
				Valid:  boxscore.BasicGameDataNode.Arena.City != "",
			},
			State: sql.NullString{
				String: "",
				Valid:  false,
			},
			Country: boxscore.BasicGameDataNode.Arena.Country,
		}

		gameUpdate := dao.GameUpdate{
			NBAHomeTeamID: boxscore.BasicGameDataNode.HomeTeamInfo.TeamID,
			NBAAwayTeamID: boxscore.BasicGameDataNode.AwayTeamInfo.TeamID,
			HomeTeamPoints: sql.NullInt64{
				Int64: int64(homeTeamPoints),
				Valid: homeTeamPointsValid,
			},
			AwayTeamPoints: sql.NullInt64{
				Int64: int64(awayTeamPoints),
				Valid: awayTeamPointsValid,
			},
			GameStatusName:                  gameStatusNameMappings[boxscore.BasicGameDataNode.StatusNum],
			ArenaName:                       boxscore.BasicGameDataNode.Arena.Name,
			Attendance:                      attendance,
			SeasonStartYear:                 boxscore.BasicGameDataNode.SeasonYear,
			SeasonStageName:                 seasonStageNameMappings[int(boxscore.BasicGameDataNode.SeasonStage)],
			Period:                          boxscore.BasicGameDataNode.PeriodNode.CurrentPeriod,
			PeriodTimeRemainingTenthSeconds: periodTimeRemainingTenthSeconds,
			DurationSeconds:                 durationSeconds,
			StartTime:                       boxscore.BasicGameDataNode.GameStartTimeUTC,
			EndTime:                         boxscore.BasicGameDataNode.GameEndTimeUTC,
			NBAGameID:                       boxscore.BasicGameDataNode.GameID,
		}

		arenaUpdatesMap[arenaUpdate.Name] = arenaUpdate
		gameUpdates = append(gameUpdates, gameUpdate)
	}

	refereeUpdates := []dao.RefereeUpdate{}
	for _, refereeUpdate := range refereeFullNameUpdateMap {
		refereeUpdates = append(refereeUpdates, refereeUpdate)
	}

	_, err = gs.RefereeService.UpdateReferees(refereeUpdates)
	if err != nil {
		return nil, err
	}

	arenaUpdates := []dao.ArenaUpdate{}
	for _, arenaUpdate := range arenaUpdatesMap {
		arenaUpdates = append(arenaUpdates, arenaUpdate)
	}

	_, err = gs.ArenaService.UpdateArenas(arenaUpdates)
	if err != nil {
		return nil, err
	}

	updatedGames, err := gs.GameDAO.UpdateGames(gameUpdates)
	if err != nil {
		return nil, err
	}

	err = gs.GameRefereeService.UpdateGameReferees(gameRefereeUpdates)
	if err != nil {
		return nil, err
	}

	return updatedGames, nil
}

func (gs GameService) nbaGameStatusNameMappings() map[int]string {
	return map[int]string{
		1: "scheduled",
		2: "started",
		3: "completed",
	}
}

func (gs GameService) getCurrentLeagueScheduleFromNBAAPI() ([]nba.Game, error) {
	nbaGames, err := nba.GetCurrentLeagueSchedule(nba.GetDailyAPIPaths().LeagueSchedule)
	if err != nil {
		return nil, err
	}

	return nbaGames, nil
}
