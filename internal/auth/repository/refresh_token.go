package repository

import (
	"errors"
	"time"

	auth "github.com/han/go-ecommerce/internal/auth/model"
	"gorm.io/gorm"
)

var ErrRefreshTokenAlreadyRevoked = errors.New("Err Refresh Token Already Revoked")

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(
	refreshToken *auth.RefreshToken,
) error {
	return r.db.Create(refreshToken).Error
}

func (r *refreshTokenRepository) FindByHash(
	tokenHash string,
) (*auth.RefreshToken, error) {
	var refreshToken auth.RefreshToken

	err := r.db.
		Where("token_hash = ?", tokenHash).
		First(&refreshToken).
		Error

	if err != nil {
		return nil, err
	}

	return &refreshToken, nil
}

func (r *refreshTokenRepository) Revoke(tokenHash string) error {
	now := time.Now()

	result := r.db.Model(&auth.RefreshToken{}).
		Where(
			"token_hash = ? AND revoked_at IS NULL",
			tokenHash,
		).
		Update("revoked_at", now)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrRefreshTokenAlreadyRevoked
	}
	return nil
}

func (r *refreshTokenRepository) Rotate(
	oldHash string,
	newToken *auth.RefreshToken,
) error {

	return r.db.Transaction(func(db *gorm.DB) error {

		now := time.Now()

		result := db.Model(&auth.RefreshToken{}).
			Where(
				"token_hash = ? AND revoked_at IS NULL",
				oldHash,
			).
			Update("revoked_at", now)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrRefreshTokenAlreadyRevoked
		}

		if err := db.Create(newToken).Error; err != nil {
			return err
		}

		return nil
	})
}
