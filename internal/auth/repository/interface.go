package repository

import (
	auth "github.com/han/go-ecommerce/internal/auth/model"
)

type UserRepository interface {
	Create(user *auth.User) error
	FindByEmail(email string) (*auth.User, error)
	GetUser(id string) (*auth.User, error)
}
