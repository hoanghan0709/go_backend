package learning

import (
	"errors"
	category "github.com/han/go-ecommerce/internal/category/model"
	common "github.com/han/go-ecommerce/internal/common/model"
	"gorm.io/gorm"
	"strings"
)

// Configuration is versioned by the admin; its contents are copied into the
// exam snapshot, so editing/deleting a template never changes an ongoing exam.
type ExamTemplate struct {
	common.BaseModel
	Name         string        `json:"name"`
	LicenseType  string        `json:"license_type" gorm:"index"`
	CategoryType string        `json:"category_type" gorm:"index"`
	Published    bool          `json:"published" gorm:"not null;default:false;index"`
	Config       GenerateInput `json:"config" gorm:"serializer:json;type:text"`
}

func (s *Service) SaveTemplate(id uint, in GenerateInput) (*ExamTemplate, error) {
	if len(in.QuestionIDs) > 0 {
		return nil, Invalid("templates use quotas, not question_ids")
	}
	if len(in.Quotas) == 0 || len(in.Quotas) > 200 {
		return nil, Invalid("templates require 1..200 quotas")
	}
	total := 0
	seen := map[uint]bool{}
	for _, q := range in.Quotas {
		if q.CategoryID == 0 || seen[q.CategoryID] || q.Count < 1 || q.Count > 200 || q.CriticalCount < 0 || q.CriticalCount > q.Count {
			return nil, Invalid("invalid or duplicate quota")
		}
		seen[q.CategoryID] = true
		total += q.Count
	}
	if err := validateTestConfig(in.TestInput, total); err != nil {
		return nil, err
	}
	out := ExamTemplate{Name: strings.TrimSpace(in.Name), LicenseType: in.LicenseType, CategoryType: in.CategoryType, Published: in.Published, Config: in}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		for _, q := range in.Quotas {
			var cat category.Category
			if err := tx.First(&cat, q.CategoryID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return Invalid("quota category does not exist")
				}
				return err
			}
			if in.CategoryType != "" && cat.Type != in.CategoryType {
				return Invalid("quota category type does not match template")
			}
		}
		if id != 0 {
			var old ExamTemplate
			if err := tx.First(&old, id).Error; err != nil {
				return err
			}
			out.BaseModel = old.BaseModel
		}
		return tx.Save(&out).Error
	})
	return &out, err
}
func (s *Service) Templates(f Filter, admin bool) ([]ExamTemplate, int64, error) {
	db := s.DB.Model(&ExamTemplate{})
	if !admin {
		db = db.Where("published = ?", true)
	} else if f.Published != nil {
		db = db.Where("published = ?", *f.Published)
	}
	if f.License != "" {
		db = db.Where("license_type = ?", f.License)
	}
	if f.CategoryType != "" {
		db = db.Where("category_type = ?", f.CategoryType)
	}
	if f.Search != "" {
		db = db.Where("name LIKE ?", "%"+f.Search+"%")
	}
	var n int64
	if err := db.Count(&n).Error; err != nil {
		return nil, 0, err
	}
	out := []ExamTemplate{}
	p, l := f.Bounds()
	err := db.Order("id DESC").Offset((p - 1) * l).Limit(l).Find(&out).Error
	return out, n, err
}
