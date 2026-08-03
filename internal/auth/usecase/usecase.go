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

type TokenGenerator interface {
	Generate(userID uint) (string, error)
}

type UseCase struct {
	repository     repo.UserRepository
	tokenGenerator TokenGenerator
}

func New(repository repo.UserRepository,
	tokenGenerator TokenGenerator) *UseCase {
	return &UseCase{repository: repository,
		tokenGenerator: tokenGenerator}
}

func (s *UseCase) Register(req dto.RegisterRequest) (*auth.User,
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

func (s *UseCase) GetUser(id string) (*auth.User, error) {
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

func (s *UseCase) Login(req dto.LoginRequest) (*LoginResult, error) {
	user, err := s.repository.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	); err != nil {
		return nil, commonError.ErrInvalidCredentials
	}

	tokenString, err := s.tokenGenerator.Generate(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	return &LoginResult{
		User:  user,
		Token: tokenString,
	}, nil
}
