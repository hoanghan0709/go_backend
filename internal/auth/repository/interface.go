package repository

import (
	"github.com/han/go-ecommerce/internal/auth/dto"
	auth "github.com/han/go-ecommerce/internal/auth/model"
)

type UserRepository interface {
	Create(user *auth.User) error
	FindByEmail(email string) (*auth.User, error)
	GetUser(id string) (*auth.User, error)
	Login(req dto.LoginRequest) (*auth.User, error)
}

type RefreshTokenRepository interface {
	Create(refreshToken *auth.RefreshToken) error
	FindByHash(tokenHash string) (*auth.RefreshToken, error)
	Revoke(tokenHash string) error
	Rotate(
		oldHash string,
		newToken *auth.RefreshToken,
	) error
}
