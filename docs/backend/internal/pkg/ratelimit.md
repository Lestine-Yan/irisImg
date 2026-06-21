# `internal/pkg/ratelimit/ratelimit.go`

> 文档路径约定：`internal/pkg/ratelimit/ratelimit.go` → `docs/backend/internal/pkg/ratelimit.md`（去掉内层同名目录）。

提供基于**令牌桶**的「按密钥」限流能力。每个密钥（按其数据库 ID 区分）维护一个独立的 `golang.org/x/time/rate` 令牌桶，速率为「每分钟 N 次」，突发容量同样为 N。

底层在内存中维护这些桶，**适用于单实例部署**；多实例场景需替换为共享存储（如 Redis）实现。

## 类型

### `Store`

```go
type Store struct {
    mu               sync.Mutex
    limiters         map[int]*rate.Limiter // 密钥 ID -> 令牌桶
    defaultPerMinute int                   // 未指定独立阈值时的全局默认
}
```

并发安全（`sync.Mutex` 保护 map）。

## 函数

### `NewStore(defaultPerMinute int) *Store`

构造限流存储。`defaultPerMinute <= 0` 时回退为 **100**。全局默认来自配置 `apikey.rate_limit_per_minute`，由 [`router`](../router/router.md) 注入。

### `Allow(keyID, perMinute int) bool`

判断指定密钥此刻是否允许放行，并消耗一个令牌。

- `perMinute <= 0` 时使用全局默认阈值。
- 同一密钥**首次出现**时按其阈值创建令牌桶：产出间隔 = `time.Minute / perMinute`，突发容量 = `perMinute`。
- 已存在的令牌桶**不会因阈值变化而重建**（阈值在密钥生命周期内视为稳定）；改了密钥的 `rate_limit` 需重启进程才完全生效。

## 与其它包的关系

```
middleware.APIKeyAuth ──► ratelimit.Store.Allow(key.ID, key.RateLimit)
```

被 [`middleware.APIKeyAuth`](../middleware/apikey.md) 在鉴权通过后调用；超限返回 429 `CodeTooManyRequests`。

## 注意

- 内存实现：进程重启后所有令牌桶清零；多实例部署各自独立计数，需要全局限流时换 Redis。
- map 只增不减：长期运行、密钥极多时可考虑加 LRU / 过期清理，当前规模无需。
