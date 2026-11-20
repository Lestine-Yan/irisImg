# `internal/service/auth.go`

认证业务逻辑层。控制器把入参原样交过来，本层负责「凭据是否正确」「该签什么 token」这两件事。

## 类型与变量

### `AuthService`

```go
type AuthService struct {
    cfg    config.AuthConfig
    jwtMgr *jwt.Manager

    mu     sync.Mutex                 // 保护 states
    states map[string]*loginState     // 客户端 IP -> 失败计数 / 锁定截止 / 最近活跃
    now    func() time.Time           // 可注入时钟，测试用；生产为 time.Now
}

type loginState struct {
    failures    int
    lockedUntil time.Time
    lastSeen    time.Time
}
```

- 持有配置（用户名/密码）与 JWT 管理器，由 [`router`](../router/router.md) 注入。
- 除按 IP 的失败计数 / 锁定状态（`mu` 保护）外无可变状态；并发安全。状态表上限 `maxLoginStates = 4096`，写满时先清理已解锁且 1 小时不活跃的条目，仍满则淘汰最久未活跃的，避免任意 IP 撑爆内存。锁定只存内存，重启清零。

### `ErrInvalidCredentials`

```go
var ErrInvalidCredentials = errors.New("invalid username or password")
```

- 唯一对外暴露的「业务错误」哨兵值，控制器用 `errors.Is` 判别后映射到 401。
- 不区分「用户名错」还是「密码错」，避免暴力猜账号。

### `ErrLoginLocked`

```go
var ErrLoginLocked = errors.New("login temporarily locked")
```

- 表示登录处于锁定冷却期（连续失败达到阈值）。控制器据此映射到 **429 + `Retry-After`**。
- 锁定期间即使用户名/密码正确也返回该错误，且不进行常量时间比对。

## 函数

### `NewAuthService(cfg config.AuthConfig, m *jwt.Manager) *AuthService`

普通构造器，没有副作用。`LoginMaxFailures` / `LoginLockoutHours` 为 0 或负数时分别回退为 3 / 168。

### `Login(req *model.LoginRequest) (*model.LoginResponse, error)`

```go
clientIP = normalizeClientIP(clientIP)
if _, locked := s.LockoutRemaining(clientIP); locked {   // ① 该 IP 锁定期间直接拒绝
    return nil, ErrLoginLocked
}
usernameOK := subtle.ConstantTimeCompare(...) == 1
passwordOK := subtle.ConstantTimeCompare(...) == 1
if !usernameOK || !passwordOK {
    s.recordFailure(clientIP)                    // ② 累加该 IP 失败；达阈值则写入 lockedUntil
    return nil, ErrInvalidCredentials
}
s.clearFailures(clientIP)                        // ③ 成功清零该 IP 的计数
token, expiresAt, err := s.jwtMgr.Issue(s.cfg.Username)
```

关键点：

- **`crypto/subtle.ConstantTimeCompare`**：常量时间比较，避免依据耗时区分用户名是否存在或密码前缀是否正确。
- **同时比对**用户名和密码：即使用户名错也走完比对，进一步消除时序差异。
- 用配置里的 `cfg.Username`（而非请求里的 `req.Username`）签发 token——多此一举但更稳：避免在配置使用大小写不同写法时出现「输入与签发不一致」的怪问题；当前两者必然相等。
- jwt 签发失败直接把 err 透传给上层，由控制器映射成 500。

### `LockoutRemaining(clientIP string) (time.Duration, bool)`

返回该 IP 的剩余锁定时间与是否处于锁定。若锁定已过期，会惰性删除该 IP 的状态并返回 `(0, false)`。控制器用它生成 `Retry-After`。

### `VerifyCredentials(username, password string) error`

用于吊销 / 删除密钥等敏感操作的**二次确认**：复用 `Login` 的常量时间比对逻辑（抽到私有 `verify` 方法共用），但不签发 token；失败返回 `ErrInvalidCredentials`。控制器据此映射成 403（而非 401，避免触发前端 `useApi` 的全局登出）。

## 与其它包的关系

```
api.AuthAPI ──► service.AuthService ──► pkg/jwt.Manager
                       │
                       └─► config.AuthConfig (来自 router 注入)
```

## 修改建议

- 切换到 bcrypt 哈希校验只动两处：`AuthConfig` 加 `password_hash` 字段、`Login` 改用 `bcrypt.CompareHashAndPassword`；外部 API 与中间件都不用改。
- 登录失败计数 / 锁定已落在 service 层（单实例内存实现，按 IP 隔离）；如需跨实例共享，把 `states` 换成 Redis 计数器，控制器与中间件的契约（`ErrLoginLocked` + `LockoutRemaining`）不用变。
- 锁定策略刻意做成「重启即解锁」：不要改成持久化，否则共享出口 IP 被攻击者刷满时所有者无法自助恢复。
