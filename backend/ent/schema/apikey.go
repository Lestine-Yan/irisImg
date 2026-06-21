package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApiKey 是用于「申请图片 / 添加图片」鉴权的 API 密钥实体。
// 明文密钥由 32 字节随机数经 base64(URL-safe, 无填充) 生成，仅在创建时返回一次；
// 数据库只保存其 SHA-256 哈希，无法反推明文。
// 密钥区分 readonly（只读，仅可访问 GET 接口）与 readwrite（读写，可 POST 添加图片）两种权限。
type ApiKey struct {
	ent.Schema
}

// Fields 定义 ApiKey 的字段。
func (ApiKey) Fields() []ent.Field {
	return []ent.Field{
		// 人类可读的标签，便于在管理界面区分不同密钥的用途。
		field.String("name").
			NotEmpty().
			Comment("密钥标签，便于识别用途"),

		// 明文密钥的 SHA-256 哈希（十六进制），唯一。校验时按此列查找。
		field.String("key_hash").
			NotEmpty().
			Unique().
			Comment("明文密钥的 SHA-256 哈希，唯一"),

		// 明文密钥前若干位，仅用于展示识别（不足以反推明文）。
		field.String("prefix").
			NotEmpty().
			Comment("明文密钥前缀，用于展示识别"),

		// 权限范围：readonly 只读 / readwrite 读写。
		field.Enum("scope").
			Values("readonly", "readwrite").
			Comment("权限范围：readonly 只读 / readwrite 读写"),

		// 限流阈值（次/分钟），0 表示沿用全局默认配置。
		field.Int("rate_limit").
			Default(0).
			NonNegative().
			Comment("限流阈值（次/分钟），0 表示使用全局默认"),

		// 是否已吊销。吊销后密钥立即失效。
		field.Bool("revoked").
			Default(false).
			Comment("是否已吊销"),

		// 最近一次使用时间，未使用过时为空。
		field.Time("last_used_at").
			Optional().
			Nillable().
			Comment("最近一次使用时间"),

		// 创建时间，落库时自动写入。
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
	}
}

// Edges 定义关联：一个密钥可以添加多张图片。
func (ApiKey) Edges() []ent.Edge {
	return []ent.Edge{
		// 该密钥添加的所有图片。
		edge.To("images", Image.Type),
	}
}

// Indexes 为高频查询字段建立索引。
func (ApiKey) Indexes() []ent.Index {
	return []ent.Index{
		// 校验时按哈希精确查找。
		index.Fields("key_hash").Unique(),
	}
}
