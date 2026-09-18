package model

import "time"

type User struct {
	ID              uint             `gorm:"primaryKey" json:"id"`
	Email           string           `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash    string           `json:"-"`
	Nickname        string           `gorm:"not null" json:"nickname"`
	Avatar          string           `json:"avatar"`
	BirthDate       *time.Time       `json:"birth_date"`
	Gender          string           `json:"gender"`
	Role            string           `gorm:"not null;default:user" json:"role"`
	CreatedAt       time.Time        `json:"created_at"`
	Moods           []Mood           `json:"-"`
	Journals        []Journal        `json:"-"`
	UserAssessments []UserAssessment `json:"-"`
}
