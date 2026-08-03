package usecase

import (
	"github.com/han/go-ecommerce/internal/auth/dto"
	auth "github.com/han/go-ecommerce/internal/auth/model"
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*auth.User, error)
	GetUser(id string) (*auth.User, error)
	Login(request dto.LoginRequest) (*LoginResult, error)
}
type LoginResult struct {
	User  *auth.User
	Token string
}
