package repository

import (
	exam "github.com/han/go-ecommerce/internal/exam/model"
	"github.com/han/go-ecommerce/internal/learning"
	"gorm.io/gorm"
	"time"
)

type Interface interface {
	Transaction(func(*gorm.DB) error) error
	ExpiredIDs(user uint, now time.Time) ([]uint, error)
	History(user uint, page, limit int) ([]exam.ExamSession, int64, error)
	HistoryFiltered(user uint, f learning.Filter) ([]exam.ExamSession, int64, error)
}
