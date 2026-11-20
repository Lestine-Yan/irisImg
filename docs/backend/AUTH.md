# 登录逻辑说明（irisImg 后端）

> 私人图床的认证模型：**只服务一个用户**，账号信息直接来自 `config/config.yaml`，登录通过 JWT 维持会话。
> 本文档面向使用/调试本服务的人，逐层解释「登录是怎么走通的」「token 怎么校验」「报错怎么读」。
> 各 `.go` 文件的逐文件文档见同目录的 [`README.md`](./README.md)。

---

## 1. 整体设计

- **没有用户表**：用户名/密码写在 `config.yaml` 的 `auth` 段，明文。部署时务必修改。
- **没有会话存储**：状态全部装在 JWT 里，服务端不保存。
- **HS256 签名**：单服务部署，对称密钥够用；密钥写在 `auth.jwt.secret`。
- **统一 401**：登录失败、token 缺失、token 错误、token 过期，对外都是 HTTP 401 + `code = 40100`，不暴露具体原因。
- **登录失败锁定**：同一客户端 IP 连续失败达到 `auth.login_max_failures`（默认 3）后，锁定该 IP `auth.login_lockout_hours`（默认 168 小时）。锁定期间即使密码正确也返回 **429 + `Retry-After`**。锁定按 IP 隔离，只存进程内存，**重启后端即解锁**——这是所有者忘记密码时的兜底恢复路径。

参与的代码文件：

| 角色 | 文件 |
| --- | --- |
| 配置 | `config/config.yaml`、`config/config.go` |
| 路由 + 装配 | `internal/router/router.go` |
| 控制器 | `internal/api/auth.go` |
| 业务逻辑 | `internal/service/auth.go` |
| JWT 工具 | `internal/pkg/jwt/jwt.go` |
| 中间件 | `internal/middleware/auth.go` |
| 入参/出参 DTO | `internal/model/auth.go` |
| 统一响应 | `internal/pkg/response/response.go` |

## 2. 配置

```yaml
auth:
  username: "admin"
  password: "admin123"            # 明文；部署务必修改
  jwt:
    secret: "please-change-me-to-a-long-random-string"
    issuer: "irisImg"
    expire_hours: 24
  login_max_failures: 3       # 连续失败多少次后锁定；0/缺省=3
  login_lockout_hours: 168    # 锁定时长（小时）；0/缺省=168（7 天）
```

- 锁定按**客户端 IP** 隔离：一个 IP 被锁不影响其他 IP；因此换 IP 可以绕过计数（Nginx 侧另有按 IP 的 5r/m 限流）。
- 锁定状态只存内存（有容量上限，最多 4096 个 IP 的状态），进程重启（或容器重启）即清零，等于内置的解锁开关。

- `expire_hours` 配置 `<= 0` 时，`jwt.NewManager` 自动回退到 24 小时。
- 改了 `secret` 之后已经签发的 token 会立刻全部失效（签名校验不通过），属于预期行为，可以拿来「踢掉所有会话」。

## 3. 接口

| 方法 | 路径 | 鉴权 | 入参 | 出参（`data` 字段） |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/auth/login` | 否 | `{username, password}` | `{token, token_type:"Bearer", expires_at}` |
| GET  | `/api/v1/auth/me`    | 是（`Authorization: Bearer <token>`） | — | `{username}` |

`/api/v1/auth/me` 既能取当前用户信息，也充当**校验 token 是否有效**的接口：返回 200 即合法，401 即无效。

### 统一响应体

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

- `code = 0` 成功；其余情况见下表。
- 失败时 `data` 字段会被 `omitempty` 省掉。

| code | HTTP | 含义 |
| --- | --- | --- |
| 0 | 200 | 成功 |
| 40000 | 400 | 入参非法（缺字段、JSON 不合法） |
| 40100 | 401 | 凭据错 / token 缺失 / token 无效 / token 过期 |
| 42900 | 429 | 登录失败次数达到阈值后的锁定（响应头 `Retry-After` 给出剩余秒数） |
| 50000 | 500 | 服务器内部错误（例如 jwt 签发失败） |

## 4. 登录流程（POST /auth/login）

```
client                router          api.AuthAPI         service.AuthService        pkg/jwt.Manager
  │  POST /auth/login   │                  │                       │                       │
  │  {username,password}│                  │                       │                       │
  │ ───────────────────►│                  │                       │                       │
  │                     │ ── handle ─────► │                       │                       │
  │                     │                  │ ShouldBindJSON         │                       │
  │                     │                  │ (失败 → 400 BadRequest)│                       │
  │                     │                  │ ── Login(req) ───────► │                       │
  │                     │                  │                       │ ConstantTimeCompare    │
  │                     │                  │                       │ username & password    │
  │                     │                  │                       │ ── Issue(username) ──► │
  │                     │                  │                       │       签 HS256 token   │
  │                     │                  │                       │ ◄────────────────────  │
  │                     │                  │ ◄── LoginResponse ───  │                       │
  │ ◄── 200 token ──────│                  │                       │                       │
