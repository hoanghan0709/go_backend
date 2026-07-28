package usecase

import "github.com/han/go-ecommerce/internal/question/model"

type QuestionService interface {
	GetListQuestion() ([]model.Question, error)
}
