package service

import (
	"context"
	"errors"
	"testing"
)

// TestAPIKeyService_TouchDebounced 覆盖 last_used_at 去抖：
// 窗口内多次调用只写一次；窗口归零后每次都写；写库失败会清除去抖记录以便下次重试。
func TestAPIKeyService_TouchDebounced(t *testing.T) {
	m := newMockAPIKeyDAO()
	svc := NewAPIKeyService(m, nil, nil, nil, nil)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if err := svc.TouchDebounced(ctx, 1); err != nil {
			t.Fatalf("第 %d 次 TouchDebounced 失败：%v", i+1, err)
		}
	}
	if m.touchCount != 1 {
		t.Fatalf("去抖窗口内应只写 1 次，实际 %d 次", m.touchCount)
	}

	// 窗口归零后每次调用都应写库。
	svc.touchInterval = 0
	if err := svc.TouchDebounced(ctx, 1); err != nil {
		t.Fatalf("窗口归零后调用失败：%v", err)
	}
	if m.touchCount != 2 {
		t.Fatalf("窗口归零后应写第 2 次，实际 %d 次", m.touchCount)
	}

	// 写库失败不应把失败状态固化：下一次调用应再次尝试。
	m.touchErr = errors.New("db down")
	if err := svc.TouchDebounced(ctx, 2); err == nil {
		t.Fatalf("DAO 报错时应返回错误")
	}
	m.touchErr = nil
	if err := svc.TouchDebounced(ctx, 2); err != nil {
		t.Fatalf("DAO 恢复后应立即重试成功，实际 %v", err)
	}
	if m.touchCount != 4 {
		t.Fatalf("失败调用与重试应各写 1 次，累计期望 4，实际 %d", m.touchCount)
	}
}
