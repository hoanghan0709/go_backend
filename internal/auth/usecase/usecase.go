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
	"time"

	commonError "github.com/han/go-ecommerce/internal/auth/common_error"
	"github.com/han/go-ecommerce/internal/auth/dto"
	auth "github.com/han/go-ecommerce/internal/auth/model"
	repo "github.com/han/go-ecommerce/internal/auth/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	ErrExpiredRefreshToken = errors.New("refresh token expired")
)

type UseCase struct {
	repository             repo.UserRepository
	refreshTokenRepository repo.RefreshTokenRepository
	tokenService           TokenService
	refreshTokenTTL        time.Duration
}

func New(
	userRepository repo.UserRepository,
	refreshTokenRepository repo.RefreshTokenRepository,
	tokenService TokenService,
	refreshTokenTTL time.Duration,
) *UseCase {
	return &UseCase{
		repository:             userRepository,
		refreshTokenRepository: refreshTokenRepository,
		tokenService:           tokenService,
		refreshTokenTTL:        refreshTokenTTL,
	}
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
			return nil, commonError.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	); err != nil {
		return nil, commonError.ErrInvalidCredentials
	}

	// tokenString, err := s.tokenService.GenerateAccessToken(user.ID)
	// if err != nil {
	// 	return nil, fmt.Errorf("generate access token: %w", err)
	// }

	// return &LoginResult{
	// 	User: user, TokenPair: TokenPair(),
	// }, nil
	accessToken, err :=
		s.tokenService.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	refreshToken, err :=
		s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshTokenHash :=
		s.tokenService.HashRefreshToken(refreshToken)
	record := &auth.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepository.Create(record); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}
	return &LoginResult{
		User: user,
		TokenPair: TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil

}

func (s *UseCase) Refresh(
	refreshToken string,
) (*TokenPair, error) {

	oldHash :=
		s.tokenService.HashRefreshToken(refreshToken)

	storedToken, err :=
		s.refreshTokenRepository.FindByHash(oldHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidRefreshToken
		}

		return nil, fmt.Errorf(
			"find refresh token: %w",
			err,
		)
	}

	if storedToken.RevokedAt != nil {
		return nil, ErrInvalidRefreshToken
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, ErrExpiredRefreshToken
	}

	newAccessToken, err :=
		s.tokenService.GenerateAccessToken(
			storedToken.UserID,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"generate access token: %w",
			err,
		)
	}

	newRefreshToken, err :=
		s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf(
			"generate refresh token: %w",
			err,
		)
	}

	newHash :=
		s.tokenService.HashRefreshToken(
			newRefreshToken,
		)

	// if err := s.refreshTokenRepository.Revoke(oldHash); err != nil {
	// 	return nil, fmt.Errorf(
	// 		"revoke old refresh token: %w",
	// 		err,
	// 	)
	// }

	newRecord := &auth.RefreshToken{
		UserID:    storedToken.UserID,
		TokenHash: newHash,
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepository.Rotate(
		oldHash,
		newRecord,
	); err != nil {
		return nil, fmt.Errorf(
			"rotate refresh token: %w",
			err,
		)
	}

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
