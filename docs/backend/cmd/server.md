# `cmd/server/main.go`

可执行程序入口。组合 config、logger、router、`net/http` 四者并启动 HTTP 服务，同时处理优雅关闭。

## 关键流程

1. **加载配置**：先读环境变量 `IRIS_CONFIG`，未设置时默认使用 `config/config.yaml`，调用 `config.Load(path)`。失败直接 `log.Fatalf` 退出。紧接着调用 `cfg.Validate()`：release 模式下校验口令/密钥非默认非空、JWT 密钥长度 ≥ 32，失败即 `log.Fatalf("insecure config: …")` 拒绝启动（fail-closed，闭合「拷贝模板未改口令即上线」的攻击链，详见 [`config.Validate`](../config/config.md)）。校验在 logger 构造前，故仍走标准库 `log` 输出到 stderr。
2. **构造结构化日志**：`logger.New(cfg.Logger)` 初始化 zap 日志器，失败直接 `log.Fatalf`；成功后 `defer lg.Sync()`。此后所有启动期日志（listening / shutting down / exited）统一走 `lg.Info`，不再用标准库 `log`。详见 [`internal/pkg/logger`](../internal/pkg/logger.md)。
3. **设置 Gin 模式**：`gin.SetMode(cfg.Server.Mode)`，可取 `debug | release | test`。
4. **打开数据库并迁移**：`entdao.Open(cfg.Database)` 打开 SQLite（纯 Go 驱动，无需 CGO），`defer dbClient.Close()`；再按 `cfg.Database.AutoMigrate` 调 `entdao.Migrate` 建表。失败直接 `log.Fatalf`。详见 [`entdao/db.md`](../internal/dao/entdao/db.md)。
5. **构建 DAO 层**：依次 `entdao.NewImageDAO(dbClient)`、`entdao.NewAPIKeyDAO(dbClient)`、`entdao.NewLogDAO(dbClient)`，分别得到 `dao.ImageDAO`、`dao.APIKeyDAO` 与 `dao.LogDAO`。
6. **构建图片存储器**：`storage.NewSaver(cfg.Storage)` 启动期 `MkdirAll` 出 `storage.root_dir`，路径或权限有问题立刻 `log.Fatalf` 暴露。详见 [`internal/pkg/storage`](../internal/pkg/storage.md)。
7. **构建路由**：`router.New(cfg, imageDAO, apiKeyDAO, logDAO, saver, lg)` 注入配置、三个 DAO、Saver 与 logger，由 router 包完成其余依赖装配（jwt manager / service / api / 限流令牌桶 / 异步日志服务），并返回 `(r *gin.Engine, logSvc service.LogService)`。`logSvc` 留给关闭阶段 flush 异步日志缓冲。
8. **启动 HTTP 服务**：`http.Server` 监听 `cfg.Server.Host:cfg.Server.Port`，在 goroutine 里调用 `ListenAndServe`；进入监听前用 `lg.Info` 打印 `server listening` 及 `addr`。
9. **优雅关闭**：监听 `SIGINT/SIGTERM`，收到信号后用 `lg.Info` 打印 `shutting down server`，再用 5 秒超时的 context 调 `srv.Shutdown(ctx)`。**`Shutdown` 超时不再用 `log.Fatalf`**：`Fatalf` 会 `os.Exit` 跳过 `logSvc.Close()` 与 `defer dbClient.Close()`/`lg.Sync()`，导致缓冲日志丢失、DB/logger 未优雅关闭；改为 `lg.Error` 记录错误后继续往下执行 `logSvc.Close()` flush 异步日志缓冲，把收尾交给 `defer`。`Record` 在 `Close` 后仍安全（`done` 通道保护，`buf` 永不关闭，在途 handler 即便仍在 `Record` 也不会 panic）。最后由 `defer dbClient.Close()` 关库，并以 `lg.Info` 打印 `server exited` 收尾。顺序为 `srv.Shutdown -> logSvc.Close() -> defer dbClient.Close()`。

## 与其它文件的关系

- 依赖 [`config`](../config/config.md)：从 yaml 读取 server / database / storage / logger 段
- 依赖 [`internal/pkg/logger`](../internal/pkg/logger.md)：构造结构化日志器，启动期日志与 router 注入共用同一个 `lg`
- 依赖 [`internal/dao/entdao`](../internal/dao/entdao/db.md)：打开数据库、迁移、构造 DAO（含 `LogDAO`）
- 依赖 [`internal/pkg/storage`](../internal/pkg/storage.md)：构造图片存储器
- 依赖 [`router`](../internal/router/router.md)：注入 DAO / Saver / logger，把 `*gin.Engine` 当作 `http.Handler` 用，并取回 `logSvc` 用于关闭阶段 flush

## 修改建议

- 端口、Host、运行模式、数据库 DSN、日志配置都来自配置；不要在这里写死。
- 全局初始化（数据库、对象存储客户端、日志器等）在 `cfg` 加载后、`router.New` 调用前完成，并通过参数注入到 router——数据库、logger 即按此模式接入。
- 关闭顺序不能随意调换：必须先 `srv.Shutdown` 停止接收新请求，再 `logSvc.Close()` flush 缓冲里残留的访问日志，最后才让 `defer dbClient.Close()` 关库，否则可能丢日志或 flush 时访问已关闭的连接。
