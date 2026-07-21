package mapper

import (
	"github.com/han/go-ecommerce/internal/module/auth/dto/response"
	"github.com/han/go-ecommerce/internal/module/auth/model"
)

func ToUserResponse(user *model.User) response.UserResponse {
	return response.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
