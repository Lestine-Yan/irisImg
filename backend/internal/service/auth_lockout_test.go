package service

import (
	"errors"
	"testing"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/jwt"
)

const (
	testIPLocked = "203.0.113.7"
	testIPOther  = "198.51.100.4"
)

// newLockoutTestSvc 构造带可控时钟的 AuthService。
// 返回的 clock 指针可在测试中推进，用于验证惰性解锁。
func newLockoutTestSvc(t *testing.T) (*AuthService, *time.Time) {
	t.Helper()
	cfg := config.AuthConfig{
		Username:          "admin",
		Password:          "secret",
		LoginMaxFailures:  3,
		LoginLockoutHours: 168,
	}
	mgr := jwt.NewManager(config.JWTConfig{
		Secret:      "01234567890123456789012345678901",
		Issuer:      "irisImg",
		ExpireHours: 1,
	})
	svc := NewAuthService(cfg, mgr)
	clock := time.Unix(1700000000, 0)
	svc.now = func() time.Time { return clock }
	return svc, &clock
}

// TestAuthService_LoginLockout 覆盖：同一 IP 连续 3 次失败触发锁定、锁定期间正确密码也被拒、
// 到期后自动解锁、成功登录后该 IP 计数清零；并验证锁定只影响该 IP，其他 IP 不受牵连。
func TestAuthService_LoginLockout(t *testing.T) {
	svc, clock := newLockoutTestSvc(t)

	fail := func(ip string) error {
		_, err := svc.Login(&model.LoginRequest{Username: "admin", Password: "wrong"}, ip)
		return err
	}
	ok := func(ip string) error {
		_, err := svc.Login(&model.LoginRequest{Username: "admin", Password: "secret"}, ip)
		return err
	}

	for i := 1; i <= 3; i++ {
		if err := fail(testIPLocked); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("第 %d 次失败应为 ErrInvalidCredentials，实际 %v", i, err)
		}
	}

	// 触发锁定的下一次请求（即使密码正确）应直接拒绝。
	if err := ok(testIPLocked); !errors.Is(err, ErrLoginLocked) {
		t.Fatalf("锁定期间正确密码也应被拒，实际 %v", err)
	}
	remaining, locked := svc.LockoutRemaining(testIPLocked)
	if !locked || remaining != 168*time.Hour {
		t.Fatalf("剩余锁定时间 = %v (locked=%v)，期望 168h", remaining, locked)
	}

	// 关键：别的 IP 不受该 IP 锁定影响（按 IP 计数，而非全局账号）。
	if err := ok(testIPOther); err != nil {
		t.Fatalf("其他 IP 应可正常登录，实际 %v", err)
	}
	if _, locked := svc.LockoutRemaining(testIPOther); locked {
		t.Fatalf("其他 IP 不应被锁定")
	}

	// 差一秒仍未解锁。
	*clock = clock.Add(168*time.Hour - time.Second)
	if _, locked := svc.LockoutRemaining(testIPLocked); !locked {
		t.Fatalf("锁定到期前 1 秒不应解锁")
	}

	// 到期后正确密码放行（惰性解锁）。
	*clock = clock.Add(time.Second)
	if err := ok(testIPLocked); err != nil {
		t.Fatalf("锁定到期后应可登录，实际 %v", err)
	}

	// 成功登录已清零该 IP 计数：再错两次不应锁定。
	if err := fail(testIPLocked); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("解锁后第 1 次失败应返回 ErrInvalidCredentials，实际 %v", err)
	}
	if err := fail(testIPLocked); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("解锁后第 2 次失败应返回 ErrInvalidCredentials，实际 %v", err)
	}
	if err := ok(testIPLocked); err != nil {
		t.Fatalf("两次失败不应触发锁定，实际 %v", err)
	}
}

// TestAuthService_LockoutResetsOnRestart 验证锁定状态只在内存中：
// 新建实例（等价于进程重启）即恢复登录能力，这是忘记密码时的兜底恢复路径。
func TestAuthService_LockoutResetsOnRestart(t *testing.T) {
	svc, _ := newLockoutTestSvc(t)
	for i := 0; i < 3; i++ {
		_, _ = svc.Login(&model.LoginRequest{Username: "admin", Password: "wrong"}, testIPLocked)
	}
	if _, locked := svc.LockoutRemaining(testIPLocked); !locked {
		t.Fatalf("前置条件：应已进入锁定")
	}

	svc2, _ := newLockoutTestSvc(t)
	if _, err := svc2.Login(&model.LoginRequest{Username: "admin", Password: "secret"}, testIPLocked); err != nil {
		t.Fatalf("重启后的新实例应可登录，实际 %v", err)
	}
}
