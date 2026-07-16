package service

import (
	"context"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

const dashboardTrendDays = 30 // 仪表盘近 N 天趋势默认天数

// DashboardService 聚合仪表盘首页所需的各项统计指标。
//
// 依赖 imageDAO / apiKeyDAO / logDAO 三个 DAO（沿用本仓库 service 仅依赖 DAO 的既有模式，
// 不依赖其他 service），在 Overview 中聚合图片总量、存储大小、密钥计数、日志总量
// 与近 N 天上传趋势，一次性返回给仪表盘页面。
type DashboardService struct {
	imageDAO  dao.ImageDAO
	apiKeyDAO dao.APIKeyDAO
	logDAO    dao.LogDAO
}

// NewDashboardService 构造 DashboardService。
func NewDashboardService(imageDAO dao.ImageDAO, apiKeyDAO dao.APIKeyDAO, logDAO dao.LogDAO) *DashboardService {
	return &DashboardService{imageDAO: imageDAO, apiKeyDAO: apiKeyDAO, logDAO: logDAO}
}

// Overview 聚合返回仪表盘首页所需的全部指标。
//
// days 控制近 N 天上传趋势窗口（<=0 兜底为默认 30）。趋势按日循环 imageDAO.CountByRange，
// 照搬 LogService.Histogram 的按日聚合模式（本地时区午夜对齐、缺日补零、升序）。
func (s *DashboardService) Overview(ctx context.Context, days int) (*model.DashboardOverview, error) {
	if days <= 0 {
		days = dashboardTrendDays
	}

	overview := &model.DashboardOverview{Days: days}

	// 图片总量与存储大小（DB SUM(size)）。
	imagesTotal, err := s.imageDAO.Count(ctx)
	if err != nil {
		return nil, err
	}
	overview.ImagesTotal = imagesTotal

	storageBytes, err := s.imageDAO.TotalSize(ctx)
	if err != nil {
		return nil, err
	}
	overview.StorageBytes = storageBytes

	// APIkey 计数：一次拉全量后内存按 Revoked 分桶，避免多次 Count 往返。
	keys, err := s.apiKeyDAO.List(ctx)
	if err != nil {
		return nil, err
	}
	overview.APIKeysTotal = len(keys)
	for _, k := range keys {
		if k.Revoked {
			overview.APIKeysRevoked++
		}
	}
	overview.APIKeysActive = overview.APIKeysTotal - overview.APIKeysRevoked

	// 日志总量。
	logsTotal, err := s.logDAO.Count(ctx)
	if err != nil {
		return nil, err
	}
	overview.LogsTotal = logsTotal

	// 近 N 天每日新增图片趋势。
	trend, trendTotal, err := s.uploadTrend(ctx, days)
	if err != nil {
		return nil, err
	}
	overview.RecentUploadTrend = trend
	overview.RecentUploadTotal = trendTotal

	return overview, nil
}

// uploadTrend 返回最近 days 天的每日新增图片数（按日期升序、缺日补零）与合计。
//
// 与 LogService.Histogram 同构：用 time.Now().Location() 构造今日午夜，逐日左闭右开区间
// 调 imageDAO.CountByRange，缺日因查询返回 0 自然补零。CountByRange 内部已对齐本地时区。
func (s *DashboardService) uploadTrend(ctx context.Context, days int) ([]model.DailyCount, int, error) {
	now := time.Now()
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	out := make([]model.DailyCount, 0, days)
	total := 0
	for i := days - 1; i >= 0; i-- {
		dayStart := today.AddDate(0, 0, -i)
		dayEnd := dayStart.AddDate(0, 0, 1)
		c, err := s.imageDAO.CountByRange(ctx, dayStart, dayEnd)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, model.DailyCount{Date: dayStart.Format("2006-01-02"), Count: int(c)})
		total += int(c)
	}
	return out, total, nil
}
