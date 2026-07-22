package usecase

// type Repository interface{}

// type Usecase struct {
// 	repo Repository
// }

// func New(repo Repository) *Usecase {
// 	return &Usecase{repo: repo}
// }
// package service

import (
	"errors"
	"fmt"

	commonError "github.com/han/go-ecommerce/internal/auth/common_error"
	"github.com/han/go-ecommerce/internal/auth/dto"
	auth "github.com/han/go-ecommerce/internal/auth/model"
	repo "github.com/han/go-ecommerce/internal/auth/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	repository repo.UserRepository
}

func New(repository repo.UserRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Register(req dto.RegisterRequest) (*auth.User,
	error) {
	_, err := s.repository.FindByEmail(req.Email)
	if err == nil {
		return nil, commonError.ErrEmailAlreadyExists
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

	user := &auth.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	if err := s.repository.Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *Service) GetUser(id string) (*auth.User, error) {
	data, err := s.repository.GetUser(id)
	if err != nil {
		return nil, err
	}
	user := &auth.User{
		Name:  data.Name,
		Email: data.Email,
	}
	return user, nil
}
