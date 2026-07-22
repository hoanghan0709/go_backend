package test

import "github.com/han/go-ecommerce/internal/question/model"

type Test struct {
	ID    uint   `gorm:"primaryKey"`
	Title string `gorm:"size:255"`

	DurationMinutes int
	PassScore       int

	Questions []model.Question `gorm:"many2many:test_questions;"`
}
