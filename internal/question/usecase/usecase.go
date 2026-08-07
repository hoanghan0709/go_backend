package usecase

import (
	"strings"

	modelAnswer "github.com/han/go-ecommerce/internal/answer/model"
	"github.com/han/go-ecommerce/internal/question/dto"
	questionErrors "github.com/han/go-ecommerce/internal/question/errors"
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
	if len(req.Answers) < 2 {
		return questionErrors.ErrInvalidAnswers
	}

	answers := make([]modelAnswer.Answer, 0, len(req.Answers))
	labels := make(map[string]struct{}, len(req.Answers))
	orders := make(map[int]struct{}, len(req.Answers))
	contents := make(map[string]struct{}, len(req.Answers))

	correctAnswers := 0

	for _, answerReq := range req.Answers {
		label := strings.ToUpper(strings.TrimSpace(answerReq.Label))
		if _, exists := labels[label]; exists {
			return questionErrors.ErrDuplicateAnswerLabel
		}
		labels[label] = struct{}{}
		content := strings.ToLower(
			strings.TrimSpace(answerReq.Content),
		)

		if _, exists := contents[content]; exists {
			return questionErrors.ErrDuplicateAnswerContent
		}

		contents[content] = struct{}{}
		if _, exists := orders[answerReq.Order]; exists {
			return questionErrors.ErrDuplicateAnswerOrder
		}
		orders[answerReq.Order] = struct{}{}

		if answerReq.IsCorrect {
			correctAnswers++
		}

		answers = append(answers, modelAnswer.Answer{
			Content:   strings.TrimSpace(answerReq.Content),
			Label:     label,
			Order:     answerReq.Order,
			IsCorrect: answerReq.IsCorrect,
		})
	}

	if correctAnswers != 1 {
		return questionErrors.ErrInvalidCorrectAnswer
	}

	question := &model.Question{
		Title:       req.Title,
		Content:     req.Content,
		Image:       req.Image,
		IsCritical:  req.IsCritical,
		LicenseType: req.LicenseType,
		CategoryID:  req.CategoryID,
		Answers:     answers,
	}
	return s.repo.CreateQuestion(question)
}
