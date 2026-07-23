package test

import common "github.com/han/go-ecommerce/internal/common/model"

type Test struct {
	common.BaseModel

	// Ví dụ: "Đề số 1"
	Name string `gorm:"size:100;not null"`

	// Mô tả
	Description string `gorm:"size:255"`

	// Thời gian làm bài (phút)
	Duration int `gorm:"default:19"`

	// Số câu hỏi
	TotalQuestion int `gorm:"default:25"`

	// Điểm đạt
	PassingScore int `gorm:"default:21"`

	// Quan hệ Many-To-Many
	Questions []TestQuestion `gorm:"foreignKey:TestID;constraint:OnDelete:CASCADE"`
}
