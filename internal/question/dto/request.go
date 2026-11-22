package dto

import "github.com/han/go-ecommerce/internal/common/enums"

type CreateQuestionRequest struct {
	Title       string                `json:"title" binding:"required"`
	Content     string                `json:"content" binding:"required"`
	Image       string                `json:"image"`
	IsCritical  bool                  `json:"is_critical"`
	LicenseType enums.LicenseType     `json:"license_type" binding:"required"`
	CategoryID  uint                  `json:"category_id" binding:"required"`
	Answers     []CreateAnswerRequest `json:"answers" binding:"required,min=2,dive"`
}
type CreateAnswerRequest struct {
	Content   string `json:"content" binding:"required"`
	Label     string `json:"label" binding:"required,max=2"`
	Order     int    `json:"order" binding:"required,min=1"`
	IsCorrect bool   `json:"is_correct"`
}
