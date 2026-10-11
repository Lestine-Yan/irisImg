// Package audiometa 提供音频文件的标准元数据解析（标签 + 技术参数 + 内嵌封面）。
//
// 设计原则：
//   - 「永不因解析失败阻断上传」：Parse 对任何输入都返回 Info（零值降级），
//     元数据缺失只是展示层信息不全，不是上传错误。格式合法性由 Sniff + 白名单把关。
//   - 四个职责分离的纯 Go 依赖集中在本包内消化，service 层保持薄：
//       github.com/dhowden/tag      —— mp3(ID3)/m4a(MP4 atom)/flac(Vorbis comment) 标签 + 封面
//       github.com/tcolgate/mp3    —— mp3 帧遍历（时长/采样率/声道）
//       github.com/go-flac/go-flac —— FLAC STREAMINFO（时长/采样率/声道/位深）
//       github.com/abema/go-mp4    —— MP4 盒子结构（mvhd 时长、hdlr 音轨校验）
package audiometa

import (
	"bytes"
	"math"
	"strings"

	mp4 "github.com/abema/go-mp4"
	"github.com/dhowden/tag"
	flac "github.com/go-flac/go-flac"
	mp3 "github.com/tcolgate/mp3"
)

// Format 是嗅探出的音频容器格式。
type Format string

// 支持的音频格式（值即小写文件扩展名，便于落盘时直接使用）。
const (
	FormatUnknown Format = ""
	FormatMP3     Format = "mp3"
	FormatM4A     Format = "m4a"
	FormatFLAC    Format = "flac"
)

// Info 汇总一份音频文件的标准元数据。
//
// 标签字段（Title/Artist/...）与封面来自标签解析（ID3 / MP4 atom / Vorbis comment）；
// 技术参数来自各格式的结构解析。任何子解析失败对应字段保持零值，绝不阻断上传。
type Info struct {
	// 标签元数据，缺失为空串/0。
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Composer    string
	Genre       string
	Year        int
	Track       int // 音轨号（总轨数不入库）
	Disc        int // 碟号（总碟数不入库）

	// 技术参数，未知为 0。
	DurationSeconds float64 // 秒；0 = 未知
	SampleRate       int     // Hz
	Channels         int     // 声道数
	BitDepth         int     // 位深；仅 FLAC 有意义（无损位深），有损格式恒 0
	BitrateKBPS      int     // 平均码率 = 文件大小*8/时长/1000，向上取整；时长未知为 0

	// 内嵌封面（ID3 APIC 帧 / MP4 covr / FLAC PICTURE 块），nil = 无。
	PictureData []byte
	// PictureExt 是封面图片扩展名（不含点号），空 = 无法识别（调用方应跳过落盘）。
	PictureExt string
}

// Parse 解析给定格式的音频内容，永不返回 error：子解析器失败时对应字段零值降级。
// size 是文件字节数（用于计算平均码率）。f 为 FormatUnknown 时返回零值 Info。
func Parse(content []byte, f Format, size int64) Info {
	switch f {
	case FormatMP3:
		return parseMP3(content, size)
	case FormatM4A:
		return parseM4A(content, size)
	case FormatFLAC:
		return parseFLAC(content, size)
	default:
		return Info{}
	}
}

// HasAudioTrack 对 MP4 家族容器做「真音频」结构校验：存在 soun 音轨且无 vide 视轨。
//
// 背景：ffmpeg 默认把 .m4a 的 ftyp major_brand 编成 isom，无法靠 brand 区分音频/视频，
// 只有走到 moov.trak.mdia.hdlr 看 HandlerType 才可靠。mp3/flac 的魔数本身就是格式的
// 完整定义，无需此校验。结构损坏（盒子读不通）也返回 false——交由上层拒收。
func HasAudioTrack(content []byte) bool {
	var hasSound, hasVideo bool
	_, err := mp4.ReadBoxStructure(bytes.NewReader(content), func(h *mp4.ReadHandle) (interface{}, error) {
		switch {
		case h.BoxInfo.Type.MatchWith(mp4.BoxTypeHdlr()):
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			if hdlr, ok := box.(*mp4.Hdlr); ok {
				switch string(hdlr.HandlerType[:]) {
				case "soun":
					hasSound = true
				case "vide":
					hasVideo = true
				}
			}
			return nil, nil
		// 逐层下钻到 moov.trak.mdia.hdlr；其余盒子（ftyp/free/mdat...）跳过。
		case h.BoxInfo.Type.MatchWith(mp4.BoxTypeMoov()),
			h.BoxInfo.Type.MatchWith(mp4.BoxTypeTrak()),
			h.BoxInfo.Type.MatchWith(mp4.BoxTypeMdia()):
			return h.Expand()
		}
		return nil, nil
	})
	if err != nil {
		return false
	}
	return hasSound && !hasVideo
}

// parseMP3 解析 mp3：ID3 标签 + 帧遍历技术参数。
// 时长按「累计每帧样本数 / 采样率」精确计算（VBR 也准确），
// 比逐帧 Duration() 相加更稳（后者逐帧截断到毫秒以下，长文件会有累积误差）。
func parseMP3(content []byte, size int64) Info {
	var info Info
	if m, err := tag.ReadFrom(bytes.NewReader(content)); err == nil {
		fillFromTags(&info, m)
	}

	d := mp3.NewDecoder(bytes.NewReader(content))
	var frame mp3.Frame
	skipped := 0
	var totalSamples int64
	sampleRate, channels := 0, 0
	for {
		if err := d.Decode(&frame, &skipped); err != nil {
			// io.EOF 或中途损坏：已累计的部分照样有效，结束遍历。
			break
		}
		if sampleRate == 0 {
			// 首个合法帧确定整文件的采样率与声道数（正常文件所有帧一致）。
			if sr := int(frame.Header().SampleRate()); sr > 0 {
				sampleRate = sr
			}
			if frame.Header().ChannelMode() == mp3.SingleChannel {
				channels = 1
			} else {
				channels = 2
			}
		}
		totalSamples += int64(frame.Samples())
	}
	if sampleRate > 0 && totalSamples > 0 {
		info.DurationSeconds = float64(totalSamples) / float64(sampleRate)
	}
	info.SampleRate = sampleRate
	info.Channels = channels
	info.BitrateKBPS = avgBitrateKbps(size, info.DurationSeconds)
	return info
}

