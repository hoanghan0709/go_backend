package database

import (
	module "github.com/han/go-ecommerce/internal/module"
	auth "github.com/han/go-ecommerce/internal/module/auth/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB(filepath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := autoMigrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

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
