package constants

import "time"

const (
	MoodHappy   = "happy"
	MoodAnxious = "anxious"
	MoodTired   = "tired"
	MoodAngry   = "angry"
	MoodCalm    = "calm"
)

var MoodTags = []string{MoodHappy, MoodAnxious, MoodTired, MoodAngry, MoodCalm}

// MoodTrashRetention 是情绪记录在回收站中的保留天数，超过即永久清除。
const MoodTrashRetention = 7 * 24 * time.Hour
