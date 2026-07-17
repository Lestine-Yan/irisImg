package config

import "testing"

// TestConfig_Validate 覆盖启动安全校验:release 模式强制拒绝默认/空口令与弱 JWT 密钥,
// debug/test 放过(开发开箱即跑)。闭合「拷贝模板未改口令即上线」的攻击链。
func TestConfig_Validate(t *testing.T) {
	strongSecret := "01234567890123456789012345678901abcdef" // 40 字符,>=32
	defaultSecret := "please-change-me-to-a-long-random-string"

	cases := []struct {
		name             string
		mode, user, pass string
		secret           string
		wantErr          bool
	}{
		{"release 默认口令+默认密钥(应拒)", "release", "admin", "admin123", defaultSecret, true},
		{"release 改口令+强密钥(应过)", "release", "admin", "real-pass-9x", strongSecret, false},
		{"debug 默认口令(应过,开发友好)", "debug", "admin", "admin123", defaultSecret, false},
		{"test 模式放过", "test", "admin", "admin123", defaultSecret, false},
		{"release 短密钥(应拒)", "release", "admin", "real-pass-9x", "short", true},
		{"release 空口令(应拒)", "release", "admin", "", strongSecret, true},
		{"release 空用户名(应拒)", "release", "", "real-pass-9x", strongSecret, true},
		{"release 改口令但密钥仍默认(应拒)", "release", "admin", "real-pass-9x", defaultSecret, true},
		{"release 用户名 admin 合法(密码非默认,应过)", "release", "admin", "real-pass-9x", strongSecret, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{Mode: c.mode},
				Auth: AuthConfig{
					Username: c.user,
					Password: c.pass,
					JWT:      JWTConfig{Secret: c.secret},
				},
			}
			err := cfg.Validate()
			gotErr := err != nil
			if gotErr != c.wantErr {
				t.Fatalf("Validate() err=%v, wantErr=%v", err, c.wantErr)
			}
		})
	}
}
