# `internal/pkg/logger/logger.go`

封装 `go.uber.org/zap`，提供强类型结构化、分级日志能力。`Logger` 是对 `*zap.Logger` 的薄包装，统一用类型化字段构造器（`zap.String` / `zap.Int` / `zap.Duration` / `zap.Error` 等）传递上下文，**禁止用 `zap.Any` 做 `interface{}` 装箱**，从而避免反射分配、实现高性能与低内存占用。`Logger` 经依赖注入贯穿 `router` / `service` / `middleware`，调用方拿到的是强类型的 `*Logger` 而非裸 `*zap.Logger`。

## 类型

### `Logger`

```go
type Logger struct {
    zap *zap.Logger
}
```

- 字段不可导出，仅通过 `New` / `NewNop` / `Named` / `With` 构造或派生。
- 底层 `*zap.Logger` 自身 goroutine 安全，`Logger` 无可变状态，方法只读字段，故全局共享一个实例即可。
- 四个便捷方法 `Debug` / `Info` / `Warn` / `Error` 接受 `context.Context`，若其中携带 request id 则自动作为 `request_id` 字段附加，使同一请求的访问日志与业务事件可关联。

### `requestIDKey`（非导出）

```go
type requestIDKey struct{}
```

携带在 context 中的 request id 的类型化键。用空结构体类型作键可避免与其它包的字符串键冲突，`ContextWithRequestID` 写入、`log` 读取都以此键为准。

## 函数

| 函数 | 签名 | 说明 |
| --- | --- | --- |
| `New` | `(cfg config.LoggerConfig) (*Logger, error)` | 按配置组装 `zapcore.Core` 构造实例 |
| `NewNop` | `() *Logger` | 返回无操作实例，仅供测试 |
| `Named` | `(name string) *Logger` | 派生带子 logger 名称的实例 |
| `With` | `(fields ...zap.Field) *Logger` | 派生附加固定字段的实例 |
| `Debug` / `Info` / `Warn` / `Error` | `(ctx, msg string, fields ...zap.Field)` | 类型化字段记录日志，自动注入 request id |
| `Sync` | `() error` | 刷新底层缓冲 |
| `Zap` | `() *zap.Logger` | 暴露底层实例 |
| `ContextWithRequestID` | `(ctx, id string) context.Context` | 把 request id 写入 context |

### `New(cfg config.LoggerConfig) (*Logger, error)`

按配置组装 `zapcore.Core`，最终 `zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))`——即所有日志带调用者信息，`Error` 及以上自动附堆栈。配置项处理如下：

| 配置项 | 取值 | 行为 |
| --- | --- | --- |
| `Level` | `debug` / `info` / `warn` / `error`（空串等同 `info`） | 经 `parseLevel` 转 `zapcore.Level`；非法值返回错误 |
| `Encoding` | `console` / 其它 | `console` 用控制台编码，**其余一律按 `json` 处理**（含空串） |
| `Output` | `stdout` / `stderr` / 文件路径 | 经 `openWriteSyncer` 返回 `zapcore.Lock` 包裹的 `WriteSyncer`；空串等同 `stdout`；文件以 `O_CREATE\|O_WRONLY\|O_APPEND`、`0o644` 打开 |
| `TimeFormat` | `iso8601` / `rfc3339` / `epoch` | 经 `timeEncoder` 选择时间编码器；空串或未知值回退到 `iso8601` |

`EncoderConfig` 固定键名（`ts` / `level` / `logger` / `caller` / `msg` / `stacktrace`），时长用 `MillisDurationEncoder`，调用者用 `ShortCallerEncoder`，级别用 `CapitalLevelEncoder`。任一配置非法（level 解析失败、output 打开失败）都返回包装后的错误，调用方应据此启动失败而非静默降级。

### `NewNop() *Logger`

返回包装 `zap.NewNop()` 的实例，所有方法均为空操作。**仅供测试**，避免单测依赖文件系统或控制台输出。

### `Named(name string) *Logger` / `With(fields ...zap.Field) *Logger`

两者都返回新的 `*Logger`，原实例不变：

- `Named` 给底层 logger 加子名称，用于区分子系统（如 `logger.Named("image")`），日志的 `logger` 字段会带上该名称。
- `With` 附加若干固定字段，后续该实例输出的每条日志都自动携带，适合在请求/任务作用域内沉淀 `key_id` / `user` 等上下文。

### `Debug` / `Info` / `Warn` / `Error`

```go
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field)
```

四个级别共用入口 `log`：若 `ctx != nil` 且从中取到非空 request id，则 `append` 一个 `zap.String("request_id", rid)` 再交给底层对应级别方法。调用方应优先用类型化字段构造器传参（`zap.String` / `zap.Int` / `zap.Duration` / `zap.Error`），**不要用 `zap.Any`**——后者会触发反射装箱，违背本包高性能初衷。`Error` 级别由 zap 自动附堆栈。

### `Sync() error`

刷新底层缓冲。进程优雅关闭时应调用，避免缓冲中的日志丢失；写 stdout 时通常返回无意义错误，可忽略。

### `Zap() *zap.Logger`

返回底层 `*zap.Logger`，供需要原生 zap 的场景使用，例如把 `ent.Driver` 的日志回调桥接进来。日常业务日志**不应**绕过 `Logger` 的便捷方法直接调 `Zap()`，否则会丢失 request id 自动注入。

### `ContextWithRequestID(ctx, id string) context.Context`

把 request id 以 `requestIDKey{}` 为键写入 context，`id` 为空时直接返回原 ctx。`middleware.RequestID` 生成/透传 request id 后调用本函数写回 `c.Request` 的 context，后续 handler 与 service 沿该 ctx 调用 `Logger` 的便捷方法时即可自动带上 `request_id` 字段。

## 与其它包的关系

```
config.LoggerConfig ──► logger.New ──► *logger.Logger
                                       │
                ┌──────────────────────┼──────────────────────────┐
                ▼                      ▼                          ▼
   middleware.RequestID        router.New(lg)              service / api
   (ContextWithRequestID)      (依赖注入入口)             (Debug/Info/... )
```

- 依赖 [`config.LoggerConfig`](../../../config/config.md)，但不直接读 `config.Global`，所有配置经 `New` 传入。
- 由 [`router.New`](../../router/router.md) 接收 `lg *logger.Logger` 作为依赖注入入口，再分发给中间件链（`Recovery` / `Logger` 等）与各 `service` / `api`，全链路共享同一实例。
- `middleware.RequestID` 调用 `ContextWithRequestID` 写入 request id，`Logger` 的便捷方法在输出时自动取回——访问日志（见 [`middleware/logger`](../../middleware/logger.md)）与业务事件由此可按 `request_id` 串联。
- `NewNop` 供各层在单测中注入，使被测代码无需感知日志落点。

## 修改建议

- 扩展日志字段（如统一加 `service` / `version`）时，优先在 `New` 里用 `zap.AddCaller` 之外的 `zap.Fields(...)` 注入，或在 `router` 装配处用 `With` 派生一次，避免每个调用点重复传。
- 新增配置项（如 `sample_ratio`、`max_size_mb` 轮转）时改 `New` 与 `config.LoggerConfig`，并同步 [`config.md`](../../../config/config.md)；调用方签名不变。
- 切换底层实现（如换 `logrus`）时只需替换 `Logger` 内部字段与各方法实现，外部依赖 `*Logger` 的代码无需改动。
- 不要在 `Logger` 里加业务逻辑（如落库访问日志）；落库由 `service.LogService` 经 `LogRecorder` 接口承担，本包只负责向控制台/文件输出结构化日志。
