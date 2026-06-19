# `internal/router/router.go`

整个后端唯一的依赖装配点和路由注册点。`cmd/server/main.go` 只调一次 `router.New(cfg)` 拿到 `*gin.Engine` 即可启动。

## 函数

### `New(cfg *config.Config) *gin.Engine`

- 使用 `gin.New()` 而非 `gin.Default()`，自己挂中间件以保留控制权：`Recovery → Logger → CORS`。
- **依赖装配**（按 dao → service → api 的顺序，但当前没有 dao）：
  1. `jwtMgr := jwt.NewManager(cfg.Auth.JWT)`
  2. `authSvc := service.NewAuthService(cfg.Auth, jwtMgr)`
  3. `authAPI := api.NewAuthAPI(authSvc)`
- **路由注册**：所有业务接口挂在 `/api/v1` 下。
  - 公开：`GET /ping`、`POST /auth/login`
  - 受保护（`middleware.JWTAuth(jwtMgr)`）：`GET /auth/me`，未来图片接口同组挂载。

## 路由地图

```
/api/v1
├── GET  /ping                   公开    api.Ping
├── POST /auth/login             公开    AuthAPI.Login
└── ── (中间件: JWTAuth)
    └── GET  /auth/me            受保护  AuthAPI.Me
```

## 修改建议

- **保持 router 是依赖装配的唯一入口**：新加业务包时，把 `New` 写进各自的构造函数，由 router 在这里组装；不要在 handler 或 service 里直接 `new` 出依赖。
- 受保护路由全部挂到 `protected := v1.Group("", middleware.JWTAuth(jwtMgr))` 下；新增公开路由直接挂 `v1`。
- 项目变大、依赖关系复杂后，可以引入 [`google/wire`](https://github.com/google/wire) 等工具做编译期 DI，但当前规模手动注入是最直观、最低心智成本的做法。
