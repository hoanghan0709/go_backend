package test

import common "github.com/han/go-ecommerce/internal/common/model"

type Test struct {
	common.BaseModel
	CategoryType string `json:"category_type" gorm:"index"`

	// Ví dụ: "Đề số 1"
	LicenseType  string `gorm:"index" json:"license_type"`
	Version      string `json:"version"`
	Published    bool   `json:"published"`
	CriticalFail bool   `json:"critical_fail"`
	Name         string `gorm:"size:100;not null" json:"name"`

	// Mô tả
	Description string `gorm:"size:255" json:"description"`

	// Thời gian làm bài (phút)
	Duration int `gorm:"default:19" json:"duration_minutes"`

	// Số câu hỏi
	TotalQuestion int `gorm:"default:25" json:"total_questions"`

	// Điểm đạt
	PassingScore int `gorm:"default:21" json:"passing_score"`

	// Quan hệ Many-To-Many
	Questions []TestQuestion `gorm:"foreignKey:TestID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
}
