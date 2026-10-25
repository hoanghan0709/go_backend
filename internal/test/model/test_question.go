package test

import (
	common "github.com/han/go-ecommerce/internal/common/model"
	question "github.com/han/go-ecommerce/internal/question/model"
)

type TestQuestion struct {
	common.BaseModel

	TestID     uint `gorm:"not null;uniqueIndex:idx_test_question;uniqueIndex:idx_test_order" json:"test_id"`
	QuestionID uint `gorm:"not null;uniqueIndex:idx_test_question" json:"question_id"`

	// Thứ tự trong đề
	Order int `gorm:"not null;uniqueIndex:idx_test_order" json:"order"`

	// Điểm của câu (nếu sau này có trọng số)
	Score float64 `gorm:"default:1" json:"score"`

	Question question.Question `json:"question,omitempty"`
}
