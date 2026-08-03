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
