# `config/config.go` 与 `config/config.yaml`

集中管理整个后端的运行参数。源码用 `gopkg.in/yaml.v3` 把 YAML 解析进 Go 结构体；其它包通过参数注入或包级变量 `config.Global` 拿到配置。

## 结构体层级

```text
Config
├── Server   (ServerConfig)        host / port / mode / trusted_proxies
├── App      (AppConfig)           name / version
├── Auth     (AuthConfig)          username / password
│   └── JWT  (JWTConfig)           secret / issuer / expire_hours
├── Database (DatabaseConfig)      driver / dsn / auto_migrate
├── APIKey   (APIKeyConfig)        rate_limit_per_minute / https_only
├── Storage  (StorageConfig)       root_dir / public_base_url / max_upload_size_mb / allowed_mime_types
├── CORS     (CORSConfig)          allow_origins
└── Logger   (LoggerConfig)        level / encoding / output / time_format
```

| 字段 | 类型 | 默认 | 用途 |
| --- | --- | --- | --- |
| `server.host` | string | `0.0.0.0` | 监听地址 |
| `server.port` | int | `8080` | 监听端口 |
| `server.mode` | string | `debug` | Gin 运行模式 (`debug | release | test`) |
| `server.trusted_proxies` | []string | `[127.0.0.1/8, ::1]` | 受信任反代 CIDR 网段；HTTPSOnly 仅对来自此列表的请求认 `X-Forwarded-Proto`，否则只认 `c.Request.TLS`。跨机反代追加反代 CIDR。空列表退化为只认 TLS。详见 [`middleware/https`](../internal/middleware/https.md) |
| `app.name` | string | `irisImg` | 应用名（出现在 `/ping` 响应里） |
| `app.version` | string | `0.1.0` | 版本号 |
| `auth.username` | string | `admin` | 唯一用户名（release 模式下 `Validate` 要求非空） |
| `auth.password` | string | `admin123` | 明文密码（release 模式下默认值 `admin123` 与空值会被 `Validate` 拒绝启动） |
| `auth.jwt.secret` | string | 占位字符串 | HS256 签名密钥；release 模式下默认占位串 / 空 / 长度 <32 会被 `Validate` 拒绝启动 |
| `auth.jwt.issuer` | string | `irisImg` | 写入 token 的 `iss` 字段 |
| `auth.jwt.expire_hours` | int | `24` | token 有效期；`<=0` 时 jwt 包会回退到 24 小时 |
| `database.driver` | string | `sqlite` | 数据库后端，目前仅支持 `sqlite` |
| `database.dsn` | string | `data/irisImg.db?...` | SQLite 文件路径 + pragma；相对路径相对进程工作目录 |
| `database.auto_migrate` | bool | `true` | 启动时是否自动建表 / 升级表结构 |
| `apikey.rate_limit_per_minute` | int | `100` | 单密钥默认限流阈值（次/分钟）；`<=0` 回退 100。密钥自身 `rate_limit` 为 0 时沿用此值 |
| `apikey.https_only` | bool | `false` | 为 true 时密钥管理等敏感接口要求 HTTPS（后端通过 `X-Forwarded-Proto` 二次校验 Nginx 反代）；本地开发置 false |
| `storage.root_dir` | string | `data/imgs` | 图片落盘根目录（相对路径相对进程工作目录，部署建议改为绝对路径）。启动时 `MkdirAll`。**fail-fast**：[`storage.NewSaver`](../internal/pkg/storage.md) 拒绝将其配成后端工作目录本身或其祖先（`.` / `..` / `/` / 工作目录父级），避免 `/imgs` 静态服务未认证暴露 `config.yaml` / 数据库 / 源码；[`serveImages`](../internal/router/static.md) 另有图片扩展名白名单兜底 |
| `storage.public_base_url` | string | `""` | 对外访问 URL 前缀。空 -> 返回 `/imgs/<rel>`（前端/Nginx 同域反代）；非空（如 `https://img.example.com`，结尾不带斜杠）-> 返回绝对地址。**裸域名（无 `http(s)://`）会被 `NewSaver` 自动补 `https://`**，但建议显式带协议或留空，避免依赖隐式行为 |
| `storage.max_upload_size_mb` | int | `20` | 单次上传字节上限（MiB）；`<=0` 回退 20 |
| `storage.allowed_mime_types` | []string | `image/png, image/jpeg, image/gif, image/webp` | 真实 MIME 白名单。后端用 `http.DetectContentType` 嗅探，不信任客户端 `Content-Type` |
| `cors.allow_origins` | []string | `["*"]` | 跨域来源白名单。nil 补 `["*"]`（开发）；显式空 `[]` 关闭跨域（生产同域）。release 模式 `Validate` 拒 `*`。详见 [`middleware/cors`](../internal/middleware/cors.md) |
| `logger.level` | string | `info` | 日志级别 (`debug | info | warn | error`) |
| `logger.encoding` | string | `json` | 输出编码 (`json | console`) |
| `logger.output` | string | `stdout` | 输出目标 (`stdout | stderr | <文件路径>`) |
| `logger.time_format` | string | `iso8601` | 时间字段格式 (`iso8601 | rfc3339 | epoch`) |

