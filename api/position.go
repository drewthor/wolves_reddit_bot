package api

type Position string

const (
	Unknown Position = ""
	Guard   Position = "guard"
	Forward Position = "forward"
	Center  Position = "center"
)

func PositionFromNBAPosition(pos string) []string {
	switch pos {
	case "G":
		return []string{string(Guard)}
	case "G-F":
		return []string{string(Guard), string(Forward)}
	case "F-G":
		return []string{string(Forward), string(Guard)}
	case "F":
		return []string{string(Forward)}
	case "F-C":
		return []string{string(Forward), string(Center)}
	case "C-F":
		return []string{string(Center), string(Forward)}
	case "C":
		return []string{string(Center)}
	default:
		return []string{string(Unknown)}
	}
}
