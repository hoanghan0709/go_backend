package service

import (
	"errors"
	"testing"

	dto "github.com/han/go-ecommerce/internal/module/auth/dto/request"
	"github.com/han/go-ecommerce/internal/module/auth/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type repositoryStub struct {
	findUser *model.User
	findErr  error
	create   func(*model.User) error
}

func (r *repositoryStub) FindByEmail(string) (*model.User, error) {
	return r.findUser, r.findErr
}

func (r *repositoryStub) Create(user *model.User) error {
	if r.create != nil {
		return r.create(user)
	}
	return nil
}

func TestRegisterCreatesUserWithHashedPassword(t *testing.T) {
	repository := &repositoryStub{
		findErr: gorm.ErrRecordNotFound,
		create: func(user *model.User) error {
			user.ID = 7
			return nil
		},
	}
	service := NewService(repository)

	user, err := service.Register(dto.RegisterRequest{
		Name: "Nguyen Van An", Email: "an@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.ID != 7 || user.Email != "an@example.com" {
		t.Fatalf("Register() user = %#v", user)
	}
	if user.Password == "password123" {
		t.Fatal("Register() stored the plain-text password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123")); err != nil {
		t.Fatalf("password hash does not match: %v", err)
	}
}

func TestRegisterRejectsExistingEmail(t *testing.T) {
	service := NewService(&repositoryStub{findUser: &model.User{}})

	_, err := service.Register(dto.RegisterRequest{Email: "an@example.com"})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("Register() error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestRegisterReturnsDatabaseLookupError(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	service := NewService(&repositoryStub{findErr: databaseErr})

	_, err := service.Register(dto.RegisterRequest{Email: "an@example.com"})
	if !errors.Is(err, databaseErr) {
		t.Fatalf("Register() error = %v, want wrapped database error", err)
	}
}
