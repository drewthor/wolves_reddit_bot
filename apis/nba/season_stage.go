package nba

type SeasonStage string

const (
	SeasonStagePreseason SeasonStage = "Preseason"
	SeasonStageRegular   SeasonStage = ""         // NBA appears to use empty string as regular season game for SeriesText in its API
	SeasonStageAllStar   SeasonStage = "AllStar"  // TODO: this is a guess
	SeasonStagePost      SeasonStage = "Playoffs" // TODO: this is a guess
	SeasonStagePlayIn    SeasonStage = "PlayIn"   // TODO: this is a guess
)
