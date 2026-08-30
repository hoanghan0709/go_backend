package answerRepository

import (
	"errors"

	answerErrors "github.com/han/go-ecommerce/internal/answer/errors"
	"github.com/han/go-ecommerce/internal/answer/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetListByQuestionID(questionID uint) ([]model.Answer, error) {
	var answers []model.Answer
	// err := r.db.Find(&answer).Error
	err := r.db.
		Where("question_id = ?", questionID).
		Order("\"order\" ASC").
		Find(&answers).
		Error

	if err != nil {
		return nil, err
	}
	return answers, nil

}
func (r *Repository) Create(model *model.Answer) error {
	return r.db.Create(model).Error
}
func (r *Repository) GetListAnswer() ([]model.Answer, error) {
	var response []model.Answer
	err := r.db.Find(&response).Error
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (r *Repository) Delete(questionID uint, answerID uint) error {
	result := r.db.
		Where("id = ? AND question_id = ?", answerID, questionID).
		Delete(&model.Answer{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return answerErrors.ErrNotFound
	}

	return nil
}

func (r *Repository) Update(answer *model.Answer) error {
	var existing model.Answer
	if err := r.db.
		Where("id = ? AND question_id = ?", answer.ID, answer.QuestionID).
		First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return answerErrors.ErrNotFound
		}
		return err
	}

	var count int64
	if err := r.db.Model(&model.Answer{}).
		Where("question_id = ? AND id <> ? AND label = ?", answer.QuestionID, answer.ID, answer.Label).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return answerErrors.ErrDuplicateLabel
	}

	if err := r.db.Model(&model.Answer{}).
		Where("question_id = ? AND id <> ? AND \"order\" = ?", answer.QuestionID, answer.ID, answer.Order).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return answerErrors.ErrDuplicateOrder
	}

	existing.Content = answer.Content
	existing.Label = answer.Label
	existing.Order = answer.Order
	existing.IsCorrect = answer.IsCorrect

	return r.db.
		Model(&existing).
		Select("content", "label", "order", "is_correct").
		Updates(&existing).
		Error
}
