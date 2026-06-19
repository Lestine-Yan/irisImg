# `internal/api/ping.go`

最简单的健康检查接口，对应路由 `GET /api/v1/ping`。

## 函数

### `Ping(c *gin.Context)`

- 直接返回成功响应，固定带 `{"pong": true}`。
- 当 `config.Global` 不为 nil 时附带 `app` 与 `version` 字段（来自 `app.name / app.version`）。
- 用 [`response.Success`](../internal/pkg/response.md) 走统一响应体。

## 用途

- 部署后用作存活探针：`curl http://host:port/api/v1/ping`
- 客户端确认后端版本是否匹配预期

## 修改建议

- 这个接口**必须保持公开**且无副作用，不要在这里挂任何鉴权或耗时操作。
- 如果想暴露 build commit、启动时间等信息，建议加到 `app` 配置或单独建 `/api/v1/version` 路由，不要污染 `/ping`。
