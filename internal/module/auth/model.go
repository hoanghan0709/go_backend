package auth

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name     string
	Email    string `gorm:"unique"`
	Password string
}

//User là database.//DTO là API
//Database ID CreatedAt UpdatedAt Password =====> Flutter đâu cần gửi CreatedAt. Nên phải tách.
