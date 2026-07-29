package dto

import "github.com/han/go-ecommerce/internal/common/enums"

type CreateQuestionRequest struct {
	Title       string            `json:"title" binding:"required"`
	Content     string            `json:"content" binding:"required"`
	Image       string            `json:"image"`
	IsCritical  bool              `json:"is_critical"`
	LicenseType enums.LicenseType `json:"license_type" binding:"required"`
	CategoryID  uint              `json:"category_id" binding:"required"`
}
