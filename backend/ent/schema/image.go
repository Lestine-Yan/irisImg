package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Image 是图床的图片元信息实体。
// 真实的图片二进制存放在本地 data/ 目录（或后续的对象存储），
// 数据库只保存检索、展示、去重所需的元数据。
type Image struct {
	ent.Schema
}

// Fields 定义 Image 的字段。
func (Image) Fields() []ent.Field {
	return []ent.Field{
		// 原始文件名，仅用于展示与下载时回填。
		field.String("filename").
			NotEmpty().
			Comment("上传时的原始文件名"),

		// 落盘后的相对路径（相对于存储根目录），用于定位文件。
		field.String("stored_path").
			NotEmpty().
			Unique().
			Comment("相对存储根目录的落盘路径"),

		// 对外访问 URL（或路径），由上传时生成。
		field.String("url").
			NotEmpty().
			Comment("对外访问地址"),

		// 文件大小，单位字节。
		field.Int64("size").
			NonNegative().
			Comment("文件大小（字节）"),

		// MIME 类型，例如 image/png。
		field.String("mime_type").
			NotEmpty().
			Comment("MIME 类型，如 image/png"),

		// 图片宽高，无法解析时为 0。
		field.Int("width").
			Default(0).
			NonNegative().
			Comment("宽度（像素），未知为 0"),
		field.Int("height").
			Default(0).
			NonNegative().
			Comment("高度（像素），未知为 0"),

		// 内容哈希（如 sha256），用于秒传 / 去重。
		field.String("hash").
			NotEmpty().
			Comment("内容哈希，用于去重"),

		// 创建时间，落库时自动写入。
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),

		// 添加该图片的 API 密钥 ID（可空）。绑定到 key edge，便于直接读取来源密钥。
		field.Int("key_id").
			Optional().
			Nillable().
			Comment("添加该图片的 API 密钥 ID，后台 JWT 上传时为空"),
	}
}

// Edges 定义关联：图片可选地归属于添加它的 API 密钥。
// 通过后台 JWT 上传的图片没有关联密钥，故该 edge 可空。
func (Image) Edges() []ent.Edge {
	return []ent.Edge{
		// 添加该图片的 API 密钥（可空）。
		edge.From("key", ApiKey.Type).
			Ref("images").
			Field("key_id").
			Unique(),
	}
}

// Indexes 为高频查询字段建立索引。
func (Image) Indexes() []ent.Index {
	return []ent.Index{
		// 按哈希去重 / 秒传查询。
		index.Fields("hash"),
		// 按创建时间倒序分页列表。
		index.Fields("created_at"),
	}
}
