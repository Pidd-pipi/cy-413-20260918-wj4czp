package model

import "time"

type UserAssessment struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	AssessmentID uint      `gorm:"index;not null" json:"assessment_id"`
	Answers      string    `gorm:"type:text" json:"answers"`
	Score        int       `json:"score"`
	Result       string    `json:"result"`
	Suggestion   string    `json:"suggestion"`
	CreatedAt    time.Time `json:"created_at"`
}
