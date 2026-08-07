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

func (r *Repository) GetList() ([]model.Answer, error) {
	var answer []model.Answer
	err := r.db.Find(&answer).Error
	if err != nil {
		return nil, err
	}
	return answer, nil

}
func (r *Repository) Create(model *model.Answer) error {
	return r.db.Create(model).Error
}
