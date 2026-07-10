package auth

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {

	return &Repository{

		db: db,
	}
}
func (r *Repository) Create(user *User) error {

	return r.db.Create(user).Error
}
func (r *Repository) FindByEmail(email string) (*User, error) {

	var user User

	err := r.db.Where("email = ?", email).First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Luồng hiện tại Handler -> Service -> Repository -> db.Create() -> SQLite
//Repository có được Hash Password không? Không. Repository chỉ biết Database Không biết JWT HTTP Password Business
