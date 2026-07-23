package test

import (
	common "github.com/han/go-ecommerce/internal/common/model"
	question "github.com/han/go-ecommerce/internal/question/model"
)

type TestQuestion struct {
	common.BaseModel

	TestID     uint `gorm:"not null;index"`
	QuestionID uint `gorm:"not null;index"`

	// Thứ tự trong đề
	Order int `gorm:"not null"`

	// Điểm của câu (nếu sau này có trọng số)
	Score float64 `gorm:"default:1"`

	Question question.Question
}
