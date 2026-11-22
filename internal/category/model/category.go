package category

import (
	common "github.com/han/go-ecommerce/internal/common/model"
)

type Category struct {
	common.BaseModel
	Type        string `gorm:"not null;default:UNCATEGORIZED;index" json:"type"`
	Name        string `gorm:"size:255;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
}
