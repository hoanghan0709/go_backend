package learning

import (
	auth "github.com/han/go-ecommerce/internal/auth/model"
	category "github.com/han/go-ecommerce/internal/category/model"
	question "github.com/han/go-ecommerce/internal/question/model"
	"gorm.io/gorm"
	"strings"
)

// Codes are API values for mobile tabs, independent of translated display names.
var CategoryTypes = []string{"UNCATEGORIZED", "CONCEPT", "RULE", "SIGN", "SA_HINH", "CULTURE", "TECHNIQUE", "STRUCTURE"}

func ValidCategoryType(value string) bool {
	for _, v := range CategoryTypes {
		if v == value {
			return true
		}
	}
	return false
}

func (s *Service) SetLicense(user uint, license string) (*auth.User, error) {
	if !ValidLicense(license) {
		return nil, Invalid("license_type must be one of A1, A, B, C1, C, D")
	}
	var out auth.User
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&out, user).Error; err != nil {
			return err
		}
		out.CurrentLicenseType = license
		return tx.Model(&out).Update("current_license_type", license).Error
	})
	return &out, err
}

// An explicit filter wins over the preference. This is not an enrollment/ACL:
// switching license must not alter any existing exam snapshot or history.
func (s *Service) EffectiveLicense(user uint, explicit string) (string, error) {
	if explicit != "" {
		if !ValidLicense(explicit) {
			return "", Invalid("invalid license_type")
		}
		return explicit, nil
	}
	var u auth.User
	if err := s.DB.Select("id", "current_license_type").First(&u, user).Error; err != nil {
		return "", err
	}
	return u.CurrentLicenseType, nil
}

func (s *Service) Categories(f Filter, admin bool) ([]category.Category, error) {
	db := s.DB.Model(&category.Category{})
	if f.CategoryType != "" {
		db = db.Where("type = ?", f.CategoryType)
	}
	if f.Search != "" {
		db = db.Where("name LIKE ?", "%"+f.Search+"%")
	}
	// Mobile tabs only include categories with published questions or sign lessons
	// for the selected license. Admin still sees empty/draft categories.
	if !admin {
		qs := s.DB.Model(&question.Question{}).Select("category_id").Where("published = ?", true)
		signs := s.DB.Table("road_signs").Select("category_id").Where("published = ? AND deleted_at IS NULL", true)
		if f.License != "" {
			qs = qs.Where("license_type = ?", f.License)
			signs = signs.Where("license_type = ?", f.License)
		}
		db = db.Where("id IN (?) OR id IN (?)", qs, signs)
	}
	var out = []category.Category{}
	err := db.Order("id").Find(&out).Error
	return out, err
}

func cleanType(v string) string { return strings.ToUpper(strings.TrimSpace(v)) }
