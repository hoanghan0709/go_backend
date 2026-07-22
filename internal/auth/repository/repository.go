package repository

import (
	auth "github.com/han/go-ecommerce/internal/auth/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(user *auth.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) FindByEmail(email string) (*auth.User, error) {
	var user auth.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *Repository) GetUser(id string) (*auth.User, error) {
	var user auth.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
