package util

import (
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"time"
)

func FormatDate(v time.Time) string { return v.Format("2006-01-02") }
func MoodText(tag string) string {
	m := map[string]string{constants.MoodHappy: "开心", constants.MoodAnxious: "焦虑", constants.MoodTired: "疲惫", constants.MoodAngry: "愤怒", constants.MoodCalm: "平静"}
	return m[tag]
}
func AssessmentText(c string) string {
	m := map[string]string{constants.AssessmentAnxiety: "焦虑", constants.AssessmentDepression: "抑郁", constants.AssessmentStress: "压力", constants.AssessmentSleep: "睡眠"}
	return m[c]
}
func ThemeColor(t string) string { return constants.ThemeColors[t] }
