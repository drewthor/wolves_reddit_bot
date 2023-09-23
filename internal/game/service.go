package game

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/internal/arena"
	"github.com/drewthor/wolves_reddit_bot/internal/game_referee"
	"github.com/drewthor/wolves_reddit_bot/internal/league"
	"github.com/drewthor/wolves_reddit_bot/internal/playbyplay"
	"github.com/drewthor/wolves_reddit_bot/internal/referee"
	"github.com/drewthor/wolves_reddit_bot/internal/season"
	"github.com/drewthor/wolves_reddit_bot/internal/team"
	"github.com/drewthor/wolves_reddit_bot/internal/team_game_stats"
	"github.com/drewthor/wolves_reddit_bot/util"
	"go.opentelemetry.io/otel"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
)

type GameStatus string

const (
	GameStatusCompleted GameStatus = "completed"
	GameStatusScheduled GameStatus = "scheduled"
	GameStatusStarted   GameStatus = "started"
)

type Service interface {
	GetGameWithID(ctx context.Context, id string) (api.Game, error)
	List(ctx context.Context) ([]api.Game, error)
	GetGameWithNBAID(ctx context.Context, nbaID string) (api.Game, error)
	UpdateGame(ctx context.Context, logger *slog.Logger, gameID string, seasonStartYear int) (api.Game, error)
	UpdateSeasonGames(ctx context.Context, logger *slog.Logger, seasonStartYear int) ([]api.Game, error)
}

func NewService(
	gameStore Store,
	arenaService arena.Service,
	gameRefereeService game_referee.Service,
	leagueService league.Service,
	playByPlayService playbyplay.Service,
	refereeService referee.Service,
	seasonService season.Service,
	teamService team.Service,
	teamGameStatsService team_game_stats.Service,
	nbaClient nba.Client,
	r2Client cloudflare.Client,
) Service {
	return &service{
		gameStore:            gameStore,
		arenaService:         arenaService,
		gameRefereeService:   gameRefereeService,
		leagueService:        leagueService,
		playByPlayService:    playByPlayService,
		refereeService:       refereeService,
		seasonService:        seasonService,
		teamService:          teamService,
		teamGameStatsService: teamGameStatsService,
		nbaClient:            nbaClient,
		r2Client:             r2Client,
	}
}

type gameUpdateRequest struct {
	nbaGameID       string
	seasonStartYear int
}

type service struct {
	gameStore Store

	arenaService         arena.Service
	gameRefereeService   game_referee.Service
	leagueService        league.Service
	playByPlayService    playbyplay.Service
	refereeService       referee.Service
	seasonService        season.Service
	teamService          team.Service
	teamGameStatsService team_game_stats.Service

	nbaClient nba.Client
	r2Client  cloudflare.Client
}

func (s *service) GetGameWithID(ctx context.Context, id string) (api.Game, error) {
	ctx, span := otel.Tracer("game").Start(ctx, "game.service.GetGameWithID")
	defer span.End()

	return s.gameStore.GetGameWithID(ctx, id)
}

func (s *service) List(ctx context.Context) ([]api.Game, error) {
	ctx, span := otel.Tracer("game").Start(ctx, "game.service.List")
	defer span.End()

	games, err := s.gameStore.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}
	return games, nil
}

func (s *service) GetGameWithNBAID(ctx context.Context, id string) (api.Game, error) {
	ctx, span := otel.Tracer("game").Start(ctx, "game.service.GetGameWithNBAID")
	defer span.End()

	return s.gameStore.GetGameWithNBAID(ctx, id)
}

func (s *service) UpdateGame(ctx context.Context, logger *slog.Logger, gameID string, seasonStartYear int) (api.Game, error) {
	ctx, span := otel.Tracer("game").Start(ctx, "game.service.UpdateGame")
	defer span.End()

	g := gameUpdateRequest{
		nbaGameID:       gameID,
		seasonStartYear: seasonStartYear,
	}

	games, err := s.updateGames(ctx, logger, []gameUpdateRequest{g})
	if err != nil {
		return api.Game{}, fmt.Errorf("failed to update game: %w", err)
	}

	if len(games) != 1 {
		return api.Game{}, fmt.Errorf("expected to update 1 game and got %d", len(games))
	}
	return games[0], nil
}