```

关键点：

1. **入参校验**：`LoginRequest` 用 `binding:"required"` 保证 username/password 都非空，缺字段直接 400。
2. **常量时间比较**：`crypto/subtle.ConstantTimeCompare` 防止根据响应耗时反推用户名是否存在或密码前缀是否对。两项**都比对完才返回**，进一步消除分支耗时差。
3. **错误统一**：用户名或密码错都返回 `ErrInvalidCredentials`，控制器映射成 `401 / 40100 / 用户名或密码错误`。
4. **签发**：`jwt.Manager.Issue` 用 HS256 签出 token，载荷包含：
   - 自定义：`username`
   - 标准：`iss`（来自配置）、`sub`（同 username）、`iat`、`nbf`（均为「现在」）、`exp`（now + expire）
5. **响应**：返回 token 字符串、`token_type: "Bearer"`、`expires_at`（Unix 秒）。

锁定相关补充：

- **触发时机**：同一 IP 的第 `login_max_failures` 次密码错误时写入该 IP 的 `lockedUntil = now + lockout`，本次仍返回 401（前端提示密码错误），**下一次请求起**返回 429。
- **锁定期间**：`Login` 在比对密码前直接返回 `ErrLoginLocked`，不查库、不增加计数（避免攻击者无限延长锁定）。
- **解锁**：到期后下一次请求惰性清除锁定状态；成功登录也会清零失败计数。
- **日志**：仅在触发锁定的一次记 `auth.login_locked`（warn），锁定期间被拒的请求只进访问日志（429），避免日志洪水。
- **恢复**：忘记密码时 SSH 到服务器重启进程即可解除锁定，再改配置里的密码。

### 示例

```bash
# 成功
curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin123"}'
# → {"code":0,"message":"success","data":{"token":"eyJ...","token_type":"Bearer","expires_at":1781961787}}

# 失败：密码错
curl -i -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"wrong"}'
# → HTTP/1.1 401   {"code":40100,"message":"用户名或密码错误"}
```

## 5. 受保护接口的访问（GET /auth/me）

```
client            middleware.JWTAuth          pkg/jwt.Manager       api.AuthAPI
  │ GET /auth/me        │                          │                     │
  │ Authorization:      │                          │                     │
  │   Bearer <token>    │                          │                     │
  │ ───────────────────►│                          │                     │
  │                     │ 读 Authorization 头      │                     │
  │                     │ 校验前缀 "Bearer "       │                     │
  │                     │ Trim 出 token 字符串     │                     │
  │                     │ ── Parse(tokenStr) ────► │                     │
  │                     │                          │ 强制 HMAC 校签      │
  │                     │                          │ 校验 exp/nbf        │
  │                     │ ◄── *Claims ───────────  │                     │
  │                     │ c.Set("username", ...)   │                     │
  │                     │ c.Next()                 │                     │
  │                     │ ── Me(c) ────────────────────────────────────► │
  │                     │                          │                     │ c.Get("username")
  │ ◄── 200 {username} ──────────────────────────────────────────────── │
