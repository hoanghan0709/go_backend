package service

import (
	"errors"
	"fmt"

	dto "github.com/han/go-ecommerce/internal/module/auth/dto/request"
	"github.com/han/go-ecommerce/internal/module/auth/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type UserRepository interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	GetUser(id string) (*model.User, error)
}

type Service struct {
	repository UserRepository
}

func NewService(repository UserRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Register(req dto.RegisterRequest) (*model.User, error) {
	_, err := s.repository.FindByEmail(req.Email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	if err := s.repository.Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *Service) GetUser(id string) (*model.User, error) {
	data, err := s.repository.GetUser(id)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Name:  data.Name,
		Email: data.Email,
	}
	return user, nil
}
