package exam

import (
	"github.com/han/go-ecommerce/internal/answer/model"
	common "github.com/han/go-ecommerce/internal/common/model"
	modelQuestion "github.com/han/go-ecommerce/internal/question/model"
)

type UserAnswer struct {
	common.BaseModel

	ExamSessionID uint

	QuestionID uint
	Question   modelQuestion.Question

	AnswerID uint
	Answer   model.Answer
	// Người dùng có chọn đúng không
	IsCorrect bool
}
