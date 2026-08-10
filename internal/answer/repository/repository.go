package answerRepository

import (
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