// parseFLAC 解析 FLAC：Vorbis comment 标签 + STREAMINFO 技术参数。
// flac.ParseMetadata 只读元数据块不读帧数据，对大文件也只消耗头部几百字节。
func parseFLAC(content []byte, size int64) Info {
	var info Info
	if m, err := tag.ReadFrom(bytes.NewReader(content)); err == nil {
		fillFromTags(&info, m)
	}
	if f, err := flac.ParseMetadata(bytes.NewReader(content)); err == nil {
		if si, err := f.GetStreamInfo(); err == nil && si != nil {
			// go-flac 的 StreamInfoBlock 已把声道数/位深从「存储值-1」归一化为真实值。
			info.SampleRate = si.SampleRate
			info.Channels = si.ChannelCount
			info.BitDepth = si.BitDepth
			if si.SampleRate > 0 && si.SampleCount > 0 {
				info.DurationSeconds = float64(si.SampleCount) / float64(si.SampleRate)
			}
		}
	}
	info.BitrateKBPS = avgBitrateKbps(size, info.DurationSeconds)
	return info
}

// parseM4A 解析 m4a：MP4 atom 标签 + mvhd 时长。
// moov 盒在文件头（faststart）或尾都能被盒子遍历命中。
func parseM4A(content []byte, size int64) Info {
	var info Info
	if m, err := tag.ReadFrom(bytes.NewReader(content)); err == nil {
		fillFromTags(&info, m)
	}

	var duration float64
	_, err := mp4.ReadBoxStructure(bytes.NewReader(content), func(h *mp4.ReadHandle) (interface{}, error) {
		switch {
		case h.BoxInfo.Type.MatchWith(mp4.BoxTypeMvhd()):
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			mvhd, ok := box.(*mp4.Mvhd)
			if !ok || mvhd.Timescale == 0 {
				return nil, nil
			}
			var d uint64
			if mvhd.GetVersion() == 1 {
				d = mvhd.DurationV1
			} else {
				d = uint64(mvhd.DurationV0)
			}
			// 全 1 / 全 F 是「时长未知」的哨兵值。
			if d != 0 && d != math.MaxUint32 && d != math.MaxUint64 {
				duration = float64(d) / float64(mvhd.Timescale)
			}
			return nil, nil
		case h.BoxInfo.Type.MatchWith(mp4.BoxTypeMoov()):
			return h.Expand()
		}
		return nil, nil
	})
	if err == nil {
		info.DurationSeconds = duration
	}
	// m4a 的采样率/声道在 stsd 里且随编码器差异大，v1 不解析（保持 0），时长与码率足够展示。
	info.BitrateKBPS = avgBitrateKbps(size, info.DurationSeconds)
	return info
}

// fillFromTags 把 dhowden/tag 解析出的标签字段与封面填进 Info。
func fillFromTags(info *Info, m tag.Metadata) {
	info.Title = m.Title()
	info.Artist = m.Artist()
	info.Album = m.Album()
	info.AlbumArtist = m.AlbumArtist()
	info.Composer = m.Composer()
	info.Genre = m.Genre()
	info.Year = m.Year()
	track, _ := m.Track()
	disc, _ := m.Disc()
	info.Track = track
	info.Disc = disc
	if pic := m.Picture(); pic != nil && len(pic.Data) > 0 {
		info.PictureData = pic.Data
		info.PictureExt = pictureExt(pic)
	}
}

// pictureExt 推断封面图片扩展名（不含点号）：
// 优先标签自带的 Ext（MP4 covr 常见），其次按 MIME 映射，最后按字节魔数嗅探
// （Go 的 http.DetectContentType 不识别 WebP，故手工判定）。
// 返回空串表示无法识别——静态扩展名白名单服务不了未知类型，调用方应跳过落盘。
func pictureExt(pic *tag.Picture) string {
	switch ext := lowerExt(pic.Ext); ext {
	case "jpg", "jpeg":
		return "jpg"
	case "png", "webp", "gif", "bmp":
		return ext
	}
	switch pic.MIMEType {
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "image/bmp":
		return "bmp"
	}
	return sniffImageExt(pic.Data)
}

// sniffImageExt 按图片字节魔数推断扩展名，未识别返回空串。
func sniffImageExt(data []byte) string {
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return "jpg"
	case len(data) >= 4 && string(data[:4]) == "\x89PNG":
		return "png"
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "webp"
	case len(data) >= 6 && string(data[:3]) == "GIF":
		return "gif"
	case len(data) >= 2 && data[0] == 'B' && data[1] == 'M':
		return "bmp"
	}
	return ""
}

// avgBitrateKbps 由文件大小与时长计算平均码率（kbps），向上取整；任一参数无效返回 0。
func avgBitrateKbps(size int64, durationSeconds float64) int {
	if size <= 0 || durationSeconds <= 0 {
		return 0
	}
	return int(math.Round(float64(size) * 8 / durationSeconds / 1000))
}

// lowerExt 把标签自带的扩展名归一化为小写、去点号的形式。
func lowerExt(ext string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), "."))
}
