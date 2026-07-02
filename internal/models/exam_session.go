package models

import "time"

type ExamSession struct {
	ID uint `gorm:"primaryKey"`

	UserID uint
	User   User

	TestID uint
	Test   Test

	Score    int
	IsPassed bool

	StartedAt  time.Time
	FinishedAt *time.Time

	UserAnswers []UserAnswer
}
