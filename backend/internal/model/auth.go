package model

// LoginRequest 是登录接口的入参 DTO。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 是登录成功后的出参 DTO。
type LoginResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"` // 固定为 "Bearer"
	ExpiresAt int64  `json:"expires_at"` // 过期时间，Unix 秒
}
