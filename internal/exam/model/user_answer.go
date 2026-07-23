package exam

import (
	common "github.com/han/go-ecommerce/internal/common/model"
	"github.com/han/go-ecommerce/internal/question/model"
)

type UserAnswer struct {
	common.BaseModel

	ExamSessionID uint

	QuestionID uint
	Question   model.Question

	AnswerID uint
	Answer   model.Answer
	// Người dùng có chọn đúng không
	IsCorrect bool
}
