package model

// DashboardOverview 是仪表盘聚合统计接口的返回结构，一次性承载首页所需的全部指标。
//
// 由 DashboardService.Overview 聚合各 DAO 的查询结果组装而成，前端仪表盘页面据此渲染
// 统计卡片、近 N 天新增图片趋势图与快捷入口。RecentUploadTrend 复用 DailyCount，
// 与日志中心直方图的 buckets 结构完全一致，便于前端共用 LogsHistogram 组件。
type DashboardOverview struct {
	// ImagesTotal 是图片总量（无过滤）。
	ImagesTotal int64 `json:"images_total"`
	// StorageBytes 是全部图片 size 字段之和（字节），即「已登记图片总大小」。
	// 取自 DB SUM(size) 而非文件系统遍历，images 表为单一事实来源。
	StorageBytes int64 `json:"storage_bytes"`
	// APIKeysTotal 是密钥总数（含已吊销；已删除为物理删除，不在统计内）。
	APIKeysTotal int `json:"apikeys_total"`
	// APIKeysActive 是未吊销的有效密钥数。
	APIKeysActive int `json:"apikeys_active"`
	// APIKeysRevoked 是已吊销密钥数。
	APIKeysRevoked int `json:"apikeys_revoked"`
	// LogsTotal 是日志总量。
	LogsTotal int64 `json:"logs_total"`
	// RecentUploadTrend 是近 N 天每日新增图片数（按日期升序、缺日补零），
	// 元素结构与日志直方图 buckets 一致，可直接喂给 LogsHistogram 组件。
	RecentUploadTrend []DailyCount `json:"recent_upload_trend"`
	// RecentUploadTotal 是近 N 天新增图片合计，供趋势图标题区高亮展示。
	RecentUploadTotal int `json:"recent_upload_total"`
	// Days 是趋势窗口天数（默认 30），回显给前端用于文案。
	Days int `json:"days"`
}
