// Package logger 提供基于 zap 的结构化、分级日志能力。
//
// 包装 go.uber.org/zap，统一用类型化 Field 构造器（zap.String/Int/Duration/Error 等），
// 避免 interface{} 装箱分配，实现高性能与低内存占用。Logger 经依赖注入贯穿
// router / service / middleware，调用方拿到的是强类型的 *Logger，而非裸 *zap.Logger。
package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 是对 *zap.Logger 的薄包装，提供类型化便捷方法与 context 感知。
//
// 所有便捷方法（Debug/Info/Warn/Error）接受 context.Context，若其中携带 request id
// 则自动作为字段附加，使同一请求的访问日志与业务事件可关联。
type Logger struct {
	zap *zap.Logger
}

// New 按配置构造 Logger。
func New(cfg config.LoggerConfig) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	encoding := cfg.Encoding
	if encoding != "console" {
		encoding = "json"
	}

	ws, err := openWriteSyncer(cfg.Output)
	if err != nil {
		return nil, err
	}

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     timeEncoder(cfg.TimeFormat),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var enc zapcore.Encoder
	if encoding == "console" {
		enc = zapcore.NewConsoleEncoder(encCfg)
	} else {
		enc = zapcore.NewJSONEncoder(encCfg)
	}

	core := zapcore.NewCore(enc, ws, level)
	zl := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &Logger{zap: zl}, nil
}

// NewNop 返回无操作的 Logger，仅供测试。
func NewNop() *Logger {
	return &Logger{zap: zap.NewNop()}
}

// openWriteSyncer 按 output 配置返回写入目标，并加锁避免并发写交错。
func openWriteSyncer(output string) (zapcore.WriteSyncer, error) {
	switch output {
	case "", "stdout":
		return zapcore.Lock(os.Stdout), nil
	case "stderr":
		return zapcore.Lock(os.Stderr), nil
	default:
		f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("logger: open output %q: %w", output, err)
		}
		return zapcore.Lock(f), nil
	}
}

func parseLevel(s string) (zapcore.Level, error) {
	switch s {
	case "", "info":
		return zapcore.InfoLevel, nil
	case "debug":
		return zapcore.DebugLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return 0, fmt.Errorf("logger: invalid level %q (want debug|info|warn|error)", s)
	}
}

func timeEncoder(format string) zapcore.TimeEncoder {
	switch format {
	case "epoch":
		return zapcore.EpochTimeEncoder
	case "rfc3339":
		return zapcore.RFC3339TimeEncoder
	default: // "" 或 "iso8601"
		return zapcore.ISO8601TimeEncoder
	}
}

// Named 返回带子 logger 名称的派生 Logger，用于区分子系统。
func (l *Logger) Named(name string) *Logger {
	return &Logger{zap: l.zap.Named(name)}
}

// With 返回附加了固定字段的派生 Logger。
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{zap: l.zap.With(fields...)}
}

// Debug / Info / Warn / Error 以类型化字段记录日志。
// 若 context 中携带 request id，自动作为字段附加。
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, l.zap.Debug, msg, fields)
}

// Info 记录 info 级别日志。
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, l.zap.Info, msg, fields)
}

// Warn 记录 warn 级别日志。
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, l.zap.Warn, msg, fields)
}

// Error 记录 error 级别日志（zap 会附带堆栈）。
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, l.zap.Error, msg, fields)
}

// log 是四个级别的共用入口，注入 context 中的 request id。
func (l *Logger) log(ctx context.Context, f func(string, ...zap.Field), msg string, fields []zap.Field) {
	if ctx != nil {
		if rid, ok := ctx.Value(requestIDKey{}).(string); ok && rid != "" {
			fields = append(fields, zap.String("request_id", rid))
		}
	}
	f(msg, fields...)
}

// Sync 刷新底层缓冲，优雅关闭时应调用。
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Zap 返回底层 *zap.Logger，供需要原生 zap 的场景（如 ent.Driver 日志回调）使用。
func (l *Logger) Zap() *zap.Logger {
	return l.zap
}

// requestIDKey 是携带在 context 中的 request id 的类型化键，避免键冲突。
type requestIDKey struct{}

// ContextWithRequestID 把 request id 写入 context，供 Logger 自动附加。
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey{}, id)
}
