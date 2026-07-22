package test

import "github.com/han/go-ecommerce/internal/question/model"

type UserAnswer struct {
	ID uint `gorm:"primaryKey"`

	ExamSessionID uint

	QuestionID uint
	Question   model.Question

	AnswerID uint
	Answer   model.Answer

	IsCorrect bool
}
