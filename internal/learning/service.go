// Package learning quản lý ngân hàng câu hỏi, đề thi và luyện tập.
// Các thao tác nhiều bảng chạy trong transaction để không xuất bản dữ liệu dở dang.
package learning

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	answer "github.com/han/go-ecommerce/internal/answer/model"
	category "github.com/han/go-ecommerce/internal/category/model"
	"github.com/han/go-ecommerce/internal/common/enums"
	question "github.com/han/go-ecommerce/internal/question/model"
	"github.com/han/go-ecommerce/internal/question/validation"
	test "github.com/han/go-ecommerce/internal/test/model"
	"gorm.io/gorm"
)

var ErrInvalid = errors.New("invalid request")

func Invalid(message string) error { return fmt.Errorf("%w: %s", ErrInvalid, message) }

type Service struct{ DB *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{DB: db} }

type QuestionInput struct {
	Title       string            `json:"title" binding:"required"`
	Content     string            `json:"content" binding:"required"`
	Image       string            `json:"image"`
	Explanation string            `json:"explanation"`
	LicenseType enums.LicenseType `json:"license_type" binding:"required"`
	CategoryID  uint              `json:"category_id" binding:"required"`
	IsCritical  bool              `json:"is_critical"`
	Published   bool              `json:"published"`
	Answers     []AnswerInput     `json:"answers" binding:"required,min=2,dive"`
}
type AnswerInput struct {
	Content   string `json:"content" binding:"required"`
	Label     string `json:"label" binding:"required,max=2"`
	Order     int    `json:"order" binding:"min=1"`
	IsCorrect bool   `json:"is_correct"`
}

func ValidLicense(l string) bool {
	switch l {
	case "A1", "A", "B", "C1", "C", "D":
		return true
	}
	return false
}
func questionFrom(in QuestionInput) (question.Question, error) {
	q := question.Question{Title: strings.TrimSpace(in.Title), Content: strings.TrimSpace(in.Content), Image: strings.TrimSpace(in.Image), Explanation: strings.TrimSpace(in.Explanation), LicenseType: in.LicenseType, CategoryID: in.CategoryID, IsCritical: in.IsCritical, Published: in.Published}
	if q.Title == "" {
		return q, Invalid("title must not be blank")
	}
	if q.Content == "" {
		return q, Invalid("content must not be blank")
	}
	if !ValidLicense(string(q.LicenseType)) {
		return q, Invalid("license_type must be one of A1, A, B, C1, C, D")
	}
	if q.CategoryID == 0 {
		return q, Invalid("category_id must be a positive integer")
	}
	if q.Published && q.Explanation == "" {
		return q, Invalid("published questions require an explanation")
	}
	for _, a := range in.Answers {
		q.Answers = append(q.Answers, answer.Answer{Content: a.Content, Label: a.Label, Order: a.Order, IsCorrect: a.IsCorrect})
	}
	if err := validation.Answers(q.Answers); err != nil {
		return q, Invalid(err.Error())
	}
	return q, nil
}
func saveQuestion(tx *gorm.DB, id uint, in QuestionInput) (*question.Question, error) {
	q, err := questionFrom(in)
	if err != nil {
		return nil, err
	}
	var cat category.Category
	if err = tx.First(&cat, in.CategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Invalid(fmt.Sprintf("category_id: category %d does not exist or was deleted", in.CategoryID))
		}
		return nil, err
	}
	if id != 0 {
		var old question.Question
		if err = tx.First(&old, id).Error; err != nil {
			return nil, err
		}
		q.BaseModel = old.BaseModel
		q.SourceURL = old.SourceURL // Giữ nguồn gốc khi admin biên tập câu đã crawl.
		q.Position = old.Position
		if old.CategoryID != q.CategoryID {
			q.Position = 0
		}
	}
	if q.Position == 0 {
		// Unscoped tính cả câu đã xóa mềm vì unique index vẫn giữ vị trí của chúng.
		var max int
		if err = tx.Unscoped().Model(&question.Question{}).Where("category_id = ?", q.CategoryID).Select("COALESCE(MAX(position),0)").Scan(&max).Error; err != nil {
			return nil, err
		}
		q.Position = max + 1
	}
	answers := q.Answers
	q.Answers = nil
	if id == 0 {
		err = tx.Omit("Category", "Answers").Create(&q).Error
	} else {
		err = tx.Omit("Category", "Answers").Save(&q).Error
		// Xóa mềm đáp án cũ để các quan hệ lịch sử vẫn tồn tại. Nhãn/thứ tự được
		// chuyển sang namespace lưu trữ trước khi thay cả tập đáp án trong transaction.
		if err == nil {
			err = tx.Model(&answer.Answer{}).Where("question_id = ?", id).Updates(map[string]interface{}{"label": gorm.Expr("'archived-' || id"), "order": gorm.Expr("-id")}).Error
		}
		if err == nil {
			err = tx.Where("question_id = ?", id).Delete(&answer.Answer{}).Error
		}
	}
	if err != nil {
		return nil, err
	}
	for i := range answers {
		answers[i].QuestionID = q.ID
	}
	if err = tx.Create(&answers).Error; err != nil {
		return nil, err
	}
	q.Answers = answers
	q.Category = cat
	return &q, nil
}
func (s *Service) SaveQuestion(id uint, in QuestionInput) (out *question.Question, err error) {
	err = s.DB.Transaction(func(tx *gorm.DB) error { var e error; out, e = saveQuestion(tx, id, in); return e })
	return
}
func (s *Service) ImportQuestions(inputs []QuestionInput) (out []question.Question, err error) {
	if len(inputs) == 0 || len(inputs) > 500 {
		return nil, Invalid("import must contain 1..500 questions")
	}
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		for i, in := range inputs {
			q, e := saveQuestion(tx, 0, in)
			if e != nil {
				return fmt.Errorf("row %d: %w", i+1, e)
			}
			out = append(out, *q)
		}
		return nil
	})
	return
}
func (s *Service) DeleteQuestion(id uint) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&test.TestQuestion{}).Where("question_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return Invalid("question is used by a test; update the test first")
		}
		r := tx.Delete(&question.Question{}, id)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

