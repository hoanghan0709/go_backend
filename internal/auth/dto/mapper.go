package dto

import auth "github.com/han/go-ecommerce/internal/auth/model"

func ToUserResponse(user *auth.User) UserResponse {
	return UserResponse{
		CurrentLicenseType: user.CurrentLicenseType,
		ID:                 user.ID,
		Role:               user.Role,
		Name:               user.Name,
		Email:              user.Email,
	}
}
func ToLoginResponse(user *auth.User,
	accessToken string,
	refreshToken string) LoginResponse {
	return LoginResponse{
		CurrentLicenseType: user.CurrentLicenseType,
		Email:              user.Email,
		ID:                 user.ID,
		Name:               user.Name,
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		ExpiresIn:          30, //12 * 60 * 60,
		Role:               user.Role,
	}
}
