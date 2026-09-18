package util

func MoodColor(level int) string {
	switch {
	case level >= 8:
		return "#4caf50"
	case level >= 6:
		return "#8bc34a"
	case level >= 4:
		return "#ffb74d"
	default:
		return "#ef5350"
	}
}
