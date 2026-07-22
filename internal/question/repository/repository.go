package repository

import (
	"github.com/han/go-ecommerce/internal/question/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}
func (r *Repository) Create(question *model.Question) error {
	return r.db.Create(question).Error
}

func (r *Repository) FindByID(id uint) (*model.Question, error) {
	var question model.Question

	err := r.db.
		Preload("Category").
		Preload("Answers").
		First(&question, id).
		Error
	if err != nil {
		return nil, err
	}

	return &question, nil
}
func (r *Repository) FindAll() ([]model.Question, error) {
	var questions []model.Question

	err := r.db.
		Preload("Category").
		Preload("Answers").
		Find(&questions).
		Error

	if err != nil {
		return nil, err
	}

	return questions, nil
}
func (r *Repository) Update(question *model.Question) error {
	return r.db.Save(question).Error
}
func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&model.Question{}, id).Error
}
