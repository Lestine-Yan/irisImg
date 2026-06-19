# `internal/pkg/jwt/jwt.go`

封装 JWT 的签发与解析，业务层只通过 `Manager` 与 JWT 打交道。底层使用 `github.com/golang-jwt/jwt/v5`，固定 HS256 签名。

## 类型

### `Claims`

```go
type Claims struct {
    Username string `json:"username"`
    jwtv5.RegisteredClaims
}
```

- 自定义字段 `Username`：用于在中间件里直接取出当前用户名。
- 嵌入 `RegisteredClaims` 以拥有标准的 `iss / sub / iat / nbf / exp`。

### `Manager`

```go
type Manager struct {
    secret []byte
    issuer string
    expire time.Duration
}
```

不可导出字段；通过 `NewManager` 构造。同一个 `Manager` 全局共享，goroutine 安全（无可变状态，方法只读字段）。

## 函数

### `NewManager(cfg config.JWTConfig) *Manager`

- 读取 `cfg.Secret / cfg.Issuer / cfg.ExpireHours`。
- `expire_hours <= 0` 时回退到 24 小时，避免误配置导致一签发即过期。

### `Issue(username string) (token string, expiresAt int64, err error)`

- 以「现在」为基准计算 `exp = now + expire`。
- 填充 `Issuer / Subject / IssuedAt / NotBefore / ExpiresAt`，外加自定义 `Username`。
- 用 HS256 签名并返回 token 字符串、`exp` 的 Unix 秒、错误。

### `Parse(tokenStr string) (*Claims, error)`

- `jwtv5.ParseWithClaims` + 自定义 keyFunc：**显式断言** `t.Method.(*jwtv5.SigningMethodHMAC)`，否则返回错误。
  - 这一步是防御 `alg=none` 与算法降级攻击。
- 解析失败、签名不对、claims 类型不匹配、`Valid` 为 false 都会返回错误（v5 默认会校验 `exp / nbf`）。
- 调用方拿到错误时**不要**根据 err 文本区分原因，统一作 401 处理（见 [`middleware/auth`](../middleware/auth.md)）。

## 与其它包的关系

- 依赖 [`config.JWTConfig`](../../config/config.md)，但不直接读 `config.Global`，所有配置通过构造函数传入。
- 被 [`service.AuthService`](../service/auth.md) 用来签发；被 [`middleware.JWTAuth`](../middleware/auth.md) 用来解析。

## 修改建议

- 切换签名算法（如换成 RS256）时，只要扩展 `Manager` 字段并改 `Issue/Parse` 内部，外部调用不变。
- 不要在这里加任何业务校验（用户是否存在、是否被踢下线等），那些属于 service 层的职责。
