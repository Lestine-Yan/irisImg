# internal/model/dashboard.go

仪表盘聚合统计接口的返回 DTO，承载首页所需的全部指标。由 [`service.DashboardService.Overview`](../service/dashboard.md) 组装，经 [`api.DashboardAPI.Overview`](../api/dashboard.md) 返回给前端仪表盘页面。

## 类型

### `DashboardOverview`

| 字段 | 类型 | 说明 |
|------|------|------|
| `ImagesTotal` | `int64` | 图片总量（无过滤） |
| `StorageBytes` | `int64` | 全部图片 `size` 之和（字节），取自 DB `SUM(size)`，空表为 0 |
| `APIKeysTotal` | `int` | 密钥总数（含已吊销；已删除为物理删除，不在统计内） |
| `APIKeysActive` | `int` | 未吊销的有效密钥数 |
| `APIKeysRevoked` | `int` | 已吊销密钥数 |
| `LogsTotal` | `int64` | 日志总量 |
| `RecentUploadTrend` | `[]DailyCount` | 近 N 天每日新增图片（升序、缺日补零），复用 [`DailyCount`](./log.md)，结构同日志直方图 buckets |
| `RecentUploadTotal` | `int` | 近 N 天新增图片合计 |
| `Days` | `int` | 趋势窗口天数（默认 30），回显给前端文案 |

## 设计要点

- 复用 [`model.DailyCount`](./log.md)（`{Date, Count}`）作为趋势单元，使前端可共用 `LogsHistogram` 组件，无需为图片趋势另造类型。
- 存储大小取 DB `SUM(size)` 而非文件系统遍历：`images.size` 字段已是单一事实来源，一条 SQL 完成；文件系统遍历慢且含孤儿文件，仅适合作可选的「磁盘实际占用」辅助指标。
- 近 N 天趋势以 `Image.CreatedAt` 为准，**不**沿用 `image.upload` 日志事件数：秒传（同 hash 已存在）也会记一次事件导致重复计数，且日志可被 ClearAll 清空不持久。

## 与其它文件的关系

- 组装方：[`internal/service/dashboard.go`](../service/dashboard.md)。
- 暴露方：[`internal/api/dashboard.go`](../api/dashboard.md)（`GET /admin/dashboard`）。
- 复用类型：[`model.DailyCount`](./log.md)。
- 端到端说明见 [`DASHBOARD.md`](../../DASHBOARD.md)。
