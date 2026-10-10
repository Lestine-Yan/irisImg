package model

import "time"

// Audio 是音频元信息的跨层数据载体（实体）。
//
// 与 Image 平行的存储字段 + 标准音频元数据（标签 + 技术参数）+ 双层封面字段：
//   - EmbeddedCoverURL：上传时从文件内嵌封面抽取落盘的 URL，此后不变；
//   - CoverURL：用户自定义封面（详情页可改），nil = 未覆盖。
// 生效封面 = CoverURL || EmbeddedCoverURL，由前端计算展示。
//
// 它独立于 Ent 生成的 ent.Audio：DAO 层负责在二者之间转换，
// 这样 service / api 层不直接依赖 Ent，便于后续替换存储实现。
type Audio struct {
	ID         int    `json:"id"`
	Filename   string `json:"filename"`    // 上传时的原始文件名
	StoredPath string `json:"stored_path"` // 相对音频存储根目录的落盘路径
	URL        string `json:"url"`         // 对外访问地址
	Size       int64  `json:"size"`         // 文件大小（字节）
	MimeType   string `json:"mime_type"`   // MIME 类型，如 audio/mpeg
	Hash       string `json:"hash"`        // 内容哈希，用于去重
	Format     string `json:"format"`      // 容器格式：mp3/m4a/flac

	// 标准音频元数据（标签解析），缺失为空串/0。
	Title       string `json:"title"`        // 标题，标签缺失时回退为文件名去扩展名
	Artist      string `json:"artist"`      // 艺术家
	Album       string `json:"album"`        // 专辑
	AlbumArtist string `json:"album_artist"` // 专辑艺术家
	Composer    string `json:"composer"`    // 作曲
	Genre       string `json:"genre"`        // 流派
	Year        int    `json:"year"`         // 发行年份，未知为 0
	TrackNumber int    `json:"track_number"` // 音轨号，未知为 0
	DiscNumber  int    `json:"disc_number"`  // 碟号，未知为 0

	// 技术参数（结构解析），未知为 0。
	DurationSeconds float64 `json:"duration_seconds"` // 时长（秒），未知为 0
	BitrateKbps     int     `json:"bitrate_kbps"`      // 平均码率（kbps）
	SampleRate      int     `json:"sample_rate"`       // 采样率（Hz）
	Channels        int     `json:"channels"`          // 声道数
	BitDepth        int     `json:"bit_depth"`         // 位深，仅 FLAC 有意义

	// 双层封面：生效封面 = CoverURL || EmbeddedCoverURL（nil = 未覆盖）。
	EmbeddedCoverURL string  `json:"embedded_cover_url"` // 内嵌封面 URL（上传时抽取），无则为空
	CoverURL         *string `json:"cover_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"` // 创建时间
	// KeyID 是添加该音频的 API 密钥 ID，可空：
	// 通过后台 JWT 上传的音频没有关联密钥，此处为 nil。
	KeyID *int `json:"key_id,omitempty"`
}

// UploadAudioInput 描述「上传一个音频」的入参，由 api 层从 HTTP 请求装配后传给 service。
//
// 设计成结构体而不是多参数，便于后续追加字段不破坏签名。
type UploadAudioInput struct {
	// Filename 是上传时客户端给出的原始文件名，仅作展示。
	// 真实落盘文件名由 hash + 嗅探得到的格式决定，不信任此字段。
	Filename string
	// Content 是音频完整字节，由 api 层在受 MaxBytesReader 保护下读出。
	Content []byte
	// KeyID 是添加该音频的 API 密钥 ID。
	// 通过 API Key 渠道上传时由中间件保证非空；后台 JWT 直传时可置 nil。
	KeyID *int
}

// AudioListQuery 描述「查询音频列表」的过滤 / 排序 / 分页条件。
//
// 与 ImageListQuery 平行。
type AudioListQuery struct {
	// KeyID 为非 nil 时仅返回该密钥添加的音频；为 nil 表示不按密钥过滤（全部）。
	KeyID *int
	// Order 控制按 created_at 排序的方向：非 "desc" 一律视为升序（"asc"）。
	Order string
	// Offset / Limit 是分页参数；Limit<=0 时由 service 兜底为默认页大小。
	Offset int
	Limit  int
}

// AudioListResult 是音频列表查询的返回结构，同时承载分页元信息。
type AudioListResult struct {
	Items []*Audio `json:"items"`
	Total int      `json:"total"`
}

// BatchDeleteAudiosRequest 是音频中心批量删除的请求体。
// 复用吊销/删除密钥、批量删图同款账号密码二次确认机制：api 层用 AuthService.VerifyCredentials
// 常量时间比对校验，失败返回 403（而非 401），避免触发前端全局登出。
// IDs 上限 100，防止误操作一次删除巨量音频。
type BatchDeleteAudiosRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	// IDs 是待删除的音频主键列表：非空、每项 > 0、上限 100。
	IDs []int `json:"ids" binding:"required,min=1,max=100,dive,gt=0"`
}

// BatchDeleteAudiosResponse 是批量删除音频的响应体。
// 不存在的 ID 静默跳过（幂等语义），两个均按「实际删除」口径返回，
// 供前端同步本地列表状态与分页计算。
type BatchDeleteAudiosResponse struct {
	// Deleted 是实际删除的记录条数（不存在的 ID 不计入）。
	Deleted int `json:"deleted"`
	// IDs 是实际被删除的音频 ID 列表。
	IDs []int `json:"ids"`
}

// UpdateAudioCoverRequest 是详情页「修改封面 URL」的请求体。
// CoverURL 为空串表示清除用户覆盖（恢复展示内嵌封面）；非空须 ≤ 2048 字符。
type UpdateAudioCoverRequest struct {
	CoverURL string `json:"cover_url" binding:"max=2048"`
}
