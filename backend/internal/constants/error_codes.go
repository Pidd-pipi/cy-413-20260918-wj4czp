package constants

const (
	CodeOK           = 0
	CodeValidation   = 1001
	CodeUnauthorized = 1002
	CodeForbidden    = 1003
	CodeNotFound     = 1004
	CodeConflict     = 1005
	CodeInternal     = 1500
)

var ErrorCodeHints = map[string]string{"Mood[mood_level]": "1-10", "Mood[mood_tags]": "happy/anxious/tired/angry/calm", "Assessment[category]": "anxiety/depression/stress/sleep"}
