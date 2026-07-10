package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {

	return &Service{

		repository: repository,
	}
}

func (s *Service) Register(req RegisterRequest) error {

	user, err := s.repository.FindByEmail(req.Email)

	if err == nil && user != nil {

		return errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(

		[]byte(req.Password),

		bcrypt.DefaultCost,
	)

	if err != nil {

		return err
	}

	user = &User{

		Name: req.Name,

		Email: req.Email,

		Password: string(hashedPassword),
	}

	return s.repository.Create(user)
}