type Filter struct {
	Page       int
	Limit      int
	License    string
	CategoryID uint
	Critical   *bool
	Search     string
}

func (f Filter) Bounds() (int, int) {
	p, l := f.Page, f.Limit
	if p < 1 {
		p = 1
	}
	if l < 1 {
		l = 20
	}
	if l > 100 {
		l = 100
	}
	return p, l
}
func filterQuestions(db *gorm.DB, f Filter) *gorm.DB {
	if f.License != "" {
		db = db.Where("license_type = ?", f.License)
	}
	if f.CategoryID != 0 {
		db = db.Where("category_id = ?", f.CategoryID)
	}
	if f.Critical != nil {
		db = db.Where("is_critical = ?", *f.Critical)
	}
	if f.Search != "" {
		db = db.Where("title LIKE ? OR content LIKE ?", "%"+f.Search+"%", "%"+f.Search+"%")
	}
	return db
}
func (s *Service) Questions(f Filter, admin bool) ([]question.Question, int64, error) {
	db := filterQuestions(s.DB.Model(&question.Question{}), f)
	if !admin {
		db = db.Where("published = ?", true)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	p, l := f.Bounds()
	out := []question.Question{}
	err := db.Preload("Category").Preload("Answers", func(tx *gorm.DB) *gorm.DB { return tx.Order("\"order\" ASC") }).Order("category_id, position, id").Limit(l).Offset((p - 1) * l).Find(&out).Error
	return out, total, err
}

type CategoryInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (s *Service) SaveCategory(id uint, in CategoryInput) (*category.Category, error) {
	c := category.Category{Name: strings.TrimSpace(in.Name), Description: strings.TrimSpace(in.Description)}
	if c.Name == "" {
		return nil, Invalid("name is required")
	}
	if id != 0 {
		if err := s.DB.First(&c, id).Error; err != nil {
			return nil, err
		}
		c.Name = strings.TrimSpace(in.Name)
		c.Description = strings.TrimSpace(in.Description)
	}
	err := s.DB.Save(&c).Error
	return &c, err
}
func (s *Service) DeleteCategory(id uint) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&question.Question{}).Where("category_id = ?", id).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return Invalid("category still contains questions")
		}
		r := tx.Delete(&category.Category{}, id)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

