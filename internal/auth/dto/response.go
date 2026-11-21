package dto

type UserResponse struct {
	CurrentLicenseType string `json:"current_license_type"`
	ID                 uint   `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	AvatarURL          string `json:"avatar_url,omitempty"`
	Role               string `json:"role,omitempty"`
}

type LoginResponse struct {
	CurrentLicenseType string `json:"current_license_type"`
	ID                 uint   `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
	ExpiresIn          int64  `json:"expires_in"`
	Role               string `json:"role,omitempty"`
}
