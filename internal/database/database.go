package database

import (
	"log"

	"github.com/han/go-ecommerce/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB là biến toàn cục chứa kết nối database
var DB *gorm.DB

// InitDB khởi tạo kết nối và thực hiện tự động tạo bảng
func InitDB(filepath string) {
	var err error

	// Khởi tạo kết nối gorm
	DB, err = gorm.Open(sqlite.Open(filepath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Không thể kết nối đến SQLite: %v", err)
	}

	log.Println("Kết nối cơ sở dữ liệu thành công!")

	// Tự động tạo bảng (Migration)
	autoMigrate()
}
func autoMigrate() {
	log.Println("Đang tự động đồng bộ cấu trúc bảng (AutoMigrate)...")

	// Truyền mảng các model vào AutoMigrate bằng toán tử ...
	err := DB.AutoMigrate(
		getModels()...,
	)
	if err != nil {
		log.Fatalf("Lỗi AutoMigrate nghiêm trọng: %v", err)
	}

	log.Println("Tự động đồng bộ cấu trúc bảng THÀNH CÔNG!")
}
func getModels() []interface{} {
	return []interface{}{
		&models.Category{},
		&models.Question{},
		&models.Answer{},
		&models.Test{},
		&models.User{},
		&models.ExamSession{},
		&models.UserAnswer{},
	}
}