func (s *service) updateGames(ctx context.Context, logger *slog.Logger, gameUpdateRequests []gameUpdateRequest) ([]api.Game, error) {
	ctx, span := otel.Tracer("game").Start(ctx, "game.service.updateGames")
	defer span.End()

	type boxscoreComposite struct {
		Detailed *nba.Boxscore
		Summary  *nba.BoxscoreSummary
	}

	boxscoreResults := make(chan boxscoreComposite, len(gameUpdateRequests))

	wg := &sync.WaitGroup{}
	maxConcurrentCalls := 5
	sem := make(chan int, maxConcurrentCalls)

	worker := func(gameUpdateRequest gameUpdateRequest, boxscoreResults chan<- boxscoreComposite) {
		defer func() {
			wg.Done()
			<-sem
		}()

		sem <- 1

		composite := boxscoreComposite{}

		boxscore, err := nba.GetBoxscoreDetailed(ctx, s.r2Client, util.NBAR2Bucket, gameUpdateRequest.nbaGameID, gameUpdateRequest.seasonStartYear)
		if err != nil {
			logger.Error("could not retrieve detailed boxscore for game", slog.String("game_id", gameUpdateRequest.nbaGameID), slog.Any("error", err))
		} else {
			composite.Detailed = &boxscore
		}

		//filepath := fmt.Sprintf(os.Getenv("STORAGE_PATH")+"/boxscoresummary/%d/%s.json", gameUpdateRequest.seasonStartYear, gameUpdateRequest.nbaGameID)
		objectKey := fmt.Sprintf("boxscoresummary/%d/%s.json", gameUpdateRequest.seasonStartYear, gameUpdateRequest.nbaGameID)
		boxscoreSummary, err := s.nbaClient.GetBoxscoreSummary(ctx, gameUpdateRequest.nbaGameID, util.WithR2OutputWriter(logger, s.r2Client, util.NBAR2Bucket, objectKey))
		if err != nil {
			logger.Error("could not retrieve boxscore summary", slog.String("game_id", gameUpdateRequest.nbaGameID), slog.Any("error", err))
		} else {
			composite.Summary = &boxscoreSummary
		}

		boxscoreResults <- composite
	}

	for _, gur := range gameUpdateRequests {
		wg.Add(1)
		go worker(gur, boxscoreResults)
	}

	go func() {
		wg.Wait()
		close(boxscoreResults)
	}()

	var arenaUpdates []arena.ArenaUpdate
	var refereeUpdates []referee.RefereeUpdate
	var gameSummaryUpdates []GameSummaryUpdate
	var gameUpdates []GameUpdate
	var teamGameStatsTotalUpdates []team_game_stats.TeamGameStatsTotalUpdate
	var gameRefereeUpdates []game_referee.GameRefereeUpdate
	var leagueUpdates []league.LeagueUpdate
	teamIDsMap := make(map[int]bool)

	gameStatusNameMappings := util.NBAGameStatusNameMappings()
	seasonStageNameMappings := util.NBASeasonStageNameMappings()

	leagueNameStartYearUpdatesMap := map[string]int{}

	for boxscoreResult := range boxscoreResults {
		boxscoreSummary := boxscoreResult.Summary
		if boxscoreSummary != nil {
			homeTeamID := boxscoreSummary.HomeTeam.TeamID

			awayTeamID := boxscoreSummary.AwayTeam.TeamID

			//periodTimeRemainingFloat, err := strconv.ParseFloat(boxscoreSummary.BasicGameDataNode.Clock, 64)
			//if err != nil {
			//	not really an error, the nba api returns an empty string if the clock has expired
			//periodTimeRemainingFloat = 0
			//}
			//periodTimeRemainingTenthSeconds := int(periodTimeRemainingFloat * 10)

			attendance := boxscoreSummary.Attendance
			//
			//durationSeconds := 0
			//durationSecondsValid := false

			//if boxscoreSummary.BasicGameDataNode.GameDurationNode != nil {
			//	if hours, err := strconv.Atoi(boxscoreSummary.BasicGameDataNode.GameDurationNode.Hours); err == nil {
			//		if minutes, err := strconv.Atoi(boxscoreSummary.BasicGameDataNode.GameDurationNode.Minutes); err == nil {
			//			durationSeconds = (hours * 60 * 60) + (minutes * 60)
			//			durationSecondsValid = true
			//		}
			//	}
			//}

			// nba api treats "" as 0
			//homeTeamPoints, err := strconv.Atoi(boxscoreSummary.GameInfo)
			//homeTeamPointsValid := true
			//if err != nil {
			//	homeTeamPointsValid = false
			//}

			// nba api treats "" as 0
			//awayTeamPoints, err := strconv.Atoi(boxscoreSummary.BasicGameDataNode.AwayTeamInfo.Points)
			//awayTeamPointsValid := true
			//if err != nil {
			//	awayTeamPointsValid = false
			//}

			endTime := sql.NullTime{
				Valid: false,
			}
			//if boxscoreSummary.EBasicGameDataNode.GameEndTimeUTC != nil {
			//	endTime.Time = *boxscoreSummary.BasicGameDataNode.GameEndTimeUTC
			//	endTime.Valid = true
			//}

			gameUpdate := GameSummaryUpdate{
				NBAHomeTeamID:  homeTeamID,
				NBAAwayTeamID:  awayTeamID,
				HomeTeamPoints: sql.NullInt64{},
				//Int64: int64(homeTeamPoints),
				//Valid: homeTeamPointsValid,
				//},
				AwayTeamPoints: sql.NullInt64{},
				//Int64: int64(awayTeamPoints),
				//Valid: awayTeamPointsValid,
				//},
				GameStatusName:  gameStatusNameMappings[boxscoreSummary.GameStatusID],
				Attendance:      attendance,
				SeasonStartYear: strconv.Itoa(boxscoreSummary.SeasonStartYear),
				//SeasonStageName:                 seasonStageNameMappings[int(boxscoreSummary.S)],
				Period: boxscoreSummary.Period,
				//PeriodTimeRemainingTenthSeconds: periodTimeRemainingTenthSeconds,
				//DurationSeconds: sql.NullInt64{
				//	Int64: int64(durationSeconds),
				//	Valid: durationSecondsValid,
				//},
				//StartTime: boxscoreSummary.BasicGameDataNode.GameStartTimeUTC.Time,
				EndTime:   endTime,
				NBAGameID: boxscoreSummary.GameID,
			}

			gameSummaryUpdates = append(gameSummaryUpdates, gameUpdate)
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

			arenaUpdate := arena.ArenaUpdate{
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
				}

				refereeUpdate := referee.RefereeUpdate{
					NBARefereeID: boxscoreOfficial.PersonID,
					FirstName:    boxscoreOfficial.FirstName,
					LastName:     boxscoreOfficial.LastName,
					JerseyNumber: jerseyNumber,
				}

				gameRefereeUpdate := game_referee.GameRefereeUpdate{
					NBAGameID:    boxscore.GameNode.GameID,
					NBARefereeID: boxscoreOfficial.PersonID,
					Assignment:   boxscoreOfficial.Assignment,
				}

				refereeUpdates = append(refereeUpdates, refereeUpdate)
				gameRefereeUpdates = append(gameRefereeUpdates, gameRefereeUpdate)
			}

			sellout, err := strconv.ParseBool(boxscore.GameNode.Sellout)
			if err != nil {
				sellout = false
			}

			durationSeconds := boxscore.GameNode.TotalDurationMinutes * 60

			endTimeUTC := sql.NullTime{Valid: false}
			if boxscore.Final() {
				endTimeUTC.Valid = true
				endTimeUTC.Time = boxscore.GameNode.GameTimeUTC.Time.Add(time.Duration(durationSeconds) * time.Second)
			}

			gameUpdate := GameUpdate{
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
				SeasonStartYear:                 boxscore.GameNode.GameTimeUTC.Time.Year(),
				SeasonStageName:                 seasonStageNameMappings[2],
				Attendance:                      boxscore.GameNode.Attendance,
				Sellout:                         sellout,
				Period:                          boxscore.GameNode.Period,
				PeriodTimeRemainingTenthSeconds: boxscore.GameNode.GameClock.DurationTenthSeconds,
				DurationSeconds:                 durationSeconds,
				RegulationPeriods:               boxscore.GameNode.RegulationPeriods,
				StartTime:                       boxscore.GameNode.GameTimeUTC.Time,
				EndTime:                         endTimeUTC,
				NBAGameID:                       boxscore.GameNode.GameID,
			}

			gameUpdates = append(gameUpdates, gameUpdate)

			for _, teamData := range []nba.BoxscoreTeam{boxscore.GameNode.HomeTeam, boxscore.GameNode.AwayTeam} {
				if teamData.ID == 0 {
				}
				teamIDsMap[teamData.ID] = true

				teamGameStatsTotalUpdate := team_game_stats.TeamGameStatsTotalUpdate{
					NBAGameID:                    boxscore.GameNode.GameID,
					NBATeamID:                    teamData.ID,
					GameTimePlayedSeconds:        teamData.Statistics.Minutes.DurationTenthSeconds / 10 / 5,
					TotalPlayerTimePlayedSeconds: teamData.Statistics.Minutes.DurationTenthSeconds / 10,
					Points:                       teamData.Statistics.Points,
					PointsAgainst:                teamData.Statistics.PointsAgainst,
					Assists:                      teamData.Statistics.Assists,
					PersonalTurnovers:            teamData.Statistics.Turnovers,
					TeamTurnovers:                teamData.Statistics.TurnoversTeam,
					TotalTurnovers:               teamData.Statistics.TurnoversTotal,
					Steals:                       teamData.Statistics.Steals,
					ThreePointersAttempted:       teamData.Statistics.ThreePointersAttempted,
					ThreePointersMade:            teamData.Statistics.ThreePointersMade,
					FieldGoalsAttempted:          teamData.Statistics.FieldGoalsAttempted,
					FieldGoalsMade:               teamData.Statistics.FieldGoalsMade,
					EffectiveAdjustedFieldGoals:  teamData.Statistics.FieldGoalsEffectiveAdjusted,
					FreeThrowsAttempted:          teamData.Statistics.FreeThrowsAttempted,
					FreeThrowsMade:               teamData.Statistics.FreeThrowsMade,
					Blocks:                       teamData.Statistics.Blocks,
					TimesBlocked:                 teamData.Statistics.BlocksReceived,
					PersonalOffensiveRebounds:    teamData.Statistics.ReboundsOffensive,
					PersonalDefensiveRebounds:    teamData.Statistics.ReboundsDefensive,
					TotalPersonalRebounds:        teamData.Statistics.ReboundsPersonal,
					TeamRebounds:                 teamData.Statistics.ReboundsTeam,
					TeamOffensiveRebounds:        teamData.Statistics.ReboundsTeamOffensive,
					TeamDefensiveRebounds:        teamData.Statistics.ReboundsTeamDefensive,
					TotalOffensiveRebounds:       teamData.Statistics.ReboundsTeamOffensive + teamData.Statistics.ReboundsOffensive,
					TotalDefensiveRebounds:       teamData.Statistics.ReboundsTeamDefensive + teamData.Statistics.ReboundsDefensive,
					TotalRebounds:                teamData.Statistics.ReboundsTotal,
					PersonalFouls:                teamData.Statistics.FoulsPersonal,
					OffensiveFouls:               teamData.Statistics.FoulsOffensive,
					FoulsDrawn:                   teamData.Statistics.FoulsDrawn,
					TeamFouls:                    teamData.Statistics.FoulsTeam,
					PersonalTechnicalFouls:       teamData.Statistics.FoulsTechnical,
					TeamTechnicalFouls:           teamData.Statistics.FoulsTeamTechnical,
					FullTimeoutsRemaining:        0,
					ShortTimeoutsRemaining:       0,
					TotalTimeoutsRemaining:       teamData.TimeoutsRemaining, // starting in 2017 full and short timeouts were combined into one group of timeouts https://www.nba.com/news/nba-board-governors-timeout-rules-game-flow-trade-deadline
					FastBreakPoints:              teamData.Statistics.PointsFastBreak,
					FastBreakPointsAttempted:     teamData.Statistics.FastBreakPointsAttempted,
					FastBreakPointsMade:          teamData.Statistics.FastBreakPointsMade,
					PointsInPaint:                teamData.Statistics.PointsInThePaint,
					PointsInPaintAttempted:       teamData.Statistics.PointsInThePaintAttempted,
					PointsInPaintMade:            teamData.Statistics.PointsInThePaintMade,
					SecondChancePoints:           teamData.Statistics.PointsSecondChance,
					SecondChancePointsAttempted:  teamData.Statistics.PointsSecondChanceAttempted,
					SecondChancePointsMade:       teamData.Statistics.PointsSecondChanceMade,
					PointsOffTurnovers:           teamData.Statistics.PointsOffTurnovers,
					BiggestLead:                  teamData.Statistics.BiggestLead,
					BiggestLeadScore:             teamData.Statistics.BiggestLeadScore,
					BiggestScoringRun:            teamData.Statistics.BiggestScoringRun,
					BiggestScoringRunScore:       teamData.Statistics.BiggestScoringRunScore,
					TimeLeadingTenthSeconds:      teamData.Statistics.TimeLeading.DurationTenthSeconds,
					LeadChanges:                  teamData.Statistics.LeadChanges,
					TimesTied:                    teamData.Statistics.TimesTied,
					TrueShootingAttempts:         teamData.Statistics.TrueShootingAttempts,
					TrueShootingPercentage:       teamData.Statistics.TrueShootingPercentage,
					BenchPoints:                  teamData.Statistics.BenchPoints,
				}

				teamGameStatsTotalUpdates = append(teamGameStatsTotalUpdates, teamGameStatsTotalUpdate)
			}
		}
	}

	for leagueName := range leagueNameStartYearUpdatesMap {
		leagueUpdates = append(leagueUpdates, league.LeagueUpdate{Name: leagueName})
	}

	leaguesUpdated, err := s.leagueService.UpdateLeagues(ctx, leagueUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update leagues %w", err)
	}

	for _, leagueUpdated := range leaguesUpdated {
		seasonStartYear := leagueNameStartYearUpdatesMap[leagueUpdated.Name]
		_, err := s.seasonService.UpdateSeasonForLeague(ctx, leagueUpdated.ID, seasonStartYear)
		if err != nil {
			return nil, fmt.Errorf("failed to update league season: %w", err)
		}

	}

	_, err = s.refereeService.UpdateReferees(ctx, refereeUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update referees: %w", err)
	}

	_, err = s.arenaService.UpdateArenas(ctx, arenaUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update arenas: %w", err)
	}

	updatedGameSummaries, err := s.gameStore.UpdateGamesSummary(ctx, gameSummaryUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update games summary: %w", err)
	}

	updateGamesMap := map[string]api.Game{}

	for _, updatedGame := range updatedGameSummaries {
		updateGamesMap[updatedGame.ID] = updatedGame
	}

	var teamIDs []int
	for teamID, _ := range teamIDsMap {
		teamIDs = append(teamIDs, teamID)
	}

	// make sure team's exist before updating games
	if err := s.teamService.EnsureTeamsExistForLeague(ctx, "00", teamIDs); err != nil {
		return nil, fmt.Errorf("failed to ensure teams exist for league when updating games: %w", err)
	}

	updatedGamesDetailed, err := s.gameStore.UpdateGames(ctx, gameUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update games detailed: %w", err)
	}

	for _, updatedGame := range updatedGamesDetailed {
		updateGamesMap[updatedGame.ID] = updatedGame
	}

	updatedGames := []api.Game{}
	var startedGameIDs []string

	for _, updatedGame := range updateGamesMap {
		updatedGames = append(updatedGames, updatedGame)
		if updatedGame.Duration != nil && *updatedGame.Duration > 0 {
			startedGameIDs = append(startedGameIDs, updatedGame.NBAGameID)
		}
	}

	_, err = s.teamGameStatsService.UpdateTeamGameStatsTotals(ctx, teamGameStatsTotalUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update team game stats totals: %w", err)
	}

	err = s.gameRefereeService.UpdateGameReferees(ctx, gameRefereeUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update game referees: %w", err)
	}

	_, err = s.playByPlayService.UpdatePlayByPlayForGames(ctx, logger, startedGameIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to update play by play for games: %w", err)
	}

	return updatedGames, nil
}

