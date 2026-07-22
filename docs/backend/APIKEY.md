# API 密钥鉴权说明（irisImg 后端）

> irisImg 在后台 JWT 登录之外，提供一套独立的 **API 密钥（apikey）鉴权体系**，用于外部程序「申请图片 / 添加图片」。
> 本文档面向使用/调试本服务的人，跨文件讲清楚「密钥怎么签发」「请求怎么校验」「报错怎么读」。
> 各 `.go` 文件的逐文件文档见各自目录的 `.md`；登录链路见 [`AUTH.md`](./AUTH.md)。

---

## 1. 整体设计

- **明文只给一次**：明文密钥 = 32 字节加密随机数经 `base64(RawURLEncoding, URL-safe 无填充)` → 固定 **43 字符**，仅在创建响应里返回一次；库里只存 **SHA-256 哈希**（十六进制 64 字符），无法反推明文。
- **两种权限**：`readonly`（只读，仅 GET）/ `readwrite`（读写，可 POST）。
- **请求头携带**：调用图片接口时放在 `X-API-Key` 头里（不是 `Authorization`）。
- **独立于 JWT**：密钥管理接口（创建/列表/吊销）需后台 JWT 登录；图片接口用密钥鉴权。两套互不依赖。
- **按密钥限流**：每把密钥一个内存令牌桶，默认 100 次/分钟。
- **HTTPS 强制**：管理接口可按配置要求 HTTPS（生产由 Nginx 反代）。

参与的代码文件：

| 角色 | 文件 |
| --- | --- |
| Schema | `ent/schema/apikey.go`、`ent/schema/image.go`（新增 key edge） |
| 配置 | `config/config.go`、`config/config.yaml`（`apikey` 段） |
| 工具 | `internal/pkg/apikey/apikey.go`、`internal/pkg/ratelimit/ratelimit.go` |
| DTO | `internal/model/apikey.go`、`internal/model/image.go`（KeyID） |
| DAO | `internal/dao/dao.go`（APIKeyDAO）、`internal/dao/entdao/apikey.go`、`.../image.go` |
| 业务逻辑 | `internal/service/apikey.go` |
| 中间件 | `internal/middleware/apikey.go`、`internal/middleware/https.go` |
| 控制器 | `internal/api/apikey.go`、`internal/api/image.go` |
| 统一响应 | `internal/pkg/response/response.go`（新增错误码） |
| 路由装配 | `internal/router/router.go` |

## 2. 配置

```yaml
apikey:
  rate_limit_per_minute: 100   # 单密钥默认限流阈值（次/分钟），0/缺省回退 100
  https_only: false            # 本地 false；生产（Nginx HTTPS 反代）置 true
```

详见 [`config.md`](./config/config.md)。

## 3. 接口一览

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| POST | `/api/v1/apikeys` | JWT + HTTPS | 创建密钥，**响应含一次性明文 `key`** |
| GET | `/api/v1/apikeys` | JWT + HTTPS | 列出全部密钥（不含明文/哈希） |
| PATCH | `/api/v1/apikeys/:id` | JWT + HTTPS | 重命名密钥 |
| POST | `/api/v1/apikeys/:id/reset` | JWT + HTTPS | 重置明文，**响应含一次性新明文 `key`**；同时取消吊销 |
| POST | `/api/v1/apikeys/:id/revoke` | JWT + HTTPS + 密码 | 吊销密钥（软删除，仍展示但无法鉴权） |
| DELETE | `/api/v1/apikeys/:id` | JWT + HTTPS + 密码 | 物理删除密钥并级联删除关联图片 |
| GET | `/api/v1/images` | API Key（任意有效） | 申请图片（**占位，返回 501**） |
| POST | `/api/v1/images` | API Key（需 readwrite） | 添加图片（**已实现**，链路见 [`IMAGE.md`](./IMAGE.md)） |

## 4. 签发链路（POST /apikeys）

```
client          api.APIKeyAPI        service.APIKeyService        pkg/apikey            dao.APIKeyDAO
  │ POST /apikeys   │                       │                         │                      │
  │ {name,scope,..} │                       │                         │                      │
  │ (JWT+HTTPS)     │                       │                         │                      │
  │ ───────────────►│ ShouldBindJSON         │                         │                      │
  │                 │ ── Create(req) ───────►│ 校验 scope               │                      │
  │                 │                       │ ── Generate() ─────────►│ 32字节随机→base64       │
  │                 │                       │                         │ +SHA256+prefix        │
  │                 │                       │ ◄── plaintext,hash,prefix│                      │
  │                 │                       │ ── Create(只存 hash) ───────────────────────────►│
  │                 │                       │ ◄────────────────────────────────────────────────│
  │ ◄── 200 {key:明文} ─ CreateAPIKeyResponse │                         │                      │
```

