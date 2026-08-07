package model

import common "github.com/han/go-ecommerce/internal/common/model"

type Answer struct {
	common.BaseModel
	Content string `gorm:"type:text;not null"`

	Label string `gorm:"size:2;not null;uniqueIndex:idx_answers_question_label"`

	Order int `gorm:"not null;uniqueIndex:idx_answers_question_order"`

	IsCorrect bool `gorm:"default:false"`

	QuestionID uint `gorm:"not null;uniqueIndex:idx_answers_question_label;uniqueIndex:idx_answers_question_order"`
}