type TestInput struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	LicenseType  string `json:"license_type" binding:"required"`
	Version      string `json:"version" binding:"required"`
	Duration     int    `json:"duration_minutes" binding:"min=1,max=180"`
	PassingScore int    `json:"passing_score" binding:"min=1"`
	CriticalFail bool   `json:"critical_fail"`
	Published    bool   `json:"published"`
	QuestionIDs  []uint `json:"question_ids"`
}

func saveTest(tx *gorm.DB, id uint, in TestInput) (*test.Test, error) {
	n := len(in.QuestionIDs)
	if err := validateTestConfig(in, n); err != nil {
		return nil, err
	}

	seen := map[uint]bool{}
	for i, qid := range in.QuestionIDs {
		if qid == 0 {
			return nil, Invalid(fmt.Sprintf("question_ids[%d] must be a positive integer", i))
		}
		if seen[qid] {
			return nil, Invalid(fmt.Sprintf("question_ids[%d]: question %d is duplicated", i, qid))
		}
		seen[qid] = true
		var q question.Question
		if err := tx.Preload("Answers").First(&q, qid).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, Invalid(fmt.Sprintf("question_ids[%d]: question %d does not exist or was deleted", i, qid))
			}
			return nil, err
		}
		if !q.Published {
			return nil, Invalid(fmt.Sprintf("question_ids[%d]: question %d is unpublished; set published=true before adding it to a test", i, qid))
		}
		if string(q.LicenseType) != in.LicenseType {
			return nil, Invalid(fmt.Sprintf("question_ids[%d]: question %d has license_type=%s; test requires %s", i, qid, q.LicenseType, in.LicenseType))
		}
		if err := validation.Answers(q.Answers); err != nil {
			return nil, Invalid(fmt.Sprintf("question_ids[%d]: question %d: %s", i, qid, err))
		}
	}
	t := test.Test{Name: strings.TrimSpace(in.Name), Description: in.Description, LicenseType: in.LicenseType, Version: strings.TrimSpace(in.Version), Duration: in.Duration, TotalQuestion: n, PassingScore: in.PassingScore, CriticalFail: in.CriticalFail, Published: in.Published}
	if id != 0 {
		var old test.Test
		if err := tx.First(&old, id).Error; err != nil {
			return nil, err
		}
		t.BaseModel = old.BaseModel
	}
	if err := tx.Omit("Questions").Save(&t).Error; err != nil {
		return nil, err
	}
	if id != 0 {
		if err := tx.Unscoped().Where("test_id = ?", id).Delete(&test.TestQuestion{}).Error; err != nil {
			return nil, err
		}
	}
	for i, qid := range in.QuestionIDs {
		t.Questions = append(t.Questions, test.TestQuestion{TestID: t.ID, QuestionID: qid, Order: i + 1, Score: 1})
	}
	if err := tx.Create(&t.Questions).Error; err != nil {
		return nil, err
	}
	return &t, nil
}
func (s *Service) SaveTest(id uint, in TestInput) (out *test.Test, err error) {
	err = s.DB.Transaction(func(tx *gorm.DB) error { var e error; out, e = saveTest(tx, id, in); return e })
	return
}

type Quota struct {
	CategoryID    uint `json:"category_id"`
	Count         int  `json:"count"`
	CriticalCount int  `json:"critical_count"`
}
type GenerateInput struct {
	TestInput
	Quotas []Quota `json:"quotas" binding:"required,min=1"`
}

