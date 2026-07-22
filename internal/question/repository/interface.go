package repository

import "github.com/han/go-ecommerce/internal/question/model"

type RepositoryI interface {
	Create(question *model.Question) error
	Update(question *model.Question) error
	Delete(id uint) error

	FindByID(id uint) (*model.Question, error)
	FindAll() ([]model.Question, error)
}
