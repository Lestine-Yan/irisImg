# `config/config.go` 与 `config/config.yaml`

集中管理整个后端的运行参数。源码用 `gopkg.in/yaml.v3` 把 YAML 解析进 Go 结构体；其它包通过参数注入或包级变量 `config.Global` 拿到配置。

## 结构体层级

```text
Config
├── Server   (ServerConfig)        host / port / mode
├── App      (AppConfig)           name / version
├── Auth     (AuthConfig)          username / password
│   └── JWT  (JWTConfig)           secret / issuer / expire_hours
├── Database (DatabaseConfig)      driver / dsn / auto_migrate
└── APIKey   (APIKeyConfig)        rate_limit_per_minute / https_only
```

| 字段 | 类型 | 默认 | 用途 |
| --- | --- | --- | --- |
| `server.host` | string | `0.0.0.0` | 监听地址 |
| `server.port` | int | `8080` | 监听端口 |
| `server.mode` | string | `debug` | Gin 运行模式 (`debug | release | test`) |
| `app.name` | string | `irisImg` | 应用名（出现在 `/ping` 响应里） |
| `app.version` | string | `0.1.0` | 版本号 |
| `auth.username` | string | `admin` | 唯一用户名（部署时务必修改） |
| `auth.password` | string | `admin123` | 明文密码（部署时务必修改） |
| `auth.jwt.secret` | string | 占位字符串 | HS256 签名密钥；上线必须替换为足够长的随机字符串 |
| `auth.jwt.issuer` | string | `irisImg` | 写入 token 的 `iss` 字段 |
| `auth.jwt.expire_hours` | int | `24` | token 有效期；`<=0` 时 jwt 包会回退到 24 小时 |
| `database.driver` | string | `sqlite` | 数据库后端，目前仅支持 `sqlite` |
| `database.dsn` | string | `data/irisImg.db?...` | SQLite 文件路径 + pragma；相对路径相对进程工作目录 |
| `database.auto_migrate` | bool | `true` | 启动时是否自动建表 / 升级表结构 |
| `apikey.rate_limit_per_minute` | int | `100` | 单密钥默认限流阈值（次/分钟）；`<=0` 回退 100。密钥自身 `rate_limit` 为 0 时沿用此值 |
| `apikey.https_only` | bool | `false` | 为 true 时密钥管理等敏感接口要求 HTTPS（后端通过 `X-Forwarded-Proto` 二次校验 Nginx 反代）；本地开发置 false |

> `database` 由 [`internal/dao/entdao`](../internal/dao/entdao/db.md) 消费。DSN 默认带 `busy_timeout` / `journal_mode(WAL)` / `foreign_keys(on)` 三个 pragma；其中 `foreign_keys` 是 Ent 自动迁移的前置要求（缺省时代码会自动补上）。`data/` 下的数据库文件不要提交仓库。

> `apikey` 段由 [`router`](../internal/router/router.md) 消费：`rate_limit_per_minute` 注入 [`ratelimit.Store`](../internal/pkg/ratelimit.md)，`https_only` 注入 [`middleware.HTTPSOnly`](../internal/middleware/https.md)。特性级说明见 [`APIKEY.md`](../APIKEY.md)。

## 关键函数

### `Load(path string) (*Config, error)`

- 读文件、`yaml.Unmarshal`，任一步出错都用 `fmt.Errorf("…: %w", err)` 包装后返回。
- 解析成功后**同时**把指针赋给包级变量 `Global`，方便不便走依赖注入的小工具（如 `api/ping.go`）直接读取 `app.name / app.version`。
- 业务代码请优先通过参数接收配置，`Global` 仅作便捷出口。

## 修改建议

- 新增配置项时三处都要改：①yaml 文件加默认值；②对应结构体加字段；③在 README 或本文件补字段说明。
- 不要把任何密钥提交到仓库；正式部署用 `IRIS_CONFIG` 指向服务器本地的私有副本。
