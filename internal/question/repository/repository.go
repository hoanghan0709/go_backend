package repository

import (
	category "github.com/han/go-ecommerce/internal/category/model"
	"github.com/han/go-ecommerce/internal/learning"
	"github.com/han/go-ecommerce/internal/question/model"
	"github.com/han/go-ecommerce/internal/question/validation"
	"gorm.io/gorm"
	"strings"
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
func (r *Repository) GetListQuestion() ([]model.Question, error) {
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
func (r *Repository) CreateQuestion(question *model.Question) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if !learning.ValidLicense(string(question.LicenseType)) || strings.TrimSpace(question.Title) == "" || strings.TrimSpace(question.Content) == "" {
			return learning.Invalid("invalid question")
		}
		if err := validation.Answers(question.Answers); err != nil {
			return learning.Invalid(err.Error())
		}
		var categoryRecord category.Category
		if err := tx.First(&categoryRecord, question.CategoryID).Error; err != nil {
			return err
		}
		var maxPosition int

		err := tx.Unscoped().Model(&model.Question{}).
			Where("category_id = ?", question.CategoryID).
			Select("COALESCE(MAX(position), 0)").
			Scan(&maxPosition).Error
		if err != nil {
			return err
		}

		question.Position = maxPosition + 1
		return tx.Create(question).Error
	})
}

func (r *Repository) Update(question *model.Question) error {
	return r.db.Save(question).Error
}
func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&model.Question{}, id).Error
}
