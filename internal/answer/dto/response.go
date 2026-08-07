package dto

type AnswerResponse struct {
	Content string `gorm:"type:text;not null" json:"content"`

	Label string `gorm:"size:2;not null;uniqueIndex:idx_answers_question_label" json:"label"`

	Order int `gorm:"not null;uniqueIndex:idx_answers_question_order" json:"order"`

	IsCorrect bool `gorm:"default:false" json:"is_correct"`

	QuestionID uint `gorm:"not null;uniqueIndex:idx_answers_question_label;uniqueIndex:idx_answers_question_order" json:"question_id"`
}
