package model

import common "github.com/han/go-ecommerce/internal/common/model"

type Answer struct {
	common.BaseModel
	Content string `gorm:"type:text;not null" json:"content"`

	// Archived answers use an internal label of the form archived-<id>.
	Label string `gorm:"size:32;not null;uniqueIndex:idx_answers_question_label" json:"label"`

	Order int `gorm:"not null;uniqueIndex:idx_answers_question_order" json:"order"`

	IsCorrect bool `gorm:"default:false" json:"is_correct"`

	// Khóa ngoại nối mỗi phương án với câu hỏi; nhiều dòng có thể cùng QuestionID.
	QuestionID uint `gorm:"not null;uniqueIndex:idx_answers_question_label;uniqueIndex:idx_answers_question_order" json:"question_id"`
}
