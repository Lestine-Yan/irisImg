# ent/schema/image.go

定义图床核心实体 **Image**（图片元信息）的 Ent schema。真实图片二进制存放在本地 `data/` 目录（或后续对象存储），数据库只保存检索、展示、去重所需的元数据。

`go generate ./ent` 会基于本 schema 生成 `backend/ent/` 下的类型安全客户端代码（`client.go`、`image*.go` 等，均为生成产物，不手工编辑）。

## 字段

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `filename` | string | 非空 | 上传时的原始文件名 |
| `stored_path` | string | 非空、唯一 | 相对存储根目录的落盘路径 |
| `url` | string | 非空 | 对外访问地址 |
| `size` | int64 | ≥0 | 文件大小（字节） |
| `mime_type` | string | 非空 | MIME 类型，如 `image/png` |
| `width` | int | ≥0，默认 0 | 宽度（像素），未知为 0 |
| `height` | int | ≥0，默认 0 | 高度（像素），未知为 0 |
| `hash` | string | 非空 | 内容哈希，用于秒传 / 去重 |
| `created_at` | time | 不可变，默认 `time.Now` | 创建时间 |

## 索引

- `hash`：按内容哈希去重 / 秒传查询。
- `created_at`：按创建时间倒序分页。

## 关联关系

当前无 Edges。

## 调用关系

schema 仅供 Ent 代码生成使用；运行时由 [`internal/dao/entdao/image.go`](../../internal/dao/entdao/image.md) 通过生成的 `ent.Client` 读写。
