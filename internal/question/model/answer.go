package model

import common "github.com/han/go-ecommerce/internal/common/model"

type Answer struct {
	common.BaseModel
	Content string `gorm:"type:text;not null"`

	Label string `gorm:"size:2;not null"`

	Order int `gorm:"not null"`

	IsCorrect bool `gorm:"default:false"`

	QuestionID uint
}
