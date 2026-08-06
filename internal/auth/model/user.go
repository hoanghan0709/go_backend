package auth

import "gorm.io/gorm"

type User struct {
	gorm.Model
	// ID       uint `gorm:"primaryKey"`
	Name          string
	Email         string
	Password      string
	RefreshTokens []RefreshToken `gorm:"foreignKey:UserID"`
}

// //User là database.//DTO là API
// //Database ID CreatedAt UpdatedAt Password =====> Flutter đâu cần gửi CreatedAt. Nên phải tách.
