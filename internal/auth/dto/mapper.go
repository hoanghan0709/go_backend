package dto

import auth "github.com/han/go-ecommerce/internal/auth/model"

func ToUserResponse(user *auth.User) UserResponse {
	return UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
func ToLoginResponse(user *auth.User,
	accessToken string,
	refreshToken string) LoginResponse {
	return LoginResponse{
		Email:        user.Email,
		ID:           user.ID,
		Name:         user.Name,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    12 * 60 * 60,
	}
}
