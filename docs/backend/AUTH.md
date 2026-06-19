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
```

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
- 想升级到 bcrypt 哈希：在 `AuthConfig` 加 `password_hash`，把 `service.AuthService.Login` 里的 `subtle.ConstantTimeCompare` 改成 `bcrypt.CompareHashAndPassword(hash, []byte(req.Password))` 即可，外部 API 和前端无感。
