package model

import (
	"github.com/han/go-ecommerce/internal/answer/model"
	category "github.com/han/go-ecommerce/internal/category/model"
	"github.com/han/go-ecommerce/internal/common/enums"
	common "github.com/han/go-ecommerce/internal/common/model"
)

type Question struct {
	common.BaseModel
	Title       string `gorm:"size:255" json:"title"`
	Content     string `gorm:"type:text" json:"content"`
	Image       string `json:"image"`
	Explanation string `gorm:"type:text" json:"explanation"`
	Published   bool   `gorm:"not null;default:false" json:"published"`

	IsCritical  bool              `json:"is_critical"`
	Position    int               `gorm:"not null;uniqueIndex:idx_questions_category_position" json:"position"`
	LicenseType enums.LicenseType `gorm:"type:varchar(10)" json:"license_type"`
	CategoryID  uint              `gorm:"not null;uniqueIndex:idx_questions_category_position" json:"category_id"`
	Category    category.Category `json:"category,omitempty"`
	Answers     []model.Answer    `json:"answers"`
}
