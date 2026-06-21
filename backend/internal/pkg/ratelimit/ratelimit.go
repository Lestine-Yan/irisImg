// Package ratelimit 提供基于令牌桶的「按密钥」限流能力。
//
// 每个密钥（按其数据库 ID 区分）维护一个独立的 golang.org/x/time/rate 令牌桶，
// 速率为「每分钟 N 次」，突发容量同样为 N。Store 在内存中维护这些桶，
// 适用于单实例部署；多实例场景需替换为共享存储（如 Redis）实现。
package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Store 维护「密钥 ID -> 令牌桶」的映射，并发安全。
type Store struct {
	mu       sync.Mutex
	limiters map[int]*rate.Limiter
	// defaultPerMinute 是未指定独立阈值时使用的全局默认（次/分钟）。
	defaultPerMinute int
}

// NewStore 构造限流存储。defaultPerMinute <= 0 时回退为 100。
func NewStore(defaultPerMinute int) *Store {
	if defaultPerMinute <= 0 {
		defaultPerMinute = 100
	}
	return &Store{
		limiters:         make(map[int]*rate.Limiter),
		defaultPerMinute: defaultPerMinute,
	}
}

// Allow 判断指定密钥此刻是否允许放行，并消耗一个令牌。
// perMinute <= 0 时使用全局默认阈值。同一密钥首次出现时按其阈值创建令牌桶；
// 已存在的令牌桶不会因阈值变化而重建（阈值在密钥生命周期内视为稳定）。
func (s *Store) Allow(keyID, perMinute int) bool {
	if perMinute <= 0 {
		perMinute = s.defaultPerMinute
	}

	s.mu.Lock()
	limiter, ok := s.limiters[keyID]
	if !ok {
		// 每分钟 perMinute 次：令牌产出间隔 = 一分钟 / perMinute；突发容量 = perMinute。
		limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(perMinute)), perMinute)
		s.limiters[keyID] = limiter
	}
	s.mu.Unlock()

	return limiter.Allow()
}
