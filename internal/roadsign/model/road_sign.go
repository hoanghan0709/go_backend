package roadsign

import (
	category "github.com/han/go-ecommerce/internal/category/model"
	common "github.com/han/go-ecommerce/internal/common/model"
)

// One teaching record per license keeps applicability explicit. No sign meaning
// or legal applicability is inferred from its name/image by the application.
type RoadSign struct {
	common.BaseModel
	Code        string            `json:"code" gorm:"index"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Meaning     string            `json:"meaning"`
	Description string            `json:"description"`
	SignType    string            `json:"sign_type" gorm:"index"`
	LicenseType string            `json:"license_type" gorm:"index"`
	CategoryID  uint              `json:"category_id" gorm:"index"`
	Category    category.Category `json:"category,omitempty"`
	Published   bool              `json:"published" gorm:"not null;default:false"`
}
