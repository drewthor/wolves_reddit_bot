package util

func NBASeasonStageNameMappings() map[int]string {
	return map[int]string{
		1: "pre",
		2: "regular",
		3: "allstar",
		4: "post",
		5: "playin",
	}
}

func NBAGameStatusNameMappings() map[int]string {
	return map[int]string{
		1: "scheduled",
		2: "started",
		3: "completed",
	}
}
