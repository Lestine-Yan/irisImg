# `internal/pkg/apikey/apikey.go`

> 文档路径约定：`internal/pkg/apikey/apikey.go` → `docs/backend/internal/pkg/apikey.md`（去掉内层同名目录）。

提供 API 密钥的**生成、哈希与格式校验**工具。不依赖任何业务包，纯函数无状态。

明文密钥由 32 字节加密随机数经 `base64(RawURLEncoding)` 编码而成，长度固定 **43 字符**；数据库只保存其 SHA-256 哈希（十六进制 64 字符），明文仅在创建时返回一次，无法从库中反推。

## 常量

| 常量 | 值 | 说明 |
|------|----|------|
| `keyBytes`（非导出） | `32` | 随机熵字节数 |
| `KeyLength` | `43` | base64(RawURLEncoding) 编码 32 字节后的明文长度（ceil(32×4/3)=43） |
| `prefixLength`（非导出） | `8` | 用于展示识别的明文前缀长度 |

## 函数

### `Generate() (plaintext, hash, prefix string, err error)`

```go
buf := make([]byte, keyBytes)
rand.Read(buf)                                    // crypto/rand
plaintext = base64.RawURLEncoding.EncodeToString(buf)
return plaintext, Hash(plaintext), plaintext[:prefixLength], nil
```

- 生成一把新密钥，返回明文、明文的 SHA-256 哈希、用于展示的前缀（前 8 位）。
- 明文**只应在创建响应中返回一次**，调用方不应持久化明文。
- 随机源失败时返回包装后的错误。

### `Hash(plaintext string) string`

返回明文密钥的 SHA-256 哈希（十六进制小写，64 字符）。校验时拿请求里的明文重新 `Hash` 后按库里的 `key_hash` 列查找。

### `IsValidFormat(s string) bool`

格式校验：长度等于 `KeyLength`（43）且为合法的 base64url 字符集（用 `base64.RawURLEncoding.DecodeString` 验证）。中间件在查库前先做此校验，挡掉明显非法的输入。

## 与其它包的关系

```
service.APIKeyService ──► pkg/apikey.Generate / Hash / IsValidFormat
```

- `Generate` 被 [`service.APIKeyService.Create`](../service/apikey.md) 用于签发。
- `Hash` + `IsValidFormat` 被 `Authenticate` 用于鉴权链路。

## 注意

- 切换密钥长度 / 编码方式只动本包；但需同步评估 `IsValidFormat` 与 `ent/schema` 的展示前缀逻辑。
- 用 `crypto/rand`（非 `math/rand`）保证密钥不可预测。
