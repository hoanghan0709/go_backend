package database

// Khi có thêm Feature
// Ví dụ tạo
// make new name=exam
// và có
// exam/
//     model/
//         exam.go
// chỉ cần thêm
// examModel "github.com/han/go-backend/internal/exam/model"
// và
// err := db.AutoMigrate(
// 	&authModel.User{},
// 	&categoryModel.Category{},
// 	&questionModel.Question{},
// 	&examModel.Exam{},<<<============ IMPORTANT
// )

import (
	"fmt"
	"github.com/han/go-ecommerce/internal/answer/model"
	auth "github.com/han/go-ecommerce/internal/auth/model"
	category "github.com/han/go-ecommerce/internal/category/model"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	"github.com/han/go-ecommerce/internal/learning"
	question "github.com/han/go-ecommerce/internal/question/model"
	roadsign "github.com/han/go-ecommerce/internal/roadsign/model"
	test "github.com/han/go-ecommerce/internal/test/model"
	"gorm.io/gorm"
)

func autoMigrate(db *gorm.DB) (err error) {
	// SQLite rebuilds tables when adding foreign keys. Disable enforcement only
	// before the migration transaction (one connection, server not started yet).
	// Validate the entire graph before commit and always restore enforcement.
	if err = db.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		return err
	}
	defer func() {
		if restore := db.Exec("PRAGMA foreign_keys = ON").Error; restore != nil && err == nil {
			err = restore
		}
	}()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(getModels()...); err != nil {
			return err
		}
		rows, err := tx.Raw("PRAGMA foreign_key_check").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()
		if rows.Next() {
			return fmt.Errorf("migration would leave invalid foreign keys; transaction rolled back")
		}
		return rows.Err()
	})
}

func getModels() []interface{} {
	return []interface{}{
		&learning.PracticeAttempt{},
		&learning.CrawlRecord{},
		&category.Category{},
		&roadsign.RoadSign{},
		&learning.ExamTemplate{},
		&question.Question{},
		&model.Answer{},
		&test.Test{},
		&test.TestQuestion{},
		&exam.ExamSession{},
		&exam.UserAnswer{},
		&auth.RefreshToken{},
		&auth.User{},
	}
}
