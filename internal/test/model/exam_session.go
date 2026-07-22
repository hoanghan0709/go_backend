package test

import (
	"time"

	auth "github.com/han/go-ecommerce/internal/auth/model"
)

type ExamSession struct {
	ID uint `gorm:"primaryKey"`

	UserID uint
	User   auth.User

	TestID uint
	Test   Test

	Score    int
	IsPassed bool

	StartedAt  time.Time
	FinishedAt *time.Time

	UserAnswers []UserAnswer
}