func (s *service) UpdateSeasonGames(ctx context.Context, logger *slog.Logger, seasonStartYear int) ([]api.Game, error) {
	ctx, span := otel.Tracer("game").Start(ctx, "game.service.UpdateSeasonGames")
	defer span.End()

	nbaGames, err := s.getSeasonLeagueScheduleFromNBAAPI(ctx, seasonStartYear)
	if err != nil {
		return nil, err
	}

	var gameUpdateRequests []gameUpdateRequest

	for _, nbaGame := range nbaGames {
		gameUpdateRequests = append(gameUpdateRequests, gameUpdateRequest{
			nbaGameID:       nbaGame.GameID,
			seasonStartYear: seasonStartYear,
		})
	}

	return s.updateGames(ctx, logger, gameUpdateRequests)
}

func (s *service) getSeasonLeagueScheduleFromNBAAPI(ctx context.Context, seasonStartYear int) ([]nba.Game, error) {
	ctx, span := otel.Tracer("game").Start(ctx, "game.service.getSeasonLeagueScheduleFromNBAAPI")
	defer span.End()

	schedule, err := nba.GetLeagueSchedule(ctx, s.r2Client, util.NBAR2Bucket, seasonStartYear)
	if err != nil {
		return nil, err
	}

	var nbaGames []nba.Game
	for _, gameDate := range schedule.LeagueSchedule.GameDates {
		nbaGames = append(nbaGames, gameDate.Games...)
	}

	return nbaGames, nil
}
