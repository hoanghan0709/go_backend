package answerRepository

import (
	"github.com/han/go-ecommerce/internal/answer/model"
)

type RepositoryI interface {
	GetListByQuestionID(questionID uint) ([]model.Answer, error)
	Create(req *model.Answer) error
	Update(req *model.Answer) error
	Delete(questionID uint, answerID uint) error
	GetListAnswer() ([]model.Answer, error)
}
