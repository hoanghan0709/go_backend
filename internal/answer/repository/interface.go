package answerRepository

import (
	"github.com/han/go-ecommerce/internal/answer/model"
)

type RepositoryI interface {
	GetList() ([]model.Answer, error)
	Create(req *model.Answer) error
}
