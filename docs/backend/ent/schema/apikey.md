# ent/schema/apikey.go

定义 API 密钥实体 **ApiKey** 的 Ent schema。密钥用于外部程序「申请图片 / 添加图片」的鉴权，独立于后台 JWT 登录。

明文密钥由 32 字节加密随机数经 base64(URL-safe, 无填充) 生成，**仅在创建时返回一次**；数据库只保存其 SHA-256 哈希，无法反推明文。密钥区分 `readonly`（只读，仅可访问 GET 接口）与 `readwrite`（读写，可 POST 添加图片）两种权限。

`go generate ./ent` 会基于本 schema 生成 `backend/ent/` 下的类型安全客户端代码（`apikey*.go` 等，均为生成产物，不手工编辑）。

## 字段

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `name` | string | 非空 | 密钥标签，便于在管理界面识别用途 |
| `key_hash` | string | 非空、唯一 | 明文密钥的 SHA-256 哈希（十六进制），校验时按此列查找 |
| `prefix` | string | 非空 | 明文密钥前缀（前 8 位），仅用于展示识别，不足以反推明文 |
| `scope` | enum | `readonly` / `readwrite` | 权限范围 |
| `rate_limit` | int | ≥0，默认 0 | 限流阈值（次/分钟），0 表示沿用全局默认配置 |
| `revoked` | bool | 默认 false | 是否已吊销；吊销后密钥立即失效 |
| `last_used_at` | time | 可空（Optional + Nillable） | 最近一次使用时间，未使用过为空 |
| `created_at` | time | 不可变，默认 `time.Now` | 创建时间 |

## 索引

- `key_hash`（唯一）：校验时按哈希精确查找。

## 关联关系（Edges）

- `images`：`edge.To("images", Image.Type)` —— 一个密钥可以添加多张图片。反向由 [`Image`](image.md) 的 `key` edge（`edge.From("key", ApiKey.Type).Ref("images")`）绑定到显式字段 `key_id`。

## 调用关系

schema 仅供 Ent 代码生成使用；运行时由 [`internal/dao/entdao/apikey.go`](../../internal/dao/entdao/apikey.md) 通过生成的 `ent.Client` 读写。错误码、权限矩阵、鉴权链路等特性级说明见 [`APIKEY.md`](../../APIKEY.md)。
