package model

import "time"

type Assessment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	Category    string    `gorm:"index;not null" json:"category"`
	Questions   string    `gorm:"type:text;not null" json:"questions"`
	ScoringRule string    `gorm:"type:text;not null" json:"scoring_rule"`
	CreatedAt   time.Time `json:"created_at"`
}
