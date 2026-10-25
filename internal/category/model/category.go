package category

import (
	common "github.com/han/go-ecommerce/internal/common/model"
)

type Category struct {
	common.BaseModel
	Name        string `gorm:"size:255;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
}
