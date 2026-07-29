package usecase

import (
	"fmt"

	"github.com/han/go-ecommerce/internal/question/dto"
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

func (s *Usecase) CreateQuestion(req *dto.CreateQuestionRequest) error {
	question := &model.Question{
		Title:       req.Title,
		Content:     req.Content,
		Image:       req.Image,
		IsCritical:  req.IsCritical,
		LicenseType: req.LicenseType,
		CategoryID:  req.CategoryID,
	}
	fmt.Print("create questions")
	return s.repo.CreateQuestion(question)
}
