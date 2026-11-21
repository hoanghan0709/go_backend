package learning

import (
	"errors"

	category "github.com/han/go-ecommerce/internal/category/model"
	question "github.com/han/go-ecommerce/internal/question/model"
	"gorm.io/gorm"
)

type PublishQuestionsInput struct {
	QuestionIDs []uint `json:"question_ids" binding:"required,min=1,max=500,unique,dive,gt=0"`
}

type PublishFailure struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

type PublishQuestionsResult struct {
	PublishedIDs []uint           `json:"published_ids"`
	Failed       []PublishFailure `json:"failed"`
}

// PublishQuestions checks the current persisted content, not a potentially stale
// copy in the browser. Updating only published preserves answer IDs and edits.
// Invalid questions remain drafts; an infrastructure error rolls back all writes.
func (s *Service) PublishQuestions(ids []uint) (*PublishQuestionsResult, error) {
	if len(ids) == 0 || len(ids) > 500 {
		return nil, Invalid("select 1..500 questions")
	}
	seen := make(map[uint]bool, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			return nil, Invalid("question_ids must be unique positive integers")
		}
		seen[id] = true
	}
	out := &PublishQuestionsResult{PublishedIDs: []uint{}, Failed: []PublishFailure{}}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			var q question.Question
			err := tx.Preload("Answers").First(&q, id).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				out.Failed = append(out.Failed, PublishFailure{id, "Câu hỏi không tồn tại hoặc đã bị xóa."})
				continue
			}
			if err != nil {
				return err
			}
			input := QuestionInput{RoadSignID: q.RoadSignID, Title: q.Title, Content: q.Content, Image: q.Image, Explanation: q.Explanation, LicenseType: q.LicenseType, CategoryID: q.CategoryID, IsCritical: q.IsCritical, Published: true}
			for _, a := range q.Answers {
				input.Answers = append(input.Answers, AnswerInput{Content: a.Content, Label: a.Label, Order: a.Order, IsCorrect: a.IsCorrect})
			}
			if _, err = questionFrom(input); err != nil {
				out.Failed = append(out.Failed, PublishFailure{id, err.Error()})
				continue
			}
			var cat category.Category
			err = tx.First(&cat, q.CategoryID).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				out.Failed = append(out.Failed, PublishFailure{id, "Chủ đề không tồn tại hoặc đã bị xóa."})
				continue
			}
			if err != nil {
				return err
			}
			q.Published = true
			if err = validateQuestionSign(tx, q); err != nil {
				if !errors.Is(err, ErrInvalid) {
					return err
				}
				out.Failed = append(out.Failed, PublishFailure{id, err.Error()})
				continue
			}
			out.PublishedIDs = append(out.PublishedIDs, id)
		}
		if len(out.PublishedIDs) == 0 {
			return nil
		}
		return tx.Model(&question.Question{}).Where("id IN ? AND published = ?", out.PublishedIDs, false).Update("published", true).Error
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
