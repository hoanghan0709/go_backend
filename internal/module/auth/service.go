package auth

import (
	// "errors"

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

type BaseResponse struct {
	StatusCode int         `json:"statusCode"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
}

func (s *Service) Register(req RegisterRequest) BaseResponse {

	user, err := s.repository.FindByEmail(req.Email)

	if err == nil && user != nil {
		return BaseResponse{
			StatusCode: 400,
			Message:    "Email already exists",
			Data:       nil,
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(

		[]byte(req.Password),

		bcrypt.DefaultCost,
	)

	if err != nil {

		return BaseResponse{
			StatusCode: 500,
			Message:    "Cannot hash password",
			Data:       nil,
		}
	}

	user = &User{

		Name: req.Name,

		Email: req.Email,

		Password: string(hashedPassword),
	}
	if err := s.repository.Create(user); err != nil {
		return BaseResponse{
			StatusCode: 500,
			Message:    "Cannot create user",
			Data:       nil,
		}
	}
	return BaseResponse{
		StatusCode: 201,
		Message:    "Register success",
		Data: map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	}
}
