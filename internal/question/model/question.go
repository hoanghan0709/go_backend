package model

import (
	"github.com/han/go-ecommerce/internal/answer/model"
	category "github.com/han/go-ecommerce/internal/category/model"
	"github.com/han/go-ecommerce/internal/common/enums"
	common "github.com/han/go-ecommerce/internal/common/model"
)

type Question struct {
	common.BaseModel
	Title   string `gorm:"size:255"`
	Content string `gorm:"type:text"`
	Image   string

	IsCritical  bool
	Position    int               `gorm:"not null;uniqueIndex:idx_questions_category_position"`
	LicenseType enums.LicenseType `gorm:"type:varchar(10)"`
	CategoryID  uint              `gorm:"not null;uniqueIndex:idx_questions_category_position"`
	Category    category.Category
	Answers     []model.Answer
}
