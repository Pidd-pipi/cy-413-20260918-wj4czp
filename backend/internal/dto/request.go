package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Nickname string `json:"nickname" validate:"required,min=2,max=40"`
}
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname" validate:"required,min=2,max=40"`
	Avatar    string `json:"avatar"`
	BirthDate string `json:"birth_date"`
	Gender    string `json:"gender" validate:"omitempty,oneof=male female other"`
}
type MoodRequest struct {
	MoodLevel  int      `json:"mood_level" validate:"required,min=1,max=10"`
	MoodTags   []string `json:"mood_tags" validate:"required,min=1,max=5"`
	Note       string   `json:"note" validate:"max=500"`
	RecordDate string   `json:"record_date" validate:"required,datetime=2006-01-02"`
}
type AssessmentRequest struct {
	Title       string `json:"title" validate:"required,max=100"`
	Description string `json:"description" validate:"max=500"`
	Category    string `json:"category" validate:"required"`
	Questions   string `json:"questions" validate:"required"`
	ScoringRule string `json:"scoring_rule" validate:"required"`
}
type TakeAssessmentRequest struct {
	Answers []int `json:"answers" validate:"required,min=1"`
}
type JournalRequest struct {
	Title     string `json:"title" validate:"required,max=120"`
	Content   string `json:"content" validate:"required,max=5000"`
	MoodLevel int    `json:"mood_level" validate:"min=1,max=10"`
	Weather   string `json:"weather" validate:"max=30"`
	IsPrivate bool   `json:"is_private"`
}
