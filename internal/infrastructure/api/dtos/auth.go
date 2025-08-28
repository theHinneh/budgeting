package dtos

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	TokenType    string   `json:"token_type"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	UID         string  `json:"uid"`
	Email       string  `json:"email"`
	DisplayName string  `json:"display_name"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type SessionInfo struct {
	ID         string `json:"id"`
	DeviceInfo string `json:"device_info"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at"`
	IsCurrent  bool   `json:"is_current"`
}

type RevokeSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}
