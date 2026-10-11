package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Audio 是音频文件的元信息实体。
// 真实的音频二进制存放在独立的音频存储目录（storage.audio_root_dir，与图片分目录），
// 数据库只保存检索、展示、去重所需的元数据。
//
// 元数据分两部分：
//   - 标签字段（title/artist/album/...）：上传时经 audiometa 从文件标签解析，
//     缺失为零值（title 为空时由 service 回退为「文件名去扩展名」）；
//   - 技术参数（时长/码率/采样率/声道/位深）：解析失败同样是零值，不阻断上传。
type Audio struct {
	ent.Schema
}

// Fields 定义 Audio 的字段。
func (Audio) Fields() []ent.Field {
	return []ent.Field{
		// 原始文件名，仅用于展示与下载时回填。
		field.String("filename").
			NotEmpty().
			Comment("上传时的原始文件名"),

		// 落盘后的相对路径（相对于音频存储根目录），用于定位文件。
		field.String("stored_path").
			NotEmpty().
			Unique().
			Comment("相对音频存储根目录的落盘路径"),

		// 对外访问 URL（或路径），由上传时生成。
		field.String("url").
			NotEmpty().
			Comment("对外访问地址"),

		// 文件大小，单位字节。
		field.Int64("size").
			NonNegative().
			Comment("文件大小（字节）"),

		// MIME 类型，规范化后的值：audio/mpeg / audio/mp4 / audio/flac。
		field.String("mime_type").
			NotEmpty().
			Comment("MIME 类型，如 audio/mpeg"),

		// 内容哈希（sha256），用于秒传 / 去重。
		field.String("hash").
			NotEmpty().
			Comment("内容哈希，用于去重"),

		// 容器格式：mp3 / m4a / flac，解析失败为空串。
		field.String("format").
			Default("").
			Comment("容器格式：mp3/m4a/flac"),

		// ---- 标准音频元数据（标签解析，零值 = 未知/缺失） ----

		field.String("title").
			Default("").
			Comment("标题，标签缺失时回退为文件名去扩展名"),
		field.String("artist").
			Default("").
			Comment("艺术家"),
		field.String("album").
			Default("").
			Comment("专辑"),
		field.String("album_artist").
			Default("").
			Comment("专辑艺术家"),
		field.String("composer").
			Default("").
			Comment("作曲"),
		field.String("genre").
			Default("").
			Comment("流派"),
		field.Int("year").
			Default(0).
			NonNegative().
			Comment("发行年份，未知为 0"),
		field.Int("track_number").
			Default(0).
			NonNegative().
			Comment("音轨号，未知为 0"),
		field.Int("disc_number").
			Default(0).
			NonNegative().
			Comment("碟号，未知为 0"),

		// ---- 技术参数（结构解析，零值 = 未知） ----

		field.Float("duration_seconds").
			Default(0).
			Comment("时长（秒），未知为 0"),
		field.Int("bitrate_kbps").
			Default(0).
			NonNegative().
			Comment("平均码率（kbps），未知为 0"),
		field.Int("sample_rate").
			Default(0).
			NonNegative().
			Comment("采样率（Hz），未知为 0"),
		field.Int("channels").
			Default(0).
			NonNegative().
			Comment("声道数，未知为 0"),
		field.Int("bit_depth").
			Default(0).
			NonNegative().
			Comment("位深，仅 FLAC 有意义，其余为 0"),

		// ---- 封面 ----

		// 内嵌封面抽取落盘后的 URL（data/audios/covers/<sha256>.<ext>，内容寻址共享），
		// 上传时写入，此后不再变化；无内嵌封面为空串。
		field.String("embedded_cover_url").
			Default("").
			Comment("内嵌封面 URL（上传时抽取），无则为空"),

		// 用户自定义封面 URL（详情页可改）。nil = 未覆盖（展示 embedded_cover_url），
		// 空串清除后回到 nil 语义由 service 转换。生效封面 = cover_url || embedded_cover_url。
		field.String("cover_url").
			Optional().
			Nillable().
			Comment("用户覆盖的封面 URL，nil 表示用内嵌封面"),

		// 创建时间，落库时自动写入。
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),

		// 添加该音频的 API 密钥 ID（可空）。绑定到 key edge，便于直接读取来源密钥。
		field.Int("key_id").
			Optional().
			Nillable().
			Comment("添加该音频的 API 密钥 ID，后台 JWT 上传时为空"),
	}
}

// Edges 定义关联：音频可选地归属于添加它的 API 密钥。
// 通过后台 JWT 上传的音频没有关联密钥，故该 edge 可空。
func (Audio) Edges() []ent.Edge {
	return []ent.Edge{
		// 添加该音频的 API 密钥（可空）。
		edge.From("key", ApiKey.Type).
			Ref("audios").
			Field("key_id").
			Unique(),
	}
}

// Indexes 为高频查询字段建立索引。
func (Audio) Indexes() []ent.Index {
	return []ent.Index{
		// 按哈希去重 / 秒传查询。
		index.Fields("hash"),
		// 按创建时间分页列表。
		index.Fields("created_at"),
	}
}