> 上表默认值（除 `database.dsn` / `storage.root_dir` 等 fail-fast 字段与 `auth` 安全字段外）由 `ApplyDefaults` 在 `Load` 阶段集中补齐，消费方不再各自兜底。机制与边界见下文 [`ApplyDefaults`](#applydefaults)。

> `database` 由 [`internal/dao/entdao`](../internal/dao/entdao/db.md) 消费。DSN 默认带 `busy_timeout` / `journal_mode(WAL)` / `foreign_keys(on)` 三个 pragma；其中 `foreign_keys` 是 Ent 自动迁移的前置要求（缺省时代码会自动补上）。`data/` 下的数据库文件不要提交仓库。

> `apikey` 段由 [`router`](../internal/router/router.md) 消费：`rate_limit_per_minute` 注入 [`ratelimit.Store`](../internal/pkg/ratelimit.md)，`https_only` 注入 [`middleware.HTTPSOnly`](../internal/middleware/https.md)。特性级说明见 [`APIKEY.md`](../APIKEY.md)。

> `storage` 段由 [`internal/pkg/storage`](../internal/pkg/storage.md) 与 [`internal/service/image`](../internal/service/image.md) 消费。**部署提示**：`root_dir` 与 Nginx 静态 `location /imgs/` 的物理路径要保持一致；`public_base_url` 若启用，需与 Nginx 暴露的图片域名一致。特性级说明见 [`IMAGE.md`](../IMAGE.md)。

> `logger` 段由 [`internal/service/log`](../internal/service/log.md) 消费以初始化 zap。**职责边界**：访问日志的异步落库由 [`LogService`](../internal/service/log.md) 单独控制，**此处仅控制 zap 输出**到 stdout/stderr/文件的部分，用于运维采集。所有字段缺省时按 `info/json/stdout/iso8601` 处理。

## 关键函数

### `Load(path string) (*Config, error)`

- 读文件、`yaml.Unmarshal`，任一步出错都用 `fmt.Errorf("…: %w", err)` 包装后返回。
- 解析成功后调用 `ApplyDefaults()` 补齐缺省字段（单一事实源），再把指针赋给包级变量 `Global`，方便不便走依赖注入的小工具（如 `api/ping.go`）直接读取 `app.name / app.version`。
- 业务代码请优先通过参数接收配置，`Global` 仅作便捷出口。

### `ApplyDefaults()`

把缺失（零值）的配置项补齐为合理默认值，建立「缺省默认」的单一事实源。由 `Load` 在 `yaml.Unmarshal` 后调用；导出是为了供测试与未来从 env/flag 加载配置的路径复用。

**兜底范围**（零值时填默认）：

| 字段 | 默认 |
| --- | --- |
| `server.host` / `server.port` / `server.mode` | `0.0.0.0` / `8080` / `debug` |
| `server.trusted_proxies` | `[127.0.0.1/8, ::1]` |
| `app.name` / `app.version` | `irisImg` / `0.1.0` |
| `database.driver` | `sqlite` |
| `auth.jwt.issuer` / `auth.jwt.expire_hours` | `irisImg` / `24` |
| `apikey.rate_limit_per_minute` | `100` |
| `storage.max_upload_size_mb` | `20` |
| `storage.allowed_mime_types` | `[png, jpeg, gif, webp]` |
| `cors.allow_origins` | `["*"]` |
| `logger.level` / `logger.encoding` / `logger.output` / `logger.time_format` | `info` / `json` / `stdout` / `iso8601` |

**刻意不兜底**：

- **fail-fast 字段** `database.dsn` / `storage.root_dir`：空值时由 [`entdao.Open`](../internal/dao/entdao/db.md) / [`storage.NewSaver`](../internal/pkg/storage.md) 返回 error、`main.go` `log.Fatalf` 拒绝启动。错配会把数据/图片写到错误位置，比启动失败更难诊断，故强制显式配置而非静默填默认。
- **安全字段** `auth.username` / `auth.password` / `auth.jwt.secret`：绝不兜底默认值，否则会绕过 `Validate` 的 fail-closed（release 模式拒绝默认/空口令与弱密钥启动）。
- **bool 字段** `database.auto_migrate` / `apikey.https_only`：零值无法区分「缺失」与「显式 false」，兜底会覆盖用户意图。`auto_migrate` 依赖 `config.yaml` 模板带 `true`。

> `storage.allowed_mime_types` 用 `== nil` 判断而非 `len == 0`，区分「缺失/null」（补默认白名单，修复旧配置缺该键时上传被全拒的静默 bug）与「显式空列表」（保留空，尊重用户「禁止所有上传」的意图）。

与 `Validate` 的关系：`Load -> ApplyDefaults ->`（`main` 调）`Validate`，二者职责互补不重叠。

### `Validate() error`

- 仅在 `server.mode == "release"` 下强制，debug/test 放过（保持开发开箱即跑）。由 [`cmd/server/main.go`](../cmd/server/main.md) 在 logger 构造前调用，失败即 `log.Fatalf` 拒绝启动（fail-closed）。
- 命中以下任一即返回 error：`auth.username` 为空；`auth.password` 为空或等于默认值 `admin123`；`auth.jwt.secret` 为空、等于默认占位串 `please-change-me-to-a-long-random-string`、或长度 < 32；`cors.allow_origins` 含 `*`（强制配确切域名或留空关闭跨域）。
- 闭合「拷贝 `config.yaml.example` 未改口令即上线」的攻击链：生产模板的默认值即取自被拒集合，故未改口令无法以 release 模式启动。用户名 `admin` 本身合法（只要密码非默认），不校验。
- 闭合「通配 CORS 上线」攻击链：`Allow-Origin: *` 配合 `Authorization` 透传是定时炸弹（引入 cookie 鉴权即升高危），release 拒 `*` 强制生产配确切域名或留空。

## YAML 示例

`config/config.yaml` 的完整内容如下，含 `logger` 段：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"   # debug | release | test
  trusted_proxies:   # 受信任反代网段，HTTPSOnly 仅对此列表的请求认 X-Forwarded-Proto
    - "127.0.0.1/8"
    - "::1"

app:
  name: "irisImg"
  version: "0.1.0"

# 持久化配置。当前使用 SQLite（纯 Go 的 modernc.org/sqlite 驱动，无需 CGO）。
# dsn 即数据库文件路径，相对路径相对于后端进程的工作目录；
# data/ 目录下的运行时数据不要提交到仓库。
database:
  driver: "sqlite"
  dsn: "data/irisImg.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)"
  auto_migrate: true

# 个人图床仅服务于本人，用户名/密码直接写在配置里。
# 部署时请务必修改默认值，并将 jwt.secret 替换为一段足够长的随机字符串。
auth:
  username: "admin"
  password: "admin123"
  jwt:
    secret: "please-change-me-to-a-long-random-string"
    issuer: "irisImg"
    expire_hours: 24

# API 密钥鉴权配置。密钥用于外部程序「申请图片 / 添加图片」，独立于后台 JWT 登录。
# https_only 在生产环境（Nginx HTTPS 反代）请置为 true，后端会校验 X-Forwarded-Proto；
# 本地开发走纯 HTTP，保持 false。
apikey:
  rate_limit_per_minute: 100
  https_only: false

# 图片落盘相关配置。
# root_dir 相对路径相对于后端进程工作目录，默认 data/imgs；
# 部署到服务器时一般改成绝对路径，并交给 Nginx 在同样的路径上做静态反代。
# public_base_url 为空时返回的 URL 是相对路径（前端/Nginx 同域代理即可）；
# 部署独立图片域名时填 https://img.example.com 这类绝对前缀（须带协议，裸域名会自动补 https://），结尾不要带斜杠。
# max_upload_size_mb 限制单次上传字节数；allowed_mime_types 是真实 MIME 白名单
# （后端用 http.DetectContentType 嗅探，不信任客户端 Content-Type）。
storage:
  root_dir: "data/imgs"
  public_base_url: ""
  max_upload_size_mb: 20
  allowed_mime_types:
    - "image/png"
    - "image/jpeg"
    - "image/gif"
    - "image/webp"

# 跨域配置。生产同域部署无跨域需求，留空关闭；开发联调填 ["*"]。
# release 模式下 Validate 拒绝 "*"，强制配确切域名或留空。
cors:
  allow_origins:
    - "*"

# 结构化日志（zap）配置。
# level: debug|info|warn|error；encoding: json|console；
# output: stdout|stderr|<文件路径>；time_format: iso8601|rfc3339|epoch。
# 访问日志同时异步落库供日志中心查询，此处控制的是 zap 输出到 stdout/文件的部分。
logger:
  level: "info"
  encoding: "json"
  output: "stdout"
  time_format: "iso8601"
```

## 修改建议

- 新增配置项时三处都要改：①yaml 文件加默认值；②对应结构体加字段；③在 README 或本文件补字段说明。
- 不要把任何密钥提交到仓库；正式部署用 `IRIS_CONFIG` 指向服务器本地的私有副本。
