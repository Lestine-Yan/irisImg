package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Lestine-Yan/irisImg/backend/internal/middleware"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// AudioAPI 是音频相关接口的控制器。与 ImageAPI 平行。
//
// 当前已落地：
//   - POST /audios（对外上传音频）：由 API Key 鉴权中间件保护，只读密钥访问 POST 会被 403。
//   - POST /api/v1/admin/audios（后台直传）：由 JWT 鉴权中间件保护，供音频中心上传，key_id 留空。
//   - GET /api/v1/admin/audios（后台列表）：由 JWT 鉴权中间件保护，供音频中心拉取音频。
//   - DELETE /api/v1/admin/audios（后台批量删除）：由 JWT 鉴权中间件保护，供音频中心批量删除，
//     需请求体携带账号密码做二次确认。
//   - PATCH /api/v1/admin/audios/:id（改封面）：更新用户覆盖的封面 URL，空串恢复内嵌默认。
//
// GET /audios（对外，API Key）暂为占位，与 GET /images 的占位语义一致。
type AudioAPI struct {
	svc     *service.AudioService
	authSvc *service.AuthService
	rec     service.LogRecorder
}

// NewAudioAPI 通过依赖注入构造控制器。authSvc 用于批量删除的账号密码二次确认；
// rec 用于记录音频上传 / 删除 / 改封面业务事件到日志中心。
func NewAudioAPI(svc *service.AudioService, authSvc *service.AuthService, rec service.LogRecorder) *AudioAPI {
	return &AudioAPI{svc: svc, authSvc: authSvc, rec: rec}
}

// recordUpload 记录一次音频上传业务事件。rec 为空时直接返回，便于测试。
func (h *AudioAPI) recordUpload(c *gin.Context, filename string) {
	if h.rec == nil {
		return
	}
	h.rec.Record(model.NewEventLog(model.EventAudioUpload, model.LevelInfo, "upload audio: "+filename, middleware.LogContextFromGin(c)))
}

// recordDelete 记录一次批量删除音频业务事件（破坏性操作，级别 warn）。rec 为空时直接返回，便于测试。
func (h *AudioAPI) recordDelete(c *gin.Context, count int) {
	if h.rec == nil {
		return
	}
	h.rec.Record(model.NewEventLog(model.EventAudioDelete, model.LevelWarn,
		fmt.Sprintf("batch delete audios: %d", count), middleware.LogContextFromGin(c)))
}

// recordUpdateCover 记录一次修改音频封面业务事件。rec 为空时直接返回，便于测试。
func (h *AudioAPI) recordUpdateCover(c *gin.Context, id int, cleared bool) {
	if h.rec == nil {
		return
	}
	action := "update"
	if cleared {
		action = "reset"
	}
	h.rec.Record(model.NewEventLog(model.EventAudioUpdateCover, model.LevelInfo,
		fmt.Sprintf("%s audio cover: #%d", action, id), middleware.LogContextFromGin(c)))
}

// List 处理 GET /audios（对外占位）。任意有效密钥均可访问。
// 对外列表语义与图片一致待定，当前显式返回 501 与鉴权失败状态码区分开。
func (h *AudioAPI) List(c *gin.Context) {
	response.Fail(c, http.StatusNotImplemented, response.CodeServerError, "音频列表接口尚未实现（占位）")
}

