package repository

import (
	exam "github.com/han/go-ecommerce/internal/exam/model"
	"gorm.io/gorm"
	"time"
)

type Interface interface {
	Transaction(func(*gorm.DB) error) error
	ExpiredIDs(user uint, now time.Time) ([]uint, error)
	History(user uint, page, limit int) ([]exam.ExamSession, int64, error)
}
