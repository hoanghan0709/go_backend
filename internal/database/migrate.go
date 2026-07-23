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
	user "github.com/han/go-ecommerce/internal/auth/model"
	category "github.com/han/go-ecommerce/internal/category/model"
	exam "github.com/han/go-ecommerce/internal/exam/model"
	question "github.com/han/go-ecommerce/internal/question/model"
	test "github.com/han/go-ecommerce/internal/test/model"
	"gorm.io/gorm"
)

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(getModels()...)
}

func getModels() []interface{} {
	return []interface{}{
		&category.Category{},
		&question.Question{},
		&question.Answer{},
		&test.Test{},
		&user.User{},
		&exam.ExamSession{},
		&exam.UserAnswer{},
	}
}
