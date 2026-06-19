# `internal/middleware/logger.go`

请求级访问日志，统一用标准库 `log`。

## 函数

### `Logger() gin.HandlerFunc`

- 进入处理前记录 `start = time.Now()` 与 `path`。
- `c.Next()` 让后续处理器跑完。
- 之后输出格式为：`[METHOD] PATH STATUS DURATION`，例如 `[POST] /api/v1/auth/login 200 1.2ms`。

## 修改建议

- 项目体量增大时建议替换为 `zap` / `logrus`，并在中间件里把 traceId、请求体大小、客户端 IP 等加进去。
- 不要在这里输出敏感数据（如登录请求体），日志会泄漏密码。
