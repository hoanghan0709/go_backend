package usecase

import "github.com/han/go-ecommerce/internal/question/repository"

type Repository interface{}

type Usecase struct {
	repo repository.RepositoryI
}

func New(repo repository.RepositoryI) *Usecase {
	return &Usecase{repo: repo}
}
