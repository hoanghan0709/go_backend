package usecaseAnswer

import (
	"github.com/han/go-ecommerce/internal/answer/dto"
	"github.com/han/go-ecommerce/internal/answer/model"
	answerRepository "github.com/han/go-ecommerce/internal/answer/repository"
)

type Usecase struct {
	repo answerRepository.RepositoryI
}

// Create implements [AnswerI].
func (u *Usecase) Create(questionID uint, req *dto.AnswerRequest) error {
	answer := &model.Answer{
		Content:    req.Content,
		Label:      req.Label,
		Order:      req.Order,
		IsCorrect:  req.IsCorrect,
		QuestionID: questionID,
	}
	return u.repo.Create(answer)
}

func New(repo answerRepository.RepositoryI) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) GetListByQuestionID(questionID uint) ([]model.Answer, error) {
	data, err := u.repo.GetListByQuestionID(questionID)
	if err != nil {
		return nil, err
	}
	return data, nil
}
