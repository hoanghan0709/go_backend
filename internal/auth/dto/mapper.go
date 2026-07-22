package dto

import auth "github.com/han/go-ecommerce/internal/auth/model"

func ToUserResponse(user *auth.User) UserResponse {
	return UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
