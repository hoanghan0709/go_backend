package models

type Answer struct {
	ID        uint   `gorm:"primaryKey"`
	Content   string `gorm:"type:text"`
	IsCorrect bool

	QuestionID uint
}
