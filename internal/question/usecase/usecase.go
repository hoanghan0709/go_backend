package usecase

import (
	"github.com/han/go-ecommerce/internal/question/model"
	"github.com/han/go-ecommerce/internal/question/repository"
)

type Usecase struct {
	repo repository.RepositoryI
}

func New(repo repository.RepositoryI) *Usecase {
	return &Usecase{repo: repo}
}

func (s *Usecase) GetListQuestion() ([]model.Question, error) {
	data, err := s.repo.GetListQuestion()
	if err != nil {
		return nil, err
	}
	return data, nil
}
