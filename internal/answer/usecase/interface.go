package usecaseAnswer

import (
	"github.com/han/go-ecommerce/internal/answer/dto"
	"github.com/han/go-ecommerce/internal/answer/model"
)

type AnswerI interface {
	GetList() ([]model.Answer, error)
	Create(req *dto.AnswerRequest) error
}