明文密钥 `key` **仅此一次返回**，服务端不再可查；丢失只能吊销重建。

## 5. 校验链路（图片接口，X-API-Key）

```
client       middleware.APIKeyAuth      service.Authenticate      ratelimit.Store     api.ImageAPI
  │ GET/POST /images    │                       │                       │                  │
  │ X-API-Key: <明文>   │                       │                       │                  │
  │ ───────────────────►│ ①header 空? →401 40110 │                       │                  │
  │                     │ ── Authenticate ─────►│ ②格式校验 IsValidFormat │                  │
  │                     │                       │ ③Hash→GetByHash 查库    │                  │
  │                     │                       │   不存在/吊销           │                  │
  │                     │ ◄── key / err ────────│                       │                  │
  │                     │   err→401 40120        │                       │                  │
  │                     │ ④非GET且非readwrite?→403│                       │                  │
  │                     │ ── Allow(id,limit) ───────────────────────────►│ 令牌桶            │
  │                     │   超限→429 42900        │                       │                  │
  │                     │ c.Set("api_key_id",id) │                       │                  │
  │                     │ Touch(尽力)            │                       │                  │
  │                     │ ── Next ──────────────────────────────────────────────────────────►│
```

## 6. 权限矩阵（scope × 方法）

| scope \ 方法 | GET（申请图片） | 非 GET（POST 添加图片） |
| --- | --- | --- |
| `readonly` | ✅ 允许 | ❌ 403 `CodeForbidden` |
| `readwrite` | ✅ 允许 | ✅ 允许 |

> 校验顺序：先验密钥有效性（步骤①②③），再验权限（④），最后限流（⑤）。无效密钥永远不会走到权限/限流判定。

## 7. 错误码表

| code | HTTP | 常量 | 含义 |
| --- | --- | --- | --- |
| 0 | 200 | `CodeOK` | 成功 |
| 40000 | 400 | `CodeBadRequest` | 入参非法（如 scope 不对、ID 非数字） |
| 40100 | 401 | `CodeUnauthorized` | JWT 未登录 / 无效（管理接口） |
| 40110 | 401 | `CodeAPIKeyMissing` | 缺少 `X-API-Key` 请求头 |
| 40120 | 401 | `CodeAPIKeyInvalid` | 密钥格式非法 / 不存在 / 已吊销 |
| 40300 | 403 | `CodeForbidden` | 只读密钥访问写接口 / 未走 HTTPS / 吊销·删除密码二次确认失败 |
| 40400 | 404 | `CodeNotFound` | 吊销 / 重置 / 删除 / 重命名时密钥不存在 |
| 42900 | 429 | `CodeTooManyRequests` | 触发限流 |
| 50000 | 500 | `CodeServerError` | 内部错误 / 占位接口 501 |

## 8. HTTPS 强制

- 部署形态：Nginx 统一 HTTPS 反代，后端本地监听 HTTP。
- [`HTTPSOnly(enabled, trustedProxies)`](./internal/middleware/https.md) 校验 `c.Request.TLS != nil || (来自可信反代 && X-Forwarded-Proto == "https")`：仅当 TCP 对端属于 `server.trusted_proxies` 网段时才认 `X-Forwarded-Proto`，否则只认 `c.Request.TLS`。
- 配置 `apikey.https_only`：本地 `false`、生产 `true`。仅挂在 `/apikeys` 管理组上（在 JWT 之后）。

## 9. 与 JWT 鉴权的区别

| 维度 | JWT（[`AUTH.md`](./AUTH.md)） | API Key（本文档） |
| --- | --- | --- |
| 用途 | 后台管理（本人登录） | 外部程序申请/添加图片 |
| 凭据载体 | `Authorization: Bearer <token>` | `X-API-Key: <明文>` |
| 是否落库 | 否（无状态） | 是（存哈希，可吊销） |
| 权限分级 | 无（单管理员） | readonly / readwrite |
| 限流 | 无 | 按密钥令牌桶 |
| 失效方式 | 改 secret / 等过期 | 吊销（`revoked=true`） |

## 10. 限流

