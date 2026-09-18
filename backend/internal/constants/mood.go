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

// MoodTrashRetention is how long a soft-deleted mood can be restored from the trash.
const MoodTrashRetention = 7 * 24 * time.Hour
