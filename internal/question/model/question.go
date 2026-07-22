package model

import model "github.com/han/go-ecommerce/internal/category/model"

type Question struct {
	ID      uint   `gorm:"primaryKey"`
	Title   string `gorm:"size:255"`
	Content string `gorm:"type:text"`
	Image   string

	IsCritical bool

	CategoryID uint
	Category   model.Category

	Answers []Answer
}
