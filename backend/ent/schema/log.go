package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	entschema "entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AccessLog 是日志中心的统一日志实体，既记录每条 HTTP 请求（event=http.request），
// 也记录关键业务事件（event=image.upload / apikey.revoke / auth.login_success / log.clear 等），
// 还记录 panic（event=panic）。
//
// ent 预声明了 Log 标识符，故 Go 侧 schema 类型用 AccessLog；实际表名经 entsql 注解定为 logs。
// 访问日志由中间件异步批量落库，供日志中心查询 / 直方图 / 清理。
// api_key_id 不建 Edge，避免高频插入的外键开销；仅以普通字段记录来源密钥 ID。
type AccessLog struct {
	ent.Schema
}

// Annotations 把表名固定为 logs（避免默认的 access_logs）。
func (AccessLog) Annotations() []entschema.Annotation {
	return []entschema.Annotation{
		entsql.Annotation{Table: "logs"},
	}
}

// Fields 定义 AccessLog 的字段。
func (AccessLog) Fields() []ent.Field {
	return []ent.Field{
		// 日志发生时间，用于排序 / 直方图 / 时间范围过滤。落库时自动写入，不可变。
		field.Time("timestamp").
			Default(time.Now).
			Immutable().
			Comment("日志发生时间，用于排序 / 直方图 / 时间范围过滤"),

		// 日志级别，按 HTTP 状态推导（≥500 error / ≥400 warn / 否则 info）或业务事件显式指定。
		field.Enum("level").
			Values("debug", "info", "warn", "error").
			Default("info").
			Comment("日志级别"),

		// 事件类型：http.request / image.upload / apikey.* / auth.* / log.clear / panic。
		field.String("event").
			NotEmpty().
			Comment("事件类型"),

		// HTTP 方法，仅访问日志有值。
		field.String("method").
			Optional().
			Comment("HTTP 方法，仅访问日志有值"),

		// 请求路径，用于关键字模糊匹配。
		field.String("path").
			Optional().
			Comment("请求路径，用于关键字模糊匹配"),

		// HTTP 状态码，非请求类事件为空（NULL）。
		field.Int("status").
			Optional().
			Nillable().
			Comment("HTTP 状态码，非请求类事件为空"),

		// 请求耗时（毫秒），仅访问日志有值。
		field.Int("duration_ms").
			Optional().
			Nillable().
			Comment("请求耗时（毫秒），仅访问日志有值"),

		// 客户端 IP。
		field.String("client_ip").
			Optional().
			Comment("客户端 IP"),

		// 请求追踪 ID，关联同一请求的访问日志与业务事件。
		field.String("request_id").
			Optional().
			Comment("请求追踪 ID，关联同一请求的访问日志与业务事件"),

		// 来源 API 密钥 ID（无关联时为空），不建 Edge。
		field.Int("api_key_id").
			Optional().
			Nillable().
			Comment("来源 API 密钥 ID，无关联时为空"),

		// 操作者用户名（JWT 登录用户）。
		field.String("username").
			Optional().
			Comment("操作者用户名（JWT 登录用户）"),

		// 可读描述 / 错误信息。
		field.String("message").
			Default("").
			Comment("可读描述 / 错误信息"),

		// 记录落库时间。
		field.Time("created_at").
			Default(time.Now).
			Comment("记录落库时间"),
	}
}

// Edges 不声明关联：日志高频写入，避免外键约束开销。
func (AccessLog) Edges() []ent.Edge {
	return nil
}

// Indexes 为高频查询字段建立索引。
func (AccessLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("timestamp"),
		index.Fields("level"),
		index.Fields("event"),
		index.Fields("request_id"),
		index.Fields("api_key_id"),
	}
}
