package exam

import (
	"github.com/han/go-ecommerce/internal/answer/model"
	common "github.com/han/go-ecommerce/internal/common/model"
	modelQuestion "github.com/han/go-ecommerce/internal/question/model"
)

type UserAnswer struct {
	common.BaseModel

	ExamSessionID uint `gorm:"uniqueIndex:idx_session_question"`

	QuestionID uint `gorm:"uniqueIndex:idx_session_question"`
	Question   modelQuestion.Question

	AnswerID uint
	Answer   model.Answer
	// Người dùng có chọn đúng không
	IsCorrect bool
}
