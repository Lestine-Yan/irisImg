package service

import (
	"crypto/subtle"
	"errors"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/jwt"
)

// ErrInvalidCredentials 表示用户名或密码错误。
// 不区分到底是哪一项错，避免泄露用户名是否存在。
var ErrInvalidCredentials = errors.New("invalid username or password")

// AuthService 处理登录鉴权与 JWT 签发。
// 私人图床仅支持单用户，账号信息直接来自配置文件。
type AuthService struct {
	cfg    config.AuthConfig
	jwtMgr *jwt.Manager
}

// NewAuthService 通过依赖注入构造 AuthService。
func NewAuthService(cfg config.AuthConfig, m *jwt.Manager) *AuthService {
	return &AuthService{cfg: cfg, jwtMgr: m}
}

// Login 校验用户名/密码，通过则签发 JWT。
// 用 subtle.ConstantTimeCompare 同时比对用户名与密码，
// 既防时序攻击，也避免根据响应耗时区分账号是否存在。
func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	if !s.verify(req.Username, req.Password) {
		return nil, ErrInvalidCredentials
	}

	token, expiresAt, err := s.jwtMgr.Issue(s.cfg.Username)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresAt: expiresAt,
	}, nil
}

// VerifyCredentials 校验用户名 / 密码，用于吊销 / 删除密钥等敏感操作的二次确认。
// 复用 Login 的常量时间比对逻辑，但不签发 token；失败返回 ErrInvalidCredentials。
func (s *AuthService) VerifyCredentials(username, password string) error {
	if !s.verify(username, password) {
		return ErrInvalidCredentials
	}
	return nil
}

// verify 用 subtle.ConstantTimeCompare 同时比对用户名与密码，
// 既防时序攻击，也避免根据响应耗时区分账号是否存在。
func (s *AuthService) verify(username, password string) bool {
	usernameOK := subtle.ConstantTimeCompare([]byte(username), []byte(s.cfg.Username)) == 1
	passwordOK := subtle.ConstantTimeCompare([]byte(password), []byte(s.cfg.Password)) == 1
	return usernameOK && passwordOK
}
