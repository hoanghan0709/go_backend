package dto

import (
	"github.com/han/go-ecommerce/internal/answer/model"
	category "github.com/han/go-ecommerce/internal/category/model"
	"github.com/han/go-ecommerce/internal/common/enums"
)

type ListQuestionResponse struct {
	ID          uint              `json:"id"`
	Explanation string            `json:"explanation"`
	Published   bool              `json:"published"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Image       string            `json:"image"`
	IsCritical  bool              `json:"is_critical"`
	Position    int               `json:"position"`
	LicenseType enums.LicenseType `json:"license_type"`
	CategoryID  uint              `json:"category_id"`
	Category    category.Category `json:"category"`
	Answers     []model.Answer    `json:"answers"`
}