func (s *Service) GenerateTest(in GenerateInput) (out *test.Test, err error) {
	total := 0
	for i, quota := range in.Quotas {
		if quota.Count < 1 || quota.Count > 200 {
			return nil, Invalid(fmt.Sprintf("quotas[%d].count must be between 1 and 200; received %d", i, quota.Count))
		}
		total += quota.Count
	}
	if err := validateTestConfig(in.TestInput, total); err != nil {
		return nil, err
	}
	// Chọn câu điểm liệt và câu thường riêng theo từng chủ đề, tránh chọn ngẫu nhiên
	// rồi vô tình thiếu/thừa số câu điểm liệt so với cấu hình của người quản trị.
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		in.QuestionIDs = nil
		seen := map[uint]bool{}
		for i, quota := range in.Quotas {
			if quota.CategoryID == 0 {
				return Invalid(fmt.Sprintf("quotas[%d].category_id must be a positive integer", i))
			}
			if seen[quota.CategoryID] {
				return Invalid(fmt.Sprintf("quotas[%d].category_id: category %d is duplicated", i, quota.CategoryID))
			}
			if quota.CriticalCount < 0 || quota.CriticalCount > quota.Count {
				return Invalid(fmt.Sprintf("quotas[%d].critical_count must be between 0 and %d; received %d", i, quota.Count, quota.CriticalCount))
			}
			var cat category.Category
			if e := tx.First(&cat, quota.CategoryID).Error; e != nil {
				if errors.Is(e, gorm.ErrRecordNotFound) {
					return Invalid(fmt.Sprintf("quotas[%d].category_id: category %d does not exist or was deleted", i, quota.CategoryID))
				}
				return e
			}

			seen[quota.CategoryID] = true
			for _, critical := range []bool{true, false} {
				count := quota.Count - quota.CriticalCount
				if critical {
					count = quota.CriticalCount
				}
				if count == 0 {
					continue
				}
				var ids []uint
				if err := tx.Model(&question.Question{}).Where("published = ? AND license_type = ? AND category_id = ? AND is_critical = ?", true, in.LicenseType, quota.CategoryID, critical).Pluck("id", &ids).Error; err != nil {
					return err
				}
				if len(ids) < count {
					return Invalid(fmt.Sprintf("quotas[%d]: category %d needs %d published %s questions (critical=%t), but only %d are available", i, quota.CategoryID, count, in.LicenseType, critical, len(ids)))
				}
				rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
				in.QuestionIDs = append(in.QuestionIDs, ids[:count]...)
			}
		}
		rand.Shuffle(len(in.QuestionIDs), func(i, j int) { in.QuestionIDs[i], in.QuestionIDs[j] = in.QuestionIDs[j], in.QuestionIDs[i] })
		var e error
		out, e = saveTest(tx, 0, in.TestInput)
		return e
	})
	return
}

type PracticeAttempt struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"-"`
	QuestionID  uint      `gorm:"index" json:"question_id"`
	CategoryID  uint      `json:"category_id"`
	LicenseType string    `json:"license_type"`
	AnswerID    uint      `json:"answer_id"`
	IsCorrect   bool      `json:"is_correct"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Service) Practice(user, qid, aid uint) (*PracticeAttempt, *question.Question, error) {
	var attempt PracticeAttempt
	var q question.Question
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Answers", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC") }).Where("published = ?", true).First(&q, qid).Error; err != nil {
			return err
		}
		found := false
		correct := false
		for _, a := range q.Answers {
			if a.ID == aid {
				found = true
				correct = a.IsCorrect
			}
		}
		if !found {
			return Invalid(fmt.Sprintf("answer_id %d does not belong to question %d; choose an answer from this question", aid, qid))
		}
		attempt = PracticeAttempt{UserID: user, QuestionID: qid, CategoryID: q.CategoryID, LicenseType: string(q.LicenseType), AnswerID: aid, IsCorrect: correct}
		return tx.Create(&attempt).Error
	})
	return &attempt, &q, err
}

// Validate business rules separately from HTTP tags: imports and direct service
// callers must receive the same checks, and generation validates before querying.
func validateTestConfig(in TestInput, n int) error {
	if strings.TrimSpace(in.Name) == "" {
		return Invalid("name must not be blank")
	}
	if strings.TrimSpace(in.Version) == "" {
		return Invalid("version must not be blank")
	}
	if !ValidLicense(in.LicenseType) {
		return Invalid("license_type must be one of A1, A, B, C1, C, D")
	}
	if in.Duration < 1 || in.Duration > 180 {
		return Invalid("duration_minutes must be between 1 and 180")
	}
	if n < 1 || n > 200 {
		return Invalid(fmt.Sprintf("test must contain 1..200 questions; received %d (question_ids or total quotas.count)", n))
	}
	if in.PassingScore < 1 || in.PassingScore > n {
		return Invalid(fmt.Sprintf("passing_score must be between 1 and %d (number of questions); received %d", n, in.PassingScore))
	}
	return nil
}
