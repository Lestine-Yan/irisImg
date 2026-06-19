# `cmd/server/main.go`

可执行程序入口。组合 config、router、`net/http` 三者并启动 HTTP 服务，同时处理优雅关闭。

## 关键流程

1. **加载配置**：先读环境变量 `IRIS_CONFIG`，未设置时默认使用 `config/config.yaml`，调用 `config.Load(path)`。失败直接 `log.Fatalf` 退出。
2. **设置 Gin 模式**：`gin.SetMode(cfg.Server.Mode)`，可取 `debug | release | test`。
3. **构建路由**：`router.New(cfg)` 把已加载的配置传入，由 router 包完成依赖装配（jwt manager / service / api）。
4. **启动 HTTP 服务**：`http.Server` 监听 `cfg.Server.Host:cfg.Server.Port`，在 goroutine 里调用 `ListenAndServe`。
5. **优雅关闭**：监听 `SIGINT/SIGTERM`，收到信号后用 5 秒超时的 context 调 `srv.Shutdown(ctx)`。

## 与其它文件的关系

- 依赖 [`config`](../config/config.md)：从 yaml 读取 server 段
- 依赖 [`router`](../internal/router/router.md)：把 `*gin.Engine` 当作 `http.Handler` 用
- 不依赖具体业务包，业务装配统一在 router 里完成

## 修改建议

- 端口、Host、运行模式都来自配置；不要在这里写死。
- 想加全局初始化（数据库、对象存储客户端、日志器等）时，建议在 `cfg` 加载后、`router.New` 调用前完成，并通过参数注入到 router。
