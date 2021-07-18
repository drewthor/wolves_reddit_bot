package service

type SeasonStageService struct {
}

func (sss *SeasonStageService) NBASeasonStageNameMappings() map[int]string {
	return map[int]string{
		1: "pre",
		2: "regular",
		3: "allstar",
		4: "post",
		5: "playin",
	}
}
