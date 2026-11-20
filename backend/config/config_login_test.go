package config

import "testing"

// TestApplyDefaults_LoginGuard 覆盖登录防爆破配置的缺省兜底与显式值保留。
func TestApplyDefaults_LoginGuard(t *testing.T) {
	t.Run("零值补默认", func(t *testing.T) {
		c := &Config{}
		c.ApplyDefaults()
		if c.Auth.LoginMaxFailures != 3 {
			t.Errorf("login_max_failures = %d, want 3", c.Auth.LoginMaxFailures)
		}
		if c.Auth.LoginLockoutHours != 168 {
			t.Errorf("login_lockout_hours = %d, want 168", c.Auth.LoginLockoutHours)
		}
	})

	t.Run("显式值保留", func(t *testing.T) {
		c := &Config{Auth: AuthConfig{LoginMaxFailures: 5, LoginLockoutHours: 1}}
		c.ApplyDefaults()
		if c.Auth.LoginMaxFailures != 5 {
			t.Errorf("login_max_failures = %d, want 5", c.Auth.LoginMaxFailures)
		}
		if c.Auth.LoginLockoutHours != 1 {
			t.Errorf("login_lockout_hours = %d, want 1", c.Auth.LoginLockoutHours)
		}
	})
}
