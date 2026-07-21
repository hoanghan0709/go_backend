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
	module "github.com/han/go-ecommerce/internal/module"
	auth "github.com/han/go-ecommerce/internal/module/auth/model"
	"gorm.io/gorm"
)

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(getModels()...)
}

func getModels() []interface{} {
	return []interface{}{
		&module.Category{},
		&module.Question{},
		&module.Answer{},
		&module.Test{},
		&auth.User{},
		&module.ExamSession{},
		&module.UserAnswer{},
	}
}
