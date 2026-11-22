package auth

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	ID uint `gorm:"primaryKey"`
	//token thuộc user nào.
	UserID uint `gorm:"not null;index"`

	// Chỉ lưu SHA-256 hash, không lưu token gốc.//hash của refresh token.
	TokenHash string `gorm:"size:64;not null;uniqueIndex"`
	//thời gian hết hạn.
	ExpiresAt time.Time `gorm:"not null;index"`
	//token đã bị thu hồi chưa
	RevokedAt *time.Time `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
