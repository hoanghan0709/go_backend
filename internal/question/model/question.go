package model

import (
	"github.com/han/go-ecommerce/internal/category/category"
	"github.com/han/go-ecommerce/internal/common/enums"
	common "github.com/han/go-ecommerce/internal/common/model"
)

type Question struct {
	common.BaseModel
	Title   string `gorm:"size:255"`
	Content string `gorm:"type:text"`
	Image   string

	IsCritical  bool
	Order       int               `gorm:"not null"`
	LicenseType enums.LicenseType `gorm:"type:varchar(10)"`
	CategoryID  uint
	Category    category.Category

	Answers []Answer
}
