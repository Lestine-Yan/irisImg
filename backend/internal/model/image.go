package model

import "time"

// Image 是图片元信息的跨层数据载体（实体）。
//
// 它独立于 Ent 生成的 ent.Image：DAO 层负责在二者之间转换，
// 这样 service / api 层不直接依赖 Ent，便于后续替换存储实现。
type Image struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`    // 上传时的原始文件名
	StoredPath string    `json:"stored_path"` // 相对存储根目录的落盘路径
	URL        string    `json:"url"`         // 对外访问地址
	Size       int64     `json:"size"`        // 文件大小（字节）
	MimeType   string    `json:"mime_type"`   // MIME 类型，如 image/png
	Width      int       `json:"width"`       // 宽度（像素），未知为 0
	Height     int       `json:"height"`      // 高度（像素），未知为 0
	Hash       string    `json:"hash"`        // 内容哈希，用于去重
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
	// KeyID 是添加该图片的 API 密钥 ID，可空：
	// 通过后台 JWT 上传的图片没有关联密钥，此处为 nil。
	KeyID *int `json:"key_id,omitempty"`
}
