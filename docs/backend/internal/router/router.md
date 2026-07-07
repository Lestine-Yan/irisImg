# `internal/router/router.go`

整个后端唯一的依赖装配点和路由注册点。`cmd/server/main.go` 打开数据库、构造 DAO、构造 storage.Saver 后调用 `router.New(cfg, imageDAO, apiKeyDAO, saver)` 拿到 `*gin.Engine` 即可启动。

## 函数

### `New(cfg *config.Config, imageDAO dao.ImageDAO, apiKeyDAO dao.APIKeyDAO, saver *storage.Saver) *gin.Engine`

- 使用 `gin.New()` 而非 `gin.Default()`，自己挂中间件以保留控制权：`Recovery → Logger → CORS`。
- `r.Static("/imgs", cfg.Storage.RootDir)`：开发期由后端直接 serve 图片落盘目录，供前端加载 `/imgs/<rel>`；生产环境建议由 Nginx 反代 `/imgs/`（见 [`IMAGE.md`](../../IMAGE.md)），此处仅兜底。
- `imageDAO` / `apiKeyDAO` 由调用方基于已打开的数据库注入（见 [`dao.md`](../dao/dao.md)）。
- `saver` 由调用方基于 `cfg.Storage` 提前构造（见 [`pkg/storage.md`](../pkg/storage.md)），启动期 `MkdirAll` 暴露路径/权限问题。
- **依赖装配**（按 dao → service → api 的顺序）：
  1. `jwtMgr := jwt.NewManager(cfg.Auth.JWT)`
  2. `authSvc := service.NewAuthService(cfg.Auth, jwtMgr)` → `authAPI := api.NewAuthAPI(authSvc)`
  3. `apiKeySvc := service.NewAPIKeyService(apiKeyDAO, imageDAO, saver)` → `apiKeyAPI := api.NewAPIKeyAPI(apiKeySvc, authSvc)`（`imageDAO`/`saver` 供删除密钥级联清理，`authSvc` 供吊销/删除密码二次确认）
  4. `imageSvc := service.NewImageService(imageDAO, saver, cfg.Storage)` → `imageAPI := api.NewImageAPI(imageSvc)`
  5. `rateStore := ratelimit.NewStore(cfg.APIKey.RateLimitPerMinute)` —— 按密钥维度限流的内存令牌桶。
- **路由注册**：所有业务接口挂在 `/api/v1` 下。
  - 公开：`GET /ping`、`POST /auth/login`
  - 受保护（`middleware.JWTAuth(jwtMgr)`）：`GET /auth/me`、`GET /admin/images`（后台图片列表）、`POST /admin/images`（后台直传上传，`key_id` 留空），均供内容中心；其下再套 `keys := protected.Group("/apikeys", middleware.HTTPSOnly(cfg.APIKey.HTTPSOnly))` 挂密钥管理接口（**JWT + HTTPS** 双重保护）；其中吊销（`POST /:id/revoke`）与删除（`DELETE /:id`）为敏感操作，handler 内部还会校验请求体携带的账号密码做二次确认。`/admin/images` 与对外 `/images`（API Key 鉴权）路径不同，避免同方法同路径重复注册冲突。
  - API 密钥保护（`middleware.APIKeyAuth(apiKeySvc, rateStore)`，独立于 JWT）：`images` 组。

## 路由地图

```
/imgs/*filepath                  静态        r.Static(cfg.Storage.RootDir)  （开发期 serve，生产由 Nginx 反代）

/api/v1
├── GET  /ping                   公开        api.Ping
├── POST /auth/login             公开        AuthAPI.Login
├── ── (中间件: JWTAuth)
│   ├── GET  /auth/me            受保护      AuthAPI.Me
│   ├── GET  /admin/images       受保护      ImageAPI.ListAdmin
│   ├── POST /admin/images       受保护      ImageAPI.CreateAdmin (后台直传，key_id=nil)
│   └── ── (中间件: HTTPSOnly)   /apikeys
│       ├── POST   ""            JWT+HTTPS   APIKeyAPI.Create
│       ├── GET    ""            JWT+HTTPS   APIKeyAPI.List
│       ├── PATCH  "/:id"        JWT+HTTPS   APIKeyAPI.Rename
│       ├── POST   "/:id/reset"  JWT+HTTPS   APIKeyAPI.Reset
│       ├── POST   "/:id/revoke" JWT+HTTPS   APIKeyAPI.Revoke（软吊销，需密码）
│       └── DELETE "/:id"        JWT+HTTPS   APIKeyAPI.Delete（硬删+级联删图，需密码）
└── ── (中间件: APIKeyAuth)      /images
    ├── GET  ""                  API Key     ImageAPI.List   (占位)
    └── POST ""                  API Key     ImageAPI.Create (已实现，需 readwrite)
```

API 密钥鉴权链路与权限矩阵的完整说明见 [`APIKEY.md`](../../APIKEY.md)；图片上传链路与 Nginx 反代约定见 [`IMAGE.md`](../../IMAGE.md)。

## 修改建议

- **保持 router 是依赖装配的唯一入口**：新加业务包时，把 `New` 写进各自的构造函数，由 router 在这里组装；不要在 handler 或 service 里直接 `new` 出依赖。
- 受保护路由全部挂到 `protected := v1.Group("", middleware.JWTAuth(jwtMgr))` 下；新增公开路由直接挂 `v1`。
- 项目变大、依赖关系复杂后，可以引入 [`google/wire`](https://github.com/google/wire) 等工具做编译期 DI，但当前规模手动注入是最直观、最低心智成本的做法。
