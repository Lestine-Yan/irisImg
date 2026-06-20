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
	"github.com/Lestine-Yan/irisImg/backend/internal/router"
	"github.com/gin-gonic/gin"
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

	// 2. 设置 gin 运行模式
	gin.SetMode(cfg.Server.Mode)

	// 3. 打开数据库（SQLite，纯 Go 驱动，无需 CGO）并按需自动迁移
	dbClient, err := entdao.Open(cfg.Database)
	if err != nil {
		log.Fatalf("open database failed: %v", err)
	}
	defer dbClient.Close()
	if err := entdao.Migrate(context.Background(), dbClient, cfg.Database); err != nil {
		log.Fatalf("migrate database failed: %v", err)
	}

	// 4. 构建 DAO 层
	imageDAO := entdao.NewImageDAO(dbClient)

	// 5. 构建路由（注入 DAO 依赖）
	r := router.New(cfg, imageDAO)

	// 6. 启动 HTTP 服务，并支持优雅关闭
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 等待终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited")
}
