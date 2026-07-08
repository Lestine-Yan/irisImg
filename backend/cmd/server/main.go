package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao/entdao"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/logger"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/storage"
	"github.com/Lestine-Yan/irisImg/backend/internal/router"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置（路径可通过环境变量覆盖）
	configPath := os.Getenv("IRIS_CONFIG")
	if configPath == "" {
		configPath = "config/config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// 2. 构造结构化日志（zap），其后所有启动日志走 logger
	lg, err := logger.New(cfg.Logger)
	if err != nil {
		log.Fatalf("init logger failed: %v", err)
	}
	defer lg.Sync()

	// 3. 设置 gin 运行模式
	gin.SetMode(cfg.Server.Mode)

	// 4. 打开数据库（SQLite，纯 Go 驱动，无需 CGO）并按需自动迁移
	dbClient, err := entdao.Open(cfg.Database)
	if err != nil {
		log.Fatalf("open database failed: %v", err)
	}
	defer dbClient.Close()
	if err := entdao.Migrate(context.Background(), dbClient, cfg.Database); err != nil {
		log.Fatalf("migrate database failed: %v", err)
	}

	// 5. 构建 DAO 层（含日志 DAO）
	imageDAO := entdao.NewImageDAO(dbClient)
	apiKeyDAO := entdao.NewAPIKeyDAO(dbClient)
	logDAO := entdao.NewLogDAO(dbClient)

	// 6. 构建图片存储器（提前 MkdirAll，权限/路径问题在启动时就暴露）
	saver, err := storage.NewSaver(cfg.Storage)
	if err != nil {
		log.Fatalf("init storage failed: %v", err)
	}

	// 7. 构建路由（注入 DAO / Saver / logger 依赖），取回 logSvc 供优雅关闭 flush
	r, logSvc := router.New(cfg, imageDAO, apiKeyDAO, logDAO, saver, lg)

	// 8. 启动 HTTP 服务，并支持优雅关闭
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		lg.Info(context.Background(), "server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 等待终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	lg.Info(context.Background(), "shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Shutdown 超时只记错误不 Fatalf：Fatalf 会 os.Exit 跳过 logSvc.Close 与 defer，
	// 导致缓冲日志丢失、DB/logger 未优雅关闭。这里继续往下执行 Close，让 defer 收尾。
	if err := srv.Shutdown(ctx); err != nil {
		lg.Error(context.Background(), "server forced to shutdown", zap.Error(err))
	}
	// 停止接收新请求后，flush 异步日志缓冲，再让 defer 关闭数据库。
	// Record 在 Close 后安全（done 通道保护，绝不 panic），即便有在途 handler 仍在 Record。
	logSvc.Close()
	lg.Info(context.Background(), "server exited")
}
