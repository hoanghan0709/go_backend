package usecase

import (
	"github.com/han/go-ecommerce/internal/auth/dto"
	auth "github.com/han/go-ecommerce/internal/auth/model"
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*auth.User, error)
	GetUser(id string) (*auth.User, error)
	Login(request dto.LoginRequest) (*LoginResult, error)
	Refresh(refreshToken string) (*TokenPair, error)
}
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type LoginResult struct {
	User *auth.User
	TokenPair
}
type TokenService interface {
	GenerateAccessToken(userID uint) (string, error)
	GenerateRefreshToken() (string, error)
	HashRefreshToken(refreshToken string) string
}
