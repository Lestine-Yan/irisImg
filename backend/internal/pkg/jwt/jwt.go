// Package jwt 封装 JWT 的签发与解析。
// 业务层只通过 Manager 与 JWT 打交道，方便后续替换签名算法或库实现。
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Claims 是本项目的自定义 JWT 载荷。
type Claims struct {
	Username string `json:"username"`
	jwtv5.RegisteredClaims
}

// Manager 负责使用固定密钥签发与校验 token。
type Manager struct {
	secret []byte
	issuer string
	expire time.Duration
}

// NewManager 根据配置构造一个 Manager。
// expire_hours 配置缺失（<=0）时回退到 24 小时。
func NewManager(cfg config.JWTConfig) *Manager {
	expire := time.Duration(cfg.ExpireHours) * time.Hour
	if expire <= 0 {
		expire = 24 * time.Hour
	}
	return &Manager{
		secret: []byte(cfg.Secret),
		issuer: cfg.Issuer,
		expire: expire,
	}
}

// Issue 为给定用户名签发一个 token，并返回过期时间（Unix 秒）。
func (m *Manager) Issue(username string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(m.expire)

	claims := Claims{
		Username: username,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   username,
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign jwt: %w", err)
	}
	return signed, expiresAt.Unix(), nil
}

// Parse 校验 token 字符串，成功时返回解析后的 Claims。
// 显式校验签名方法为 HS256，避免 alg=none 等攻击。
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	parsed, err := jwtv5.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtv5.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
