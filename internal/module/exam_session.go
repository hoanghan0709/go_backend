package module

import (
	"time"

	"github.com/han/go-ecommerce/internal/module/auth/model"
)

type ExamSession struct {
	ID uint `gorm:"primaryKey"`

	UserID uint
	User   model.User

	TestID uint
	Test   Test

	Score    int
	IsPassed bool

	StartedAt  time.Time
	FinishedAt *time.Time

	UserAnswers []UserAnswer
}
