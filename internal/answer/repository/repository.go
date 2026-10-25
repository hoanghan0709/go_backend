package answerRepository

import (
	"errors"
	"strings"

	answerErrors "github.com/han/go-ecommerce/internal/answer/errors"
	"github.com/han/go-ecommerce/internal/answer/model"
	"github.com/han/go-ecommerce/internal/learning"
	question "github.com/han/go-ecommerce/internal/question/model"
	"github.com/han/go-ecommerce/internal/question/validation"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func New(db *gorm.DB) *Repository { return &Repository{db: db} }
func (r *Repository) GetListByQuestionID(id uint) ([]model.Answer, error) {
	out := []model.Answer{}
	err := r.db.Where("question_id = ?", id).Order("\"order\" ASC").Find(&out).Error
	return out, err
}
func (r *Repository) GetListAnswer() ([]model.Answer, error) {
	out := []model.Answer{}
	err := r.db.Order("id").Find(&out).Error
	return out, err
}

// Mọi thay đổi đáp án được kiểm tra trên tập đáp án sau thay đổi, trong cùng
// transaction. Muốn chuyển đáp án đúng nên dùng PUT question để thay cả tập.
func (r *Repository) mutate(qid, aid uint, a *model.Answer, operation string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var q question.Question
		if err := tx.First(&q, qid).Error; err != nil {
			return err
		}
		var answers []model.Answer
		if err := tx.Where("question_id = ?", qid).Order("\"order\"").Find(&answers).Error; err != nil {
			return err
		}
		found := false
		next := []model.Answer{}
		if a != nil {
			a.Content = strings.TrimSpace(a.Content)
			a.Label = strings.ToUpper(strings.TrimSpace(a.Label))
		}
		for _, old := range answers {
			if old.ID == aid {
				found = true
				if operation == "update" {
					next = append(next, *a)
				}
			} else {
				next = append(next, old)
			}
		}
		if operation == "create" {
			next = append(next, *a)
		} else if !found {
			return answerErrors.ErrNotFound
		}
		if err := validation.Answers(next); err != nil {
			return learning.Invalid(err.Error())
		}
		switch operation {
		case "create":
			return tx.Create(a).Error
		case "update":
			return tx.Model(&model.Answer{}).Where("id = ? AND question_id = ?", aid, qid).Updates(map[string]interface{}{"content": a.Content, "label": a.Label, "order": a.Order, "is_correct": a.IsCorrect}).Error
		default:
			if err := tx.Model(&model.Answer{}).Where("id = ?", aid).Updates(map[string]interface{}{"label": gorm.Expr("'archived-' || id"), "order": gorm.Expr("-id")}).Error; err != nil {
				return err
			}
			return tx.Delete(&model.Answer{}, aid).Error
		}
	})
}
func (r *Repository) Create(a *model.Answer) error { return r.mutate(a.QuestionID, 0, a, "create") }
func (r *Repository) Update(a *model.Answer) error {
	err := r.mutate(a.QuestionID, a.ID, a, "update")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return answerErrors.ErrNotFound
	}
	return err
}
func (r *Repository) Delete(qid, aid uint) error { return r.mutate(qid, aid, nil, "delete") }
