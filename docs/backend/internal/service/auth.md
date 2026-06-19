# `internal/service/auth.go`

认证业务逻辑层。控制器把入参原样交过来，本层负责「凭据是否正确」「该签什么 token」这两件事。

## 类型与变量

### `AuthService`

```go
type AuthService struct {
    cfg    config.AuthConfig
    jwtMgr *jwt.Manager
}
```

- 持有配置（用户名/密码）与 JWT 管理器，由 [`router`](../router/router.md) 注入。
- 没有任何可变状态，goroutine 安全。

### `ErrInvalidCredentials`

```go
var ErrInvalidCredentials = errors.New("invalid username or password")
```

- 唯一对外暴露的「业务错误」哨兵值，控制器用 `errors.Is` 判别后映射到 401。
- 不区分「用户名错」还是「密码错」，避免暴力猜账号。

## 函数

### `NewAuthService(cfg config.AuthConfig, m *jwt.Manager) *AuthService`

普通构造器，没有副作用。

### `Login(req *model.LoginRequest) (*model.LoginResponse, error)`

```go
usernameOK := subtle.ConstantTimeCompare(...) == 1
passwordOK := subtle.ConstantTimeCompare(...) == 1
if !usernameOK || !passwordOK {
    return nil, ErrInvalidCredentials
}
token, expiresAt, err := s.jwtMgr.Issue(s.cfg.Username)
```

关键点：

- **`crypto/subtle.ConstantTimeCompare`**：常量时间比较，避免依据耗时区分用户名是否存在或密码前缀是否正确。
- **同时比对**用户名和密码：即使用户名错也走完比对，进一步消除时序差异。
- 用配置里的 `cfg.Username`（而非请求里的 `req.Username`）签发 token——多此一举但更稳：避免在配置使用大小写不同写法时出现「输入与签发不一致」的怪问题；当前两者必然相等。
- jwt 签发失败直接把 err 透传给上层，由控制器映射成 500。

## 与其它包的关系

```
api.AuthAPI ──► service.AuthService ──► pkg/jwt.Manager
                       │
                       └─► config.AuthConfig (来自 router 注入)
```

## 修改建议

- 切换到 bcrypt 哈希校验只动两处：`AuthConfig` 加 `password_hash` 字段、`Login` 改用 `bcrypt.CompareHashAndPassword`；外部 API 与中间件都不用改。
- 加「登录失败次数限制 / 锁定」属于 service 层：在 `AuthService` 里维护一个内存计数器或接 redis 即可，不要污染中间件。
