# internal/service/dashboard.go

仪表盘聚合统计的业务逻辑层，一次性汇总首页所需的图片总量、存储占用、APIkey 计数、日志总量与近 N 天上传趋势。只读、无副作用、不记业务事件。

## 类型

- `DashboardService`：持有 `imageDAO` / `apiKeyDAO` / `logDAO` 三个 DAO（沿用本仓库 service 仅依赖 DAO 的既有模式，不依赖其他 service）。
- `NewDashboardService(imageDAO, apiKeyDAO, logDAO) *DashboardService`：构造函数。
- 常量 `dashboardTrendDays = 30`：趋势窗口默认天数。

## 方法

### `Overview(ctx, days) (*model.DashboardOverview, error)`

`days <= 0` 兜底为 `dashboardTrendDays`（30）。聚合步骤：

1. 图片总量 / 存储大小：`imageDAO.Count` / `imageDAO.TotalSize`。
2. APIkey 计数：`apiKeyDAO.List` 一次拉全量后内存按 `Revoked` 分桶，得总数 / 有效 / 已吊销（避免多次 Count 往返）。
3. 日志总量：`logDAO.Count`。
4. 近 N 天上传趋势：调 `uploadTrend`。

任一 DAO 调用出错即整体返回错误（不部分降级），由 handler 转 500。

### `uploadTrend(ctx, days) ([]model.DailyCount, int, error)`

与 [`LogService.Histogram`](./log.md) 同构的按日循环：用 `time.Now().Location()` 构造今日午夜，逐日左闭右开区间调 `imageDAO.CountByRange(dayStart, dayEnd)`，缺日因查询返回 0 自然补零，结果按日期升序；同时累加合计。`CountByRange` 返回 `int64`，写入 `DailyCount.Count`（`int`）时做窄化转换（单日图片数不会溢出 `int`）。

## 与其它文件的关系

- 依赖：[`dao.ImageDAO`](../dao/dao.md)（`Count`/`TotalSize`/`CountByRange`）、[`dao.APIKeyDAO`](../dao/dao.md)（`List`）、[`dao.LogDAO`](../dao/dao.md)（`Count`）。
- 返回：[`model.DashboardOverview`](../model/dashboard.md)。
- 被调方：[`api.DashboardAPI.Overview`](../api/dashboard.md)。
- 装配：[`router.New`](../router/router.md) 构造后注入。
- 端到端说明见 [`DASHBOARD.md`](../../DASHBOARD.md)。
