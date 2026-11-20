package service

import (
	"crypto/subtle"
	"errors"
	"sync"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/jwt"
)

// ErrInvalidCredentials 表示用户名或密码错误。
// 不区分到底是哪一项错，避免泄露用户名是否存在。
var ErrInvalidCredentials = errors.New("invalid username or password")

// ErrLoginLocked 表示登录因该客户端 IP 连续失败次数过多而处于锁定冷却期。
// 与 ErrInvalidCredentials 区分：前者对应 429（稍后再试），后者对应 401（凭据错误）。
var ErrLoginLocked = errors.New("login temporarily locked")

// unknownClientIP 是拿不到客户端 IP 时的兜底键（正常请求不会命中）。
const unknownClientIP = "unknown"

// maxLoginStates 限制失败计数表的条目上限，防止大量不同来源 IP 造成内存膨胀。
// 达到上限时优先清理已解锁且长时间不活跃的条目，仍满则淘汰最久未活跃的一条。
const maxLoginStates = 4096

// loginState 是单个客户端 IP 的登录失败状态。
type loginState struct {
	failures    int
	lockedUntil time.Time
	lastSeen    time.Time
}

// AuthService 处理登录鉴权与 JWT 签发。
// 私人图床仅支持单用户，账号信息直接来自配置文件。
//
// 除凭据校验外，本服务还承担「登录失败锁定」：按客户端 IP 计数，
// 同一 IP 连续失败达到阈值后锁定一段时间，其他 IP 不受影响。
// 锁定状态只存进程内存，重启即清零——这也是所有者忘记密码时的兜底恢复路径。
type AuthService struct {
	cfg    config.AuthConfig
	jwtMgr *jwt.Manager

	// mu 保护 states。
	mu     sync.Mutex
	states map[string]*loginState
	// now 允许测试注入时钟；生产固定为 time.Now。
	now func() time.Time
}

// NewAuthService 通过依赖注入构造 AuthService。
// 登录锁定参数缺省（<=0）时回退为 3 次失败 / 168 小时（7 天）。
func NewAuthService(cfg config.AuthConfig, m *jwt.Manager) *AuthService {
	if cfg.LoginMaxFailures <= 0 {
		cfg.LoginMaxFailures = 3
	}
	if cfg.LoginLockoutHours <= 0 {
		cfg.LoginLockoutHours = 168
	}
	return &AuthService{
		cfg:    cfg,
		jwtMgr: m,
		states: make(map[string]*loginState),
		now:    time.Now,
	}
}

// Login 校验用户名/密码，通过则签发 JWT。
// 用 subtle.ConstantTimeCompare 同时比对用户名与密码，
// 既防时序攻击，也避免根据响应耗时区分账号是否存在。
//
// clientIP 为鉴权中间件给出的客户端地址（受 server.trusted_proxies 约束）。
// 失败按该 IP 累加；达到阈值后该 IP 进入锁定，锁定期间即使密码正确也返回 ErrLoginLocked。
func (s *AuthService) Login(req *model.LoginRequest, clientIP string) (*model.LoginResponse, error) {
	clientIP = normalizeClientIP(clientIP)

	if _, locked := s.LockoutRemaining(clientIP); locked {
		return nil, ErrLoginLocked
	}

	if !s.verify(req.Username, req.Password) {
		s.recordFailure(clientIP)
		return nil, ErrInvalidCredentials
	}

	// 成功即清除该 IP 的失败计数（不影响其他 IP）。
	s.clearFailures(clientIP)

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
// 复用 Login 的常量时间比对逻辑，但不签发 token、也不参与登录失败计数；
// 失败返回 ErrInvalidCredentials。
func (s *AuthService) VerifyCredentials(username, password string) error {
	if !s.verify(username, password) {
		return ErrInvalidCredentials
	}
	return nil
}

// LockoutRemaining 返回指定 IP 的剩余锁定时长；处于锁定期时第二个返回值为 true。
// 若锁定已过期，会顺带清除该 IP 的状态（惰性解锁）。
func (s *AuthService) LockoutRemaining(clientIP string) (time.Duration, bool) {
	clientIP = normalizeClientIP(clientIP)

	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.states[clientIP]
	if !ok || st.lockedUntil.IsZero() {
		return 0, false
	}
	remaining := st.lockedUntil.Sub(s.now())
	if remaining <= 0 {
		delete(s.states, clientIP)
		return 0, false
	}
	return remaining, true
}

// recordFailure 累加该 IP 的失败次数；达到阈值时开始锁定并清零计数。
// 这样解锁后重新拥有完整的尝试次数。
func (s *AuthService) recordFailure(clientIP string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	st, ok := s.states[clientIP]
	if !ok {
		// 仅在新增条目时做容量控制，避免每次请求都遍历。
		if len(s.states) >= maxLoginStates {
			s.evictLocked(now)
		}
		st = &loginState{}
		s.states[clientIP] = st
	}
	st.lastSeen = now
	st.failures++
	if st.failures < s.cfg.LoginMaxFailures {
		return
	}
	st.failures = 0
	st.lockedUntil = now.Add(time.Duration(s.cfg.LoginLockoutHours) * time.Hour)
}

// clearFailures 在登录成功后删除该 IP 的失败状态。
func (s *AuthService) clearFailures(clientIP string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, clientIP)
}

// evictLocked 在表满时淘汰条目：先清理「已解锁且 1 小时未活跃」的，
// 仍满则删除 lastSeen 最早的一条。调用方需持有 mu。
func (s *AuthService) evictLocked(now time.Time) {
	for ip, st := range s.states {
		unlocked := st.lockedUntil.IsZero() || !now.Before(st.lockedUntil)
		if unlocked && now.Sub(st.lastSeen) > time.Hour {
			delete(s.states, ip)
		}
	}
	if len(s.states) < maxLoginStates {
		return
	}

	var oldestIP string
	var oldest time.Time
	for ip, st := range s.states {
		if oldestIP == "" || st.lastSeen.Before(oldest) {
			oldestIP, oldest = ip, st.lastSeen
		}
	}
	if oldestIP != "" {
		delete(s.states, oldestIP)
	}
}

// normalizeClientIP 统一空值兜底，避免空字符串成为共享键。
func normalizeClientIP(ip string) string {
	if ip == "" {
		return unknownClientIP
	}
	return ip
}

// verify 用 subtle.ConstantTimeCompare 同时比对用户名与密码，
// 既防时序攻击，也避免根据响应耗时区分账号是否存在。
func (s *AuthService) verify(username, password string) bool {
	usernameOK := subtle.ConstantTimeCompare([]byte(username), []byte(s.cfg.Username)) == 1
	passwordOK := subtle.ConstantTimeCompare([]byte(password), []byte(s.cfg.Password)) == 1
	return usernameOK && passwordOK
}
