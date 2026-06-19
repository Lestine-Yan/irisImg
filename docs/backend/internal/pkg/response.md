# `internal/pkg/response/response.go`

统一响应体定义与若干快捷方法，所有 HTTP 接口都通过它写出 JSON，避免每个 handler 自己拼 `c.JSON(...)`。

## 响应结构

```go
type Body struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

- `code`：业务状态码，与 HTTP 状态码**分离**。前端凭 `code == 0` 判断成功。
- `message`：可读提示，可直接展示给用户。
- `data`：成功时承载业务数据；失败时省略（`omitempty`）。

## 业务状态码

| 常量 | 值 | 含义 |
| --- | --- | --- |
| `CodeOK` | `0` | 成功 |
| `CodeBadRequest` | `40000` | 入参非法 |
| `CodeUnauthorized` | `40100` | 未登录 / token 无效 / 凭据错 |
| `CodeNotFound` | `40400` | 资源不存在 |
| `CodeServerError` | `50000` | 服务器内部错误 |

## 快捷方法

| 方法 | HTTP | code |
| --- | --- | --- |
| `Success(c, data)` | 200 | 0 |
| `Fail(c, httpStatus, code, msg)` | 自定义 | 自定义 |
| `BadRequest(c, msg)` | 400 | 40000 |
| `Unauthorized(c, msg)` | 401 | 40100 |
| `NotFound(c, msg)` | 404 | 40400 |
| `ServerError(c, msg)` | 500 | 50000 |

`Fail` 是兜底，所有具名方法都是它的薄封装；新增业务码时优先扩 `CodeXxx` + 对应快捷方法，不要让 handler 直接调 `Fail`。

## 修改建议

- 新增字段（例如 `traceId`）时改 `Body` 结构 + `Success/Fail` 内部填充，调用方无需变。
- `Data` 用 `interface{}` 是为了灵活；如果某个接口返回的结构稳定，**仍建议**在 `model` 里建一个具名 struct 而不是 `gin.H`，方便文档化和前后端约定。
