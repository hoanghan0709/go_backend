package usecaseAnswer

import (
	"github.com/han/go-ecommerce/internal/answer/dto"
	"github.com/han/go-ecommerce/internal/answer/model"
)

type AnswerI interface {
	GetListByQuestionID(questionID uint) ([]model.Answer, error)
	Create(questionID uint, req *dto.AnswerRequest) error
	Update(questionID uint, answerID uint, req *dto.AnswerRequest) error
	Delete(questionID uint, answerID uint) error
	GetListAnswer() ([]model.Answer, error)
}
