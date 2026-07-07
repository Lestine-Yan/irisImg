package service

import (
	"errors"
	"testing"

	"github.com/Lestine-Yan/irisImg/backend/config"
)

// TestAuthService_VerifyCredentials 覆盖敏感操作的密码二次确认：
// 正确账号密码通过；用户名或密码任一错误均返回 ErrInvalidCredentials（不区分以防空号枚举）。
func TestAuthService_VerifyCredentials(t *testing.T) {
	svc := NewAuthService(config.AuthConfig{Username: "admin", Password: "secret"}, nil)

	if err := svc.VerifyCredentials("admin", "secret"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := svc.VerifyCredentials("admin", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong password, got %v", err)
	}
	if err := svc.VerifyCredentials("nope", "secret"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong username, got %v", err)
	}
}
