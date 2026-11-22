package repository

import (
	"time"

	"github.com/han/go-ecommerce/internal/common/enums"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	"github.com/han/go-ecommerce/internal/learning"
	"gorm.io/gorm"
)

// Repository cung cấp transaction cho toàn bộ vòng đời phiên thi; việc lưu đáp án
// và kiểm tra trạng thái phải dùng cùng transaction để không sửa được bài đã nộp.
type Repository struct{ DB *gorm.DB }

func New(db *gorm.DB) *Repository { return &Repository{DB: db} }

func (r *Repository) Transaction(fn func(*gorm.DB) error) error { return r.DB.Transaction(fn) }
func (r *Repository) ExpiredIDs(user uint, now time.Time) ([]uint, error) {
	var ids []uint
	err := r.DB.Model(&exam.ExamSession{}).Where("user_id = ? AND status = ? AND expires_at <= ?", user, enums.ExamDoing, now).Pluck("id", &ids).Error
	return ids, err
}
func (r *Repository) History(user uint, page, limit int) ([]exam.ExamSession, int64, error) {
	return r.HistoryFiltered(user, learning.Filter{Page: page, Limit: limit})
}
func (r *Repository) HistoryFiltered(user uint, f learning.Filter) ([]exam.ExamSession, int64, error) {
	page, limit := f.Bounds()
	db := r.DB.Model(&exam.ExamSession{}).Where("user_id = ?", user)
	// Filters use immutable snapshot metadata, not today's template/question data.
	licenseFilter := "json_extract(snapshot, '$.license_type') = ?"
	categoryFilter := "json_extract(snapshot, '$.category_type') = ?"
	if db.Dialector.Name() == "postgres" {
		licenseFilter = "CAST(NULLIF(snapshot, '') AS jsonb) ->> 'license_type' = ?"
		categoryFilter = "CAST(NULLIF(snapshot, '') AS jsonb) ->> 'category_type' = ?"
	}
	if f.License != "" {
		db = db.Where(licenseFilter, f.License)
	}
	if f.CategoryType != "" {
		db = db.Where(categoryFilter, f.CategoryType)
	}
	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	var n int64
	if err := db.Count(&n).Error; err != nil {
		return nil, 0, err
	}
	out := []exam.ExamSession{}
	err := db.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&out).Error
	return out, n, err
}
