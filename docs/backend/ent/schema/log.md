# ent/schema/log.go

定义日志中心的统一日志实体 **AccessLog**（表名经 `entsql` 注解固定为 `logs`）。该实体既记录每条 HTTP 请求（`event=http.request`），也记录关键业务事件（`image.upload` / `apikey.revoke` / `auth.login_success` / `log.clear` 等），还记录 panic（`event=panic`），供日志中心查询 / 直方图 / 清理。

ent 在代码生成阶段会预声明 `Log` 标识符，故 Go 侧 schema 类型用 `AccessLog` 规避冲突；实际表名则经 `Annotations()` 中的 `entsql.Annotation{Table: "logs"}` 固定为 `logs`，避免默认的 `access_logs`。访问日志由中间件异步批量落库。

`go generate ./ent` 会基于本 schema 生成 `backend/ent/` 下的类型安全客户端代码（`accesslog*.go` 等，均为生成产物，不手工编辑）。

## 字段

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `timestamp` | time | 不可变，默认 `time.Now` | 日志发生时间，用于排序 / 直方图 / 时间范围过滤 |
| `level` | enum | `debug` / `info` / `warn` / `error`，默认 `info` | 日志级别，按 HTTP 状态推导（≥500 error / ≥400 warn / 否则 info）或业务事件显式指定 |
| `event` | string | 非空（`NotEmpty`） | 事件类型：`http.request` / `image.upload` / `apikey.*` / `auth.*` / `log.clear` / `panic` |
| `method` | string | 可空（Optional） | HTTP 方法，仅访问日志有值 |
| `path` | string | 可空（Optional） | 请求路径，用于关键字模糊匹配 |
| `status` | int | 可空（Optional + Nillable） | HTTP 状态码，非请求类事件为空（NULL） |
| `duration_ms` | int | 可空（Optional + Nillable） | 请求耗时（毫秒），仅访问日志有值 |
| `client_ip` | string | 可空（Optional） | 客户端 IP |
| `request_id` | string | 可空（Optional） | 请求追踪 ID，关联同一请求的访问日志与业务事件 |
| `api_key_id` | int | 可空（Optional + Nillable） | 来源 API 密钥 ID，无关联时为空；不建 Edge |
| `username` | string | 可空（Optional） | 操作者用户名（JWT 登录用户） |
| `message` | string | 默认 `""` | 可读描述 / 错误信息 |
| `created_at` | time | 默认 `time.Now` | 记录落库时间 |

## 索引

- `timestamp`：按时间排序 / 时间范围过滤。
- `level`：按级别筛选。
- `event`：按事件类型筛选。
- `request_id`：按追踪 ID 串联同一请求的访问日志与业务事件。
- `api_key_id`：按来源密钥筛选。

## 关联关系（Edges）

- 无。日志为高频写入实体，为避免外键约束带来的插入开销，`api_key_id` 仅以普通字段记录来源密钥 ID，不建立 `edge.To` / `edge.From`。

## 调用关系

schema 仅供 Ent 代码生成使用；运行时由 [`internal/dao/entdao/log.go`](../../internal/dao/entdao/log.md) 通过生成的 `ent.Client` 读写。日志链路的异步落库、查询过滤、直方图聚合、清理策略等特性级说明见 [`LOG.md`](../../LOG.md)。
