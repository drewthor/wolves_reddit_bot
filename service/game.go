package service

import (
	"database/sql"
	"log"
	"strconv"
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

func (gs GameService) UpdateGames(seasonStartYear int) ([]api.Game, error) {
	nbaGames, err := gs.getSeasonLeagueScheduleFromNBAAPI(seasonStartYear)
	if err != nil {
		return nil, err
	}

	type boxscoreComposite struct {
		Detailed *nba.Boxscore
		Old      *nba.BoxscoreOld
	}

	boxscoreResults := make(chan boxscoreComposite, len(nbaGames))

	wg := sync.WaitGroup{}
	maxConcurrentCalls := 20
	sem := make(chan int, maxConcurrentCalls)

	worker := func(nbaGame nba.Game, boxscoreResults chan<- boxscoreComposite) {
		defer func() {
			wg.Done()
			<-sem
		}()

		sem <- 1

		composite := boxscoreComposite{}

		boxscore, err := nba.GetBoxscoreDetailed(nbaGame.GameID, seasonStartYear)
		if err != nil {
			log.Printf("could not retrieve detailed boxscore for gameID: %s\n", nbaGame.GameID)
		} else {
			composite.Detailed = &boxscore
		}

		boxscoreOld, err := nba.GetOldBoxscore(nbaGame.GameID, nbaGame.StartDateEastern, seasonStartYear)
		if err != nil {
			log.Printf("could not retrieve old boxscore for gameID: %s\n", nbaGame.GameID)
		} else {
			composite.Old = &boxscoreOld
		}

		boxscoreResults <- composite
	}

	for _, nbaGame := range nbaGames {
		wg.Add(1)
		go worker(nbaGame, boxscoreResults)
	}

	go func() {
		wg.Wait()
		close(boxscoreResults)
	}()

	arenaUpdates := []dao.ArenaUpdate{}
	refereeUpdates := []dao.RefereeUpdate{}
	gameUpdatesOld := []dao.GameUpdateOld{}
	gameUpdates := []dao.GameUpdate{}
	gameRefereeUpdates := []dao.GameRefereeUpdate{}

	gameStatusNameMappings := gs.nbaGameStatusNameMappings()
	seasonStageNameMappings := gs.SeasonStageService.NBASeasonStageNameMappings()

	for boxscoreResult := range boxscoreResults {
		boxscoreOld := boxscoreResult.Old
		if boxscoreOld != nil {
			homeTeamID, err := strconv.Atoi(boxscoreOld.BasicGameDataNode.HomeTeamInfo.TeamID)
			if err != nil {
				log.Printf("failed to convert nba home team id: %s to int\n", boxscoreOld.BasicGameDataNode.HomeTeamInfo.TeamID)
				continue
			}

			awayTeamID, err := strconv.Atoi(boxscoreOld.BasicGameDataNode.AwayTeamInfo.TeamID)
			if err != nil {
				log.Printf("failed to convert nba away team id: %s to int\n", boxscoreOld.BasicGameDataNode.AwayTeamInfo.TeamID)
				continue
			}

			periodTimeRemainingFloat, err := strconv.ParseFloat(boxscoreOld.BasicGameDataNode.Clock, 64)
			if err != nil {
				// not really an error, the nba api returns an empty string if the clock has expired
				periodTimeRemainingFloat = 0
			}
			periodTimeRemainingTenthSeconds := int(periodTimeRemainingFloat * 10)

			attendance, err := strconv.Atoi(boxscoreOld.BasicGameDataNode.Attendance)
			if err != nil {
				// nba api treats "" as 0
				attendance = 0
			}

			durationSeconds := 0
			durationSecondsValid := false

			if boxscoreOld.BasicGameDataNode.GameDurationNode != nil {
				if hours, err := strconv.Atoi(boxscoreOld.BasicGameDataNode.GameDurationNode.Hours); err == nil {
					if minutes, err := strconv.Atoi(boxscoreOld.BasicGameDataNode.GameDurationNode.Minutes); err == nil {
						durationSeconds = (hours * 60 * 60) + (minutes * 60)
						durationSecondsValid = true
					}
				}
			}

			// nba api treats "" as 0
			homeTeamPoints, err := strconv.Atoi(boxscoreOld.BasicGameDataNode.HomeTeamInfo.Points)
			homeTeamPointsValid := true
			if err != nil {
				homeTeamPointsValid = false
			}

			// nba api treats "" as 0
			awayTeamPoints, err := strconv.Atoi(boxscoreOld.BasicGameDataNode.AwayTeamInfo.Points)
			awayTeamPointsValid := true
			if err != nil {
				awayTeamPointsValid = false
			}

			endTime := sql.NullTime{
				Valid: false,
			}
			if boxscoreOld.BasicGameDataNode.GameEndTimeUTC != nil {
				endTime.Time = *boxscoreOld.BasicGameDataNode.GameEndTimeUTC
				endTime.Valid = true
			}

			gameUpdate := dao.GameUpdateOld{
				NBAHomeTeamID: homeTeamID,
				NBAAwayTeamID: awayTeamID,
				HomeTeamPoints: sql.NullInt64{
					Int64: int64(homeTeamPoints),
					Valid: homeTeamPointsValid,
				},
				AwayTeamPoints: sql.NullInt64{
					Int64: int64(awayTeamPoints),
					Valid: awayTeamPointsValid,
				},
				GameStatusName:                  gameStatusNameMappings[boxscoreOld.BasicGameDataNode.StatusNum],
				Attendance:                      attendance,
				SeasonStartYear:                 boxscoreOld.BasicGameDataNode.SeasonYear,
				SeasonStageName:                 seasonStageNameMappings[int(boxscoreOld.BasicGameDataNode.SeasonStage)],
				Period:                          boxscoreOld.BasicGameDataNode.PeriodNode.CurrentPeriod,
				PeriodTimeRemainingTenthSeconds: periodTimeRemainingTenthSeconds,
				DurationSeconds: sql.NullInt64{
					Int64: int64(durationSeconds),
					Valid: durationSecondsValid,
				},
				StartTime: boxscoreOld.BasicGameDataNode.GameStartTimeUTC,
				EndTime:   endTime,
				NBAGameID: boxscoreOld.BasicGameDataNode.GameID,
			}

			gameUpdatesOld = append(gameUpdatesOld, gameUpdate)
		}

		boxscore := boxscoreResult.Detailed
		if boxscore != nil {
			// arena
			arenaCity := sql.NullString{
				Valid: false,
			}
			if boxscore.GameNode.Arena.City != nil {
				arenaCity.String = *boxscore.GameNode.Arena.City
				arenaCity.Valid = true
			}

			arenaState := sql.NullString{
				Valid: false,
			}
			if boxscore.GameNode.Arena.State != nil {
				arenaState.String = *boxscore.GameNode.Arena.State
				arenaState.Valid = true
			}

			arenaUpdate := dao.ArenaUpdate{
				NBAArenaID: boxscore.GameNode.Arena.ID,
				Name:       boxscore.GameNode.Arena.Name,
				City:       arenaCity,
				State:      arenaState,
				Country:    boxscore.GameNode.Arena.Country,
			}

			arenaUpdates = append(arenaUpdates, arenaUpdate)

			// official
			for _, boxscoreOfficial := range boxscore.GameNode.Officials {
				jerseyNumber, err := strconv.Atoi(boxscoreOfficial.JerseyNumber)
				if err != nil {
					log.Printf("could not convert nba official id: %d jersey: %s\n", boxscoreOfficial.PersonID, boxscoreOfficial.JerseyNumber)
				}

				refereeUpdate := dao.RefereeUpdate{
					NBARefereeID: boxscoreOfficial.PersonID,
					FirstName:    boxscoreOfficial.FirstName,
					LastName:     boxscoreOfficial.LastName,
					JerseyNumber: jerseyNumber,
				}

				gameRefereeUpdate := dao.GameRefereeUpdate{
					NBAGameID:    boxscore.GameNode.GameID,
					NBARefereeID: boxscoreOfficial.PersonID,
					Assignment:   boxscoreOfficial.Assignment,
				}

				refereeUpdates = append(refereeUpdates, refereeUpdate)
				gameRefereeUpdates = append(gameRefereeUpdates, gameRefereeUpdate)
			}

			// game
			sellout, err := strconv.ParseBool(boxscore.GameNode.Sellout)
			if err != nil {
				log.Printf("could not parse sellout: %s to bool\n", boxscore.GameNode.Sellout)
				sellout = false
			}

			gameUpdate := dao.GameUpdate{
				NBAHomeTeamID: boxscore.GameNode.HomeTeam.ID,
				NBAAwayTeamID: boxscore.GameNode.AwayTeam.ID,
				HomeTeamPoints: sql.NullInt64{
					Int64: int64(boxscore.GameNode.HomeTeam.Points),
					Valid: true,
				},
				AwayTeamPoints: sql.NullInt64{
					Int64: int64(boxscore.GameNode.AwayTeam.Points),
					Valid: true,
				},
				GameStatusName:                  gameStatusNameMappings[boxscore.GameNode.GameStatus],
				NBAArenaID:                      boxscore.GameNode.Arena.ID,
				Attendance:                      boxscore.GameNode.Attendance,
				Sellout:                         sellout,
				Period:                          boxscore.GameNode.Period,
				PeriodTimeRemainingTenthSeconds: boxscore.GameNode.GameClock.Duration,
				DurationSeconds:                 boxscore.GameNode.TotalDurationMinutes * 60,
				RegulationPeriods:               boxscore.GameNode.RegulationPeriods,
				StartTime:                       boxscore.GameNode.GameTimeUTC,
				NBAGameID:                       boxscore.GameNode.GameID,
			}

			gameUpdates = append(gameUpdates, gameUpdate)
		}
	}

	_, err = gs.RefereeService.UpdateReferees(refereeUpdates)
	if err != nil {
		return nil, err
	}

	_, err = gs.ArenaService.UpdateArenas(arenaUpdates)
	if err != nil {
		return nil, err
	}

	updatedGamesOld, err := gs.GameDAO.UpdateGamesOld(gameUpdatesOld)
	if err != nil {
		return nil, err
	}

	updateGamesMap := map[string]api.Game{}

	for _, updatedGame := range updatedGamesOld {
		updateGamesMap[updatedGame.ID] = updatedGame
	}

	updatedGamesDetailed, err := gs.GameDAO.UpdateGames(gameUpdates)
	if err != nil {
		return nil, err
	}

	for _, updatedGame := range updatedGamesDetailed {
		updateGamesMap[updatedGame.ID] = updatedGame
	}

	updatedGames := []api.Game{}

	for _, updatedGame := range updateGamesMap {
		updatedGames = append(updatedGames, updatedGame)
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

func (gs GameService) getSeasonLeagueScheduleFromNBAAPI(seasonStartYear int) ([]nba.Game, error) {
	nbaGames, err := nba.GetSeasonLeagueSchedule(seasonStartYear)
	if err != nil {
		return nil, err
	}

	return nbaGames, nil
}