// ListAdmin 处理 GET /api/v1/admin/audios（JWT 保护），供后台音频中心拉取音频列表。
//
//	Query 参数:
//	  - key_id    可选；缺省=不按密钥过滤（全部），>=1 时只返回该密钥添加的音频。
//	  - order     可选；asc/desc，默认 asc（时间升序，与图片列表一致）。
//	  - page      可选；默认 1，<1 视为非法。
//	  - page_size 可选；默认 24，<1 视为非法。
//
//	响应: { items: [model.Audio], total, page, page_size }
func (h *AudioAPI) ListAdmin(c *gin.Context) {
	var keyID *int
	if raw := c.Query("key_id"); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil || id < 1 {
			response.BadRequest(c, "无效的 key_id")
			return
		}
		keyID = &id
	}

	order := c.DefaultQuery("order", "asc")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		response.BadRequest(c, "无效的 page")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "24"))
	if err != nil || pageSize < 1 {
		response.BadRequest(c, "无效的 page_size")
		return
	}

	q := model.AudioListQuery{
		KeyID:  keyID,
		Order:  order,
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	}
	result, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.ServerError(c, "查询音频列表失败："+err.Error())
		return
	}
	response.Success(c, gin.H{
		"items":     result.Items,
		"total":     result.Total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Create 处理 POST /audios，落地一个音频：
//
//	请求：multipart/form-data，字段名 "file"，需 readwrite 密钥（由中间件保证）。
//	接受 mp3 / m4a / flac（按内容魔数嗅探，不信任扩展名）。
//	响应：200 + 完整的 model.Audio（含元数据、双层封面、key_id 等）。
//
// 错误：
//   - 400 缺少文件 / 内容为空 / 格式不在白名单（含视频 mp4）
//   - 413 超过 storage.audio_max_upload_size_mb
//   - 500 其它内部错误
func (h *AudioAPI) Create(c *gin.Context) {
	filename, content, ok := readUploadFile(c, h.svc.MaxBytes())
	if !ok {
		return
	}

	// 中间件保证此处 key_id 必然存在且 >0。
	keyID := c.GetInt(middleware.ContextKeyAPIKeyID)

	a, err := h.svc.Upload(c.Request.Context(), &model.UploadAudioInput{
		Filename: filename,
		Content:  content,
		KeyID:    &keyID,
	})
	if err != nil {
		respondUploadErrorWith(c, err, "音频")
		return
	}
	h.recordUpload(c, filename)
	response.Success(c, a)
}

// CreateAdmin 处理 POST /api/v1/admin/audios，后台 JWT 直传一个音频：
//
//	请求：multipart/form-data，字段名 "file"，由 JWT 鉴权中间件保护（无需 X-API-Key）。
//	响应：200 + 完整的 model.Audio（key_id 为 nil，代表 admin 直传）。
//
// 与 Create 的差别仅在不走 API Key 通道、KeyID 传 nil：上传的音频不关联任何密钥，
// 只会在音频中心「全部」里出现，详情里来源展示为 admin。
func (h *AudioAPI) CreateAdmin(c *gin.Context) {
	filename, content, ok := readUploadFile(c, h.svc.MaxBytes())
	if !ok {
		return
	}

	a, err := h.svc.Upload(c.Request.Context(), &model.UploadAudioInput{
		Filename: filename,
		Content:  content,
		KeyID:    nil,
	})
	if err != nil {
		respondUploadErrorWith(c, err, "音频")
		return
	}
	h.recordUpload(c, filename)
	response.Success(c, a)
}

// BatchDeleteAdmin 处理 DELETE /api/v1/admin/audios，批量删除音频（物理文件 + 元信息记录）。
// 供音频中心多选删除使用。内嵌封面（内容寻址共享文件）刻意不清理，见 service 注释。
//
//	请求：JSON body，含账号密码（二次确认）与待删除 ID 列表（非空、每项 >0、上限 100）。
//	响应：200 + { deleted: 实际删除条数, ids: 实际被删除的音频 ID 列表 }。
//
// 与批量删图同款二次确认：VerifyCredentials 失败返回 403（而非 401），避免触发前端全局登出。
// 不存在的 ID 静默跳过（幂等语义），不构成错误。
func (h *AudioAPI) BatchDeleteAdmin(c *gin.Context) {
	var req model.BatchDeleteAudiosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.authSvc.VerifyCredentials(req.Username, req.Password); err != nil {
		response.Forbidden(c, "用户名或密码错误")
		return
	}

	deleted, ids, err := h.svc.BatchDelete(c.Request.Context(), req.IDs)
	if err != nil {
		response.ServerError(c, "批量删除音频失败："+err.Error())
		return
	}
	h.recordDelete(c, deleted)
	response.Success(c, model.BatchDeleteAudiosResponse{Deleted: deleted, IDs: ids})
}

// UpdateCoverAdmin 处理 PATCH /api/v1/admin/audios/:id，修改音频封面：
//
//	请求：JSON body { "cover_url": "<URL 或空串>" }；空串 = 清除用户覆盖，恢复展示内嵌封面。
//	      URL 长度上限 2048（binding 校验）。
//	响应：200 + 更新后的完整 model.Audio（前端据此就地替换列表项）。
//
// 错误：
//   - 400 id 非法 / body 解析失败 / cover_url 超 2048 字符
//   - 404 音频不存在
//   - 500 其它内部错误
func (h *AudioAPI) UpdateCoverAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		response.BadRequest(c, "无效的音频 ID")
		return
	}

	var req model.UpdateAudioCoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	a, err := h.svc.UpdateCover(c.Request.Context(), id, req.CoverURL)
	if err != nil {
		if errors.Is(err, service.ErrAudioNotFound) {
			response.NotFound(c, "音频不存在")
			return
		}
		response.ServerError(c, "修改音频封面失败："+err.Error())
		return
	}
	h.recordUpdateCover(c, id, req.CoverURL == "")
	response.Success(c, a)
}