- 实现见 [`ratelimit.md`](./internal/pkg/ratelimit.md)：每把密钥一个 `golang.org/x/time/rate` 令牌桶，速率「每分钟 N 次」、突发容量 N。
- N 取密钥自身 `rate_limit`（>0 时），否则取全局 `apikey.rate_limit_per_minute`（默认 100）。
- **内存实现**：单实例有效；进程重启清零；多实例需换 Redis。

## 11. image.key 关联

- `image` 表新增可空外键 `key_id`（`*int`），通过 Ent `key` edge（`edge.From("key", ApiKey.Type).Ref("images")`）绑定，记录图片由哪把密钥添加。
- 通过后台 JWT 上传的图片 `key_id` 为空；通过密钥 POST 添加的图片由 [`api.ImageAPI.Create`](./internal/api/image.md) 回填中间件注入的 `api_key_id`。
- 详见 [`ent/schema/image.md`](./ent/schema/image.md)、[`entdao/image.md`](./internal/dao/entdao/image.md)、[`model/image.md`](./internal/model/image.md)。

## 11.1 密钥管理操作（重命名 / 重置 / 吊销 / 删除）

后台 JWT 登录后可对密钥做以下管理操作，均挂在 `/apikeys` 下（JWT + HTTPS）：

| 操作 | 接口 | 行为 |
| --- | --- | --- |
| 重命名 | `PATCH /apikeys/:id` | 改 `name`，不影响明文与状态 |
| 重置明文 | `POST /apikeys/:id/reset` | 生成新明文/哈希/前缀，旧明文立即失效；**同时取消吊销**（重新激活）。响应含一次性新明文 |
| 吊销 | `POST /apikeys/:id/revoke` | 软删除：`revoked=true`，密钥仍展示、仍可操作，仅无法通过鉴权 |
| 删除 | `DELETE /apikeys/:id` | 物理删除密钥，并**级联删除**由该密钥上传的全部图片（物理文件 + 记录） |

**密码二次确认**：吊销与删除为敏感操作，请求体需携带 `{username, password}`，后端用 `subtle.ConstantTimeCompare` 校验（见 [`service/auth.go`](./internal/service/auth.md) 的 `VerifyCredentials`）。校验失败返回 **403**（而非 401）——前端 `useApi` 会把 401 当作 JWT 失效并全局登出，403 则交由调用方就地提示「用户名或密码错误」。

**已吊销密钥**：不影响展示与操作——仍出现在列表里（带「已吊销」标记），仍可重命名 / 重置（重置会重新激活）/ 删除；只是 `Authenticate` 对其返回 `ErrKeyRevoked`，无法通过 `X-API-Key` 鉴权。

**删除的级联清理**（见 [`service.APIKeyService.Delete`](./internal/service/apikey.md)）：取关联图片 → best-effort 删物理文件 → 删图片记录（解除外键约束）→ 删密钥。非事务，失败可重试。

## 12. 示例

```bash
# 1) 后台登录拿 JWT（见 AUTH.md）
TOKEN="eyJ..."

# 2) 创建一把读写密钥（明文仅此一次返回）
curl -X POST http://localhost:8080/api/v1/apikeys \
     -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
     -d '{"name":"my-uploader","scope":"readwrite"}'
# → {"code":0,...,"data":{"id":1,"prefix":"AbCdEf01","key":"<43字符明文>",...}}

# 3) 用密钥访问图片接口
KEY="<43字符明文>"
curl http://localhost:8080/api/v1/images -H "X-API-Key: $KEY"
# → GET 当前 501（占位）；缺头→401 40110；只读密钥 POST→403 40300；超限→429 42900；
#   POST 上传请见 IMAGE.md。

# 4) 吊销
curl -X DELETE http://localhost:8080/api/v1/apikeys/1 -H "Authorization: Bearer $TOKEN"
```

## 13. 安全注意事项

- 明文密钥等价口令，泄露需立即吊销重建；库里只存哈希，泄露库文件不会直接暴露明文。
- 生产务必 `https_only: true`。后端仅对 `server.trusted_proxies` 网段内的请求信任 `X-Forwarded-Proto`（默认本地回环，跨机反代需追加反代 CIDR）；不可信 peer 只认 `c.Request.TLS`，伪造该头无效。仍建议 Nginx 用 `proxy_set_header X-Forwarded-Proto $scheme` 覆盖该头作纵深防御。
- 内存限流不抗多实例与重启；对抗滥用建议叠加反代层（Nginx/网关）限流。
