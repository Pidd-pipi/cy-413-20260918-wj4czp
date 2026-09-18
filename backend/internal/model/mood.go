package model

import "time"

type Mood struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	MoodLevel  int       `gorm:"not null" json:"mood_level"`
	MoodTags   string    `gorm:"type:text;not null" json:"mood_tags"`
	Note       string    `json:"note"`
	RecordDate time.Time `gorm:"index;not null" json:"record_date"`
	CreatedAt  time.Time `json:"created_at"`
}
