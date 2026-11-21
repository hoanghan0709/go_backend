package usecase

import (
	"encoding/json"
	"errors"
	"github.com/han/go-ecommerce/internal/common/enums"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	"github.com/han/go-ecommerce/internal/learning"
	question "github.com/han/go-ecommerce/internal/question/model"
	"github.com/han/go-ecommerce/internal/question/validation"
	"gorm.io/gorm"
	"time"
)

type RandomRequest struct {
	LicenseType  string `json:"license_type"`
	TemplateID   uint   `json:"template_id"`
	CategoryType string `json:"category_type"`
}

// StartRandom creates a private session directly, not a public reusable Test.
// Selection, validation and snapshot creation share one transaction. No partial
// session is left if the bank has insufficient data or an invalid answer set.
func (u *Usecase) StartRandom(user uint, in RandomRequest) (out *View, err error) {
	err = u.repo.Transaction(func(tx *gorm.DB) error {
		s := learning.New(tx)
		license, e := s.EffectiveLicense(user, in.LicenseType)
		if e != nil {
			return e
		}
		if license == "" {
			return learning.Invalid("choose a license_type or save your current license first")
		}
		if in.CategoryType != "" && !learning.ValidCategoryType(in.CategoryType) {
			return learning.Invalid("invalid category_type")
		}
		query := tx.Where("published = ? AND license_type = ?", true, license)
		if in.TemplateID != 0 {
			query = query.Where("id = ?", in.TemplateID)
			if in.CategoryType != "" {
				query = query.Where("category_type = ?", in.CategoryType)
			}
		} else {
			query = query.Where("category_type = ?", in.CategoryType)
		}
		var template learning.ExamTemplate
		if e = query.Order("id DESC").First(&template).Error; e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				return learning.Invalid("no published exam template matches this license/category; ask admin to configure one")
			}
			return e
		}
		ids, e := learning.SelectQuestionIDs(tx, template.Config)
		if e != nil {
			return e
		}
		snap := exam.Snapshot{Name: template.Name, LicenseType: license, Version: template.Config.Version, CategoryType: template.CategoryType, Questions: []exam.SnapshotQuestion{}}
		for _, id := range ids {
			var q question.Question
			if e = tx.Preload("Answers", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC") }).First(&q, id).Error; e != nil {
				return e
			}
			if e = validation.Answers(q.Answers); e != nil {
				return learning.Invalid(e.Error())
			}
			sq := exam.SnapshotQuestion{ID: q.ID, RoadSignID: q.RoadSignID, Title: q.Title, Content: q.Content, Image: q.Image, CategoryID: q.CategoryID, IsCritical: q.IsCritical, Explanation: q.Explanation, Answers: []exam.SnapshotAnswer{}}
			for _, a := range q.Answers {
				correct := a.IsCorrect
				sq.Answers = append(sq.Answers, exam.SnapshotAnswer{ID: a.ID, Content: a.Content, Label: a.Label, Order: a.Order, IsCorrect: &correct})
			}
			snap.Questions = append(snap.Questions, sq)
		}
		raw, e := json.Marshal(snap)
		if e != nil {
			return e
		}
		now := u.Now()
		session := exam.ExamSession{UserID: user, TemplateID: template.ID, Snapshot: string(raw), StartedAt: now, ExpiresAt: now.Add(time.Duration(template.Config.Duration) * time.Minute), PassingScore: template.Config.PassingScore, CriticalFail: template.Config.CriticalFail, Status: enums.ExamDoing}
		if e = tx.Omit("User", "Test", "UserAnswers").Create(&session).Error; e != nil {
			return e
		}
		out, e = view(tx, &session, now)
		return e
	})
	return
}
