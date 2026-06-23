# `cmd/server/main.go`

可执行程序入口。组合 config、router、`net/http` 三者并启动 HTTP 服务，同时处理优雅关闭。

## 关键流程

1. **加载配置**：先读环境变量 `IRIS_CONFIG`，未设置时默认使用 `config/config.yaml`，调用 `config.Load(path)`。失败直接 `log.Fatalf` 退出。
2. **设置 Gin 模式**：`gin.SetMode(cfg.Server.Mode)`，可取 `debug | release | test`。
3. **打开数据库并迁移**：`entdao.Open(cfg.Database)` 打开 SQLite（纯 Go 驱动，无需 CGO），`defer dbClient.Close()`；再按 `cfg.Database.AutoMigrate` 调 `entdao.Migrate` 建表。失败直接 `log.Fatalf`。详见 [`entdao/db.md`](../internal/dao/entdao/db.md)。
4. **构建 DAO 层**：`entdao.NewImageDAO(dbClient)` 与 `entdao.NewAPIKeyDAO(dbClient)`，分别得到 `dao.ImageDAO` 与 `dao.APIKeyDAO`。
5. **构建图片存储器**：`storage.NewSaver(cfg.Storage)` 启动期 `MkdirAll` 出 `storage.root_dir`，路径或权限有问题立刻 `log.Fatalf` 暴露。详见 [`internal/pkg/storage.md`](../internal/pkg/storage.md)。
6. **构建路由**：`router.New(cfg, imageDAO, apiKeyDAO, saver)` 注入配置、两个 DAO 与 Saver，由 router 包完成其余依赖装配（jwt manager / service / api / 限流令牌桶）。
7. **启动 HTTP 服务**：`http.Server` 监听 `cfg.Server.Host:cfg.Server.Port`，在 goroutine 里调用 `ListenAndServe`。
8. **优雅关闭**：监听 `SIGINT/SIGTERM`，收到信号后用 5 秒超时的 context 调 `srv.Shutdown(ctx)`。

## 与其它文件的关系

- 依赖 [`config`](../config/config.md)：从 yaml 读取 server / database / storage 段
- 依赖 [`internal/dao/entdao`](../internal/dao/entdao/db.md)：打开数据库、迁移、构造 DAO
- 依赖 [`internal/pkg/storage`](../internal/pkg/storage.md)：构造图片存储器
- 依赖 [`router`](../internal/router/router.md)：把 `*gin.Engine` 当作 `http.Handler` 用

## 修改建议

- 端口、Host、运行模式、数据库 DSN 都来自配置；不要在这里写死。
- 全局初始化（数据库、对象存储客户端、日志器等）在 `cfg` 加载后、`router.New` 调用前完成，并通过参数注入到 router——数据库即按此模式接入。
