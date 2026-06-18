# irisImg backend

基于 Gin 的图床后端，采用按层划分的项目结构。

## 目录结构

```
backend/
├── cmd/server/main.go          # 程序入口
├── config/                     # 配置加载与默认配置
│   ├── config.go
│   └── config.yaml
├── internal/
│   ├── router/                 # 路由注册 + 依赖装配
│   ├── api/                    # 控制器：参数校验、调用 service
│   ├── service/                # 业务逻辑层
│   ├── dao/                    # 数据访问层（当前为内存实现）
│   ├── model/                  # 实体与 DTO
│   ├── middleware/             # 中间件（CORS、日志）
│   └── pkg/response/           # 统一响应封装
├── go.mod
└── README.md
```

调用链：`router → middleware → api → service → dao → model`

## 运行

```bash
cd backend
go mod tidy        # 下载依赖（gin、yaml.v3）
go run ./cmd/server
```

默认监听 `0.0.0.0:8080`，配置可通过环境变量 `IRIS_CONFIG` 指定路径。

## 接口示例

```bash
# 健康检查
curl http://localhost:8080/api/v1/ping

# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com"}'

# 查询用户
curl http://localhost:8080/api/v1/users/1

# 用户列表
curl http://localhost:8080/api/v1/users
```

## 后续可演进方向

- DAO 层接入数据库（gorm + sqlite/mysql/postgres），保留接口不变即可平滑替换
- 配置加载切换为 viper，支持环境变量与多格式
- 日志中间件替换为 zap，引入 trace id
- 引入 wire 做依赖注入
- 业务模块新增（图片上传、token 鉴权等）按 `model → dao → service → api → router` 的顺序补齐
