package entdao

import (
	"context"
	"testing"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// TestLogDAO_ListFilterClientIP 覆盖日志按客户端 IP 前缀过滤。
func TestLogDAO_ListFilterClientIP(t *testing.T) {
	d := newTestLogDAO(t)
	ctx := context.Background()
	now := time.Now()

	for _, ip := range []string{"203.0.113.7", "203.0.113.9", "198.51.100.4", ""} {
		if _, err := d.Create(ctx, &model.Log{
			Timestamp: now,
			Level:     model.LevelInfo,
			Event:     model.EventHTTPRequest,
			ClientIP:  ip,
		}); err != nil {
			t.Fatalf("create %q: %v", ip, err)
		}
	}

	cases := []struct {
		name   string
		prefix string
		want   int
	}{
		{"网段前缀", "203.0.113.", 2},
		{"完整 IP 精确命中", "198.51.100.4", 1},
		{"无匹配", "192.0.2.", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			items, total, err := d.List(ctx, model.LogQuery{ClientIP: c.prefix, Limit: 10})
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if total != c.want || len(items) != c.want {
				t.Fatalf("prefix %q: total=%d len=%d, want %d", c.prefix, total, len(items), c.want)
			}
		})
	}
}
