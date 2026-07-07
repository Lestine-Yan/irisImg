package model

import "time"

// 密钥权限范围常量。
const (
	// ScopeReadOnly 只读密钥：仅能访问 GET 接口（申请图片）。
	ScopeReadOnly = "readonly"
	// ScopeReadWrite 读写密钥：可访问 GET 及 POST 接口（添加图片）。
	ScopeReadWrite = "readwrite"
)

// APIKey 是 API 密钥的跨层数据载体（实体）。
//
// 独立于 Ent 生成的 ent.ApiKey：DAO 层负责二者转换，使 service / api 层不直接依赖 Ent。
// KeyHash 仅在 DAO / service 内部流转，序列化为 JSON 时始终忽略，避免泄露。
type APIKey struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`         // 密钥标签
	Prefix     string     `json:"prefix"`       // 明文前缀，用于展示识别
	Scope      string     `json:"scope"`        // readonly / readwrite
	KeyHash    string     `json:"-"`            // SHA-256 哈希，不对外暴露
	RateLimit  int        `json:"rate_limit"`   // 限流阈值（次/分钟），0 表示用全局默认
	Revoked    bool       `json:"revoked"`      // 是否已吊销
	LastUsedAt *time.Time `json:"last_used_at"` // 最近使用时间，未使用过为 nil
	CreatedAt  time.Time  `json:"created_at"`   // 创建时间
}

// CreateAPIKeyRequest 是创建密钥的请求体。
type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required"`                           // 密钥标签
	Scope     string `json:"scope" binding:"required,oneof=readonly readwrite"` // 权限范围
	RateLimit int    `json:"rate_limit"`                                        // 可选，限流阈值，0 表示用全局默认
}

// CreateAPIKeyResponse 是创建密钥的响应体。明文 Key 仅在此返回一次。
type CreateAPIKeyResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	Scope     string    `json:"scope"`
	Key       string    `json:"key"` // 明文密钥，仅创建时返回一次，请妥善保存
	RateLimit int       `json:"rate_limit"`
	CreatedAt time.Time `json:"created_at"`
}

// APIKeyInfo 是密钥列表项，不包含明文与哈希。
type APIKeyInfo struct {
	ID         int        `json:"id"`
	Name       string    `json:"name"`
	Prefix     string    `json:"prefix"`
	Scope      string    `json:"scope"`
	RateLimit  int        `json:"rate_limit"`
	Revoked    bool       `json:"revoked"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// RenameAPIKeyRequest 是重命名密钥的请求体。
type RenameAPIKeyRequest struct {
	Name string `json:"name" binding:"required,max=64"` // 新的密钥标签
}

// ResetAPIKeyResponse 是重置密钥明文后的响应体，与创建响应同构。
// 新明文 Key 仅在此返回一次，调用方需自行妥善保存；重置同时会取消吊销状态。
type ResetAPIKeyResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	Key       string    `json:"key"`       // 新明文密钥，仅此一次返回
	Revoked   bool      `json:"revoked"`   // 重置后恒为 false
	CreatedAt time.Time `json:"created_at"`
}

// DestructiveAPIKeyRequest 是吊销 / 删除密钥这类敏感操作的请求体。
// 后端用 subtle.ConstantTimeCompare 校验账号密码，作为 JWT 登录态之上的二次确认。
type DestructiveAPIKeyRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// DeleteAPIKeyResponse 是删除密钥的响应体，附带被级联删除的图片数量。
type DeleteAPIKeyResponse struct {
	ID            int  `json:"id"`
	Deleted       bool `json:"deleted"`
	ImagesRemoved int  `json:"images_removed"`
}
