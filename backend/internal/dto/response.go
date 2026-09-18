package dto

import (
	"github.com/blueship581/mindgarden/backend/internal/model"
	"time"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MoodTrashItem 是回收站中的一条情绪记录：日期、等级、标签、备注保持原样。
type MoodTrashItem struct {
	model.Mood
	ExpiresAt time.Time `json:"expires_at"`
}
