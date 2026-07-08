# `internal/middleware/recovery.go`

Panic 恢复中间件，替换 `gin.Recovery()`。捕获后续中间件 / handler 抛出的 panic，用 zap 记录堆栈并落库一条 panic 日志，再统一返回 500，使崩溃也进入日志中心可查。

## 函数

### `Recovery(l *logger.Logger, svc *service.LogService) gin.HandlerFunc`

闭包持有 zap 日志器 `*logger.Logger` 与日志服务 `*service.LogService`，由 [`router`](../router/router.md) 注入。用 `defer recover()` 包住 `c.Next()`，处理流程：

1. `c.Next()` 让后续中间件（CORS / Logger 等）与 handler 正常执行。
2. 若 `recover()` 捕获到 panic：
   1. `debug.Stack()` 取堆栈；
   2. `l != nil` 时调 `l.Error` 记录，字段：`panic`（`zap.Any`）、`stack`、`method`、`path`；request_id 由 logger 从 `c.Request.Context()` 自动附加；
   3. `svc != nil` 时经 `LogContextFromGin(c)`（定义于 `requestid.go`）取 request id / 用户名 / 密钥 ID / 客户端 IP，调 `svc.Record` 落库一条 `model.Log`：`Level=LevelError`、`Event=EventPanic`、`Method/Path` 取自请求、`Message="panic: %v"`；
   4. [`response.ServerError`](../pkg/response.md)(c, "服务器内部错误") 写出 HTTP 500（业务码 `CodeServerError`=50000），再 `c.Abort()` 终止链路。

落库的 panic 日志**不显式设置 `Status` 字段**（`model.Log.Status` 为 nil）。因为 panic 会沿 `c.Next()` 调用栈向上展开，跳过外层 [`Logger`](logger.md) 中 `c.Next()` 之后的访问日志代码，故该请求不会再产生 `http.request` 条目；实际写入客户端的 500 由 `response.ServerError` 完成，并未落库到任何 `Status` 字段——即此条 panic 日志的状态隐含为 500。

## 与其它包的关系

```
RequestID ──► Recovery ──► CORS ──► Logger ──► handler (panic!)
                 │
                 │ defer recover()
                 ├─► logger.Error(panic, stack, method, path)
                 ├─► LogContextFromGin ──► LogService.Record(event=panic, level=error)
                 └─► response.ServerError(500) + Abort
```

中间件链顺序见 [`router`](../router/router.md)：`RequestID -> Recovery -> CORS -> Logger`。Recovery 位于 Logger 之外，因此 panic 不会逃逸到 gin 默认的 500 处理，而是被这里捕获并落库；panic 日志与同一请求的其它条目通过 request id 关联。

## 修改建议

- `l` / `svc` 为 nil 时仅跳过对应步骤（仍返回 500），避免日志依赖缺失时让 panic 直接冲垮进程；生产环境应保证二者均注入。
- 落库走 `LogService.Record` 的异步通道，panic 日志不阻塞响应；但要确保进程退出前关闭 `LogService` 以 flush 缓冲（见 [`router`](../router/router.md) 返回值说明）。
- 当前 `Message` 用 `fmt.Sprintf("panic: %v", r)`，完整堆栈只进 zap 字段；若日志中心需要按堆栈检索，可把 `stack` 一并写入 `Message` 或为 `model.Log` 新增字段。
- 不要在此处把 `r`（可能含敏感数据）原样回写给客户端，统一返回 "服务器内部错误"。