```

中间件的失败分支（任何一步出错都直接 `c.Abort()` + 401）：

| 触发条件 | message |
| --- | --- |
| 没带 `Authorization` 头 | `缺少 Authorization 请求头` |
| 头部不是 `Bearer ` 开头 | `Authorization 格式应为 Bearer <token>` |
| `Bearer ` 后面是空 | `token 不能为空` |
| `Manager.Parse` 任意 err（签名错/格式错/过期/算法不对） | `token 无效或已过期` |

### 示例

```bash
TOKEN="eyJ..."  # 上一步登录拿到的 token

# 合法
curl http://localhost:8080/api/v1/auth/me -H "Authorization: Bearer $TOKEN"
# → {"code":0,"message":"success","data":{"username":"admin"}}

# 缺头
curl -i http://localhost:8080/api/v1/auth/me
# → 401  {"code":40100,"message":"缺少 Authorization 请求头"}

# 伪造 / 改过密钥后
curl -i http://localhost:8080/api/v1/auth/me -H "Authorization: Bearer not.a.real.token"
# → 401  {"code":40100,"message":"token 无效或已过期"}
```

## 6. JWT 载荷长什么样

签发后的 token 是三段 Base64URL（`header.payload.signature`）。把 payload 段解码后大致是：

```json
{
  "username": "admin",
  "iss": "irisImg",
  "sub": "admin",
  "iat": 1781875387,
  "nbf": 1781875387,
  "exp": 1781961787
}
```

- 调试时可以贴到 https://jwt.io 看（仅用于已废弃/测试用 token；线上 token 不要贴外网）。
- 服务端校验时会**显式断言** `*jwtv5.SigningMethodHMAC`，防止 `alg=none` 或被切换到 RS256 公钥伪造的攻击。

## 7. 前端如何对接

1. 调 `POST /api/v1/auth/login` 拿到 `data.token`。
2. 把 token 存在内存或 `localStorage / cookie`（取决于安全要求）。
3. 后续所有受保护请求，都加请求头：

   ```
   Authorization: Bearer <token>
   ```

4. 收到 401（`code = 40100`）时清 token，跳回登录页。
5. 想做无感续期：在 token 接近 `expires_at` 时让用户重新输入密码登录，或者在 service 里加 refresh token 接口（当前未实现）。

## 8. 常见排错

| 现象 | 可能原因 | 检查 |
| --- | --- | --- |
| 永远 `用户名或密码错误`，但密码没错 | 改了 `config.yaml` 但没重启 | 重启 `go run ./cmd/server` |
| 重启后所有 token 失效 | 改过 `auth.jwt.secret` | 这是预期行为；让用户重新登录 |
| 浏览器调登录跨域失败 | `middleware/cors.go` 默认 `*` 但浏览器对带 `Authorization` 的请求严格 | 上线前把 CORS 收紧到具体域名 |
| token 一签发立刻过期 | `expire_hours` 写成 0/负数 | jwt 包会回退到 24 小时；但建议显式配 ≥1 |
| 用 `Bearer xxx` 还是 401 | token 字符串里包含换行/空格 | 用 `curl --data-binary` 传时注意，或前端先 `trim` |

## 9. 安全注意事项

- `auth.password` 是明文 → **不要把 `config.yaml` 提交到仓库**。当前 `config.yaml` 在仓库里只是默认模板，部署时通过 `IRIS_CONFIG=/path/to/private.yaml` 指向服务器本地版本。
- `auth.jwt.secret` 至少 32 字节随机字符串。
- 生产部署务必走 HTTPS；HS256 + 明文 token 在 HTTP 下等于裸奔。
- **可信代理**：后端启动时用 `server.trusted_proxies` 调 `gin.Engine.SetTrustedProxies`，`c.ClientIP()` 只采信来自可信反代的 `X-Forwarded-For`；Nginx 模板改为只传 `$remote_addr`，客户端自带的伪造链会被丢弃。
- **按 IP 锁定的代价**：攻击者只能锁住自己所用的 IP，无法远程锁死所有者；但若与所有者共享出口 IP（NAT / 公司代理），会被一起锁住。锁定只存内存，SSH 重启即恢复。
- 想升级到 bcrypt 哈希：在 `AuthConfig` 加 `password_hash`，把 `service.AuthService.Login` 里的 `subtle.ConstantTimeCompare` 改成 `bcrypt.CompareHashAndPassword(hash, []byte(req.Password))` 即可，外部 API 和前端无感。
