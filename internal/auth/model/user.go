package auth

import "gorm.io/gorm"

type User struct {
	gorm.Model
	// ID       uint `gorm:"primaryKey"`
	CurrentLicenseType string `gorm:"size:10" json:"current_license_type"`
	Name               string
	Email              string
	Password           string         `json:"-"`
	Role               string         `gorm:"not null;default:learner" json:"role"`
	RefreshTokens      []RefreshToken `gorm:"foreignKey:UserID"`
}

// //User là database.//DTO là API
// //Database ID CreatedAt UpdatedAt Password =====> Flutter đâu cần gửi CreatedAt. Nên phải tách.
