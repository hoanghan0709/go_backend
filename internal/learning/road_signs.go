package learning

import (
	"errors"
	category "github.com/han/go-ecommerce/internal/category/model"
	question "github.com/han/go-ecommerce/internal/question/model"
	roadsign "github.com/han/go-ecommerce/internal/roadsign/model"
	"gorm.io/gorm"
	"strings"
)

var SignTypes = []string{"PROHIBITION", "WARNING", "MANDATORY", "INFORMATION", "SUPPLEMENTARY", "OTHER"}

func ValidSignType(v string) bool {
	for _, t := range SignTypes {
		if t == v {
			return true
		}
	}
	return false
}

type RoadSignInput struct {
	Code        string `json:"code" binding:"required,max=50"`
	Name        string `json:"name" binding:"required,max=255"`
	Image       string `json:"image"`
	Meaning     string `json:"meaning"`
	Description string `json:"description"`
	SignType    string `json:"sign_type" binding:"required"`
	LicenseType string `json:"license_type" binding:"required"`
	CategoryID  uint   `json:"category_id" binding:"required"`
	Published   bool   `json:"published"`
}

func (s *Service) SaveRoadSign(id uint, in RoadSignInput) (*roadsign.RoadSign, error) {
	out := roadsign.RoadSign{Code: strings.TrimSpace(in.Code), Name: strings.TrimSpace(in.Name), Image: strings.TrimSpace(in.Image), Meaning: strings.TrimSpace(in.Meaning), Description: strings.TrimSpace(in.Description), SignType: in.SignType, LicenseType: in.LicenseType, CategoryID: in.CategoryID, Published: in.Published}
	if out.Code == "" || out.Name == "" || !ValidLicense(out.LicenseType) || !ValidSignType(out.SignType) {
		return nil, Invalid("code/name, valid license_type and sign_type are required")
	}
	if out.Published && (out.Image == "" || out.Meaning == "") {
		return nil, Invalid("published road signs require image and meaning")
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var cat category.Category
		if err := tx.First(&cat, in.CategoryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return Invalid("category does not exist")
			}
			return err
		}
		if cat.Type != "SIGN" {
			return Invalid("road signs require a category with type SIGN")
		}
		if id != 0 {
			var old roadsign.RoadSign
			if err := tx.First(&old, id).Error; err != nil {
				return err
			}
			out.BaseModel = old.BaseModel
			var n int64
			linked := tx.Model(&question.Question{}).Where("road_sign_id = ?", id)
			if err := linked.Count(&n).Error; err != nil {
				return err
			}
			if n > 0 && (old.LicenseType != out.LicenseType || old.CategoryID != out.CategoryID) {
				return Invalid("cannot change license/category while questions reference this sign")
			}
			if !out.Published {
				if err := linked.Where("published = ?", true).Count(&n).Error; err != nil {
					return err
				}
				if n > 0 {
					return Invalid("unpublish linked questions before unpublishing this sign")
				}
			}
		}
		// Code is unique within a license among live records. A deleted record may be
		// replaced deliberately without reviving links from an old question.
		var n int64
		if err := tx.Model(&roadsign.RoadSign{}).Where("code = ? AND license_type = ? AND id <> ?", out.Code, out.LicenseType, id).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return Invalid("road sign code already exists for this license")
		}
		return tx.Omit("Category").Save(&out).Error
	})
	return &out, err
}
func (s *Service) RoadSigns(f Filter, admin bool) ([]roadsign.RoadSign, int64, error) {
	db := s.DB.Model(&roadsign.RoadSign{})
	if !admin {
		db = db.Where("published = ?", true)
	} else if f.Published != nil {
		db = db.Where("published = ?", *f.Published)
	}
	if f.License != "" {
		db = db.Where("license_type = ?", f.License)
	}
	if f.CategoryID != 0 {
		db = db.Where("category_id = ?", f.CategoryID)
	}
	if f.SignType != "" {
		db = db.Where("sign_type = ?", f.SignType)
	}
	if f.Search != "" {
		db = db.Where("code LIKE ? OR name LIKE ? OR meaning LIKE ?", "%"+f.Search+"%", "%"+f.Search+"%", "%"+f.Search+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := []roadsign.RoadSign{}
	p, l := f.Bounds()
	err := db.Preload("Category").Order("code, id").Offset((p - 1) * l).Limit(l).Find(&out).Error
	return out, total, err
}
func (s *Service) RoadSign(id uint, license string, admin bool) (*roadsign.RoadSign, error) {
	db := s.DB.Preload("Category")
	if !admin {
		db = db.Where("published = ?", true)
	}
	if license != "" {
		db = db.Where("license_type = ?", license)
	}
	var out roadsign.RoadSign
	err := db.First(&out, id).Error
	return &out, err
}
func (s *Service) DeleteRoadSign(id uint) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&question.Question{}).Where("road_sign_id = ?", id).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return Invalid("road sign is referenced by questions; remove links first")
		}
		r := tx.Delete(&roadsign.RoadSign{}, id)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
func validateQuestionSign(tx *gorm.DB, q question.Question) error {
	if q.RoadSignID == nil {
		return nil
	}
	if *q.RoadSignID == 0 {
		return Invalid("road_sign_id must be positive or null")
	}
	var sign roadsign.RoadSign
	if err := tx.First(&sign, *q.RoadSignID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Invalid("road_sign_id does not exist")
		}
		return err
	}
	if sign.LicenseType != string(q.LicenseType) || sign.CategoryID != q.CategoryID {
		return Invalid("road sign and question must have the same license_type and category_id")
	}
	if q.Published && !sign.Published {
		return Invalid("publish the road sign before publishing its question")
	}
	return nil
}
