package module

type UserAnswer struct {
	ID uint `gorm:"primaryKey"`

	ExamSessionID uint

	QuestionID uint
	Question   Question

	AnswerID uint
	Answer   Answer

	IsCorrect bool
}
