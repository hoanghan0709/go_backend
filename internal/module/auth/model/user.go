package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	// ID       uint `gorm:"primaryKey"`
	Name     string
	Email    string
	Password string
}

// //User là database.//DTO là API
// //Database ID CreatedAt UpdatedAt Password =====> Flutter đâu cần gửi CreatedAt. Nên phải tách.
