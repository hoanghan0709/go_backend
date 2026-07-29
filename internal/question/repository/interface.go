package repository

import "github.com/han/go-ecommerce/internal/question/model"

type RepositoryI interface {
	GetListQuestion() ([]model.Question, error)
	CreateQuestion(question *model.Question) error
}
