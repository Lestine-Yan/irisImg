// Package audiometa 的嗅探逻辑：基于字节魔数识别音频容器格式。
//
// 刻意不复用 http.DetectContentType：它的嗅探表覆盖 mp3（ID3）但不覆盖
// m4a（ftyp 盒）与 flac（"fLaC"），而本项目坚持「嗅探内容而非信任客户端
// Content-Type / 扩展名」的纵深防御，故自持一份最小判定集。
package audiometa

// Sniff 基于内容头部魔数嗅探音频格式，返回 (格式, 规范化 MIME)。
// 未识别返回 (FormatUnknown, "")，由上层按不支持的类型拒收。
//
// 判定规则（顺序即优先级）：
//   - FLAC：头 4 字节 "fLaC"                        -> (flac, "audio/flac")
//   - MP4 家族：[4:8] == "ftyp"                     -> (m4a, "audio/mp4")
//   - MP3：头 3 字节 "ID3"（ID3v2 标签在前）        -> (mp3, "audio/mpeg")
//   - MP3：裸 MPEG 帧同步（0xFF 0xEx + 头部位合法） -> (mp3, "audio/mpeg")
//
// MP4 家族不按 ftyp major brand 区分 M4A/isom：ffmpeg 默认把 .m4a 编成
// major_brand=isom，严格按 brand 会误拒大量正常文件；真正的音频/视频区分由
// HasAudioTrack 走到 moov.trak.mdia.hdlr 做结构校验。
//
// head 传入内容的前若干字节即可（建议 ≥ 16 字节，函数内部对所有访问做了长度保护）。
func Sniff(head []byte) (Format, string) {
	// FLAC："fLaC" 魔数。
	if len(head) >= 4 && string(head[:4]) == "fLaC" {
		return FormatFLAC, "audio/flac"
	}

	// MP4 家族：4 字节盒子大小 + "ftyp" 类型 + major brand。仅凭 ftyp 判家族，
	// 不再细化 brand（见函数注释）。
	if len(head) >= 8 && string(head[4:8]) == "ftyp" {
		return FormatM4A, "audio/mp4"
	}

	// MP3：ID3v2 标签头在最前（绝大多数 mp3 文件）。
	if len(head) >= 3 && string(head[:3]) == "ID3" {
		return FormatMP3, "audio/mpeg"
	}

	// MP3：裸 MPEG 帧同步字 0xFF 0xEx，并校验头部位防误判
	//（对齐 tcolgate/mp3 的帧合法性判定：version/layer 非保留值、
	// 采样率索引 != 3、码率索引非 0 且非 0xF）。
	if len(head) >= 4 && head[0] == 0xFF && head[1]&0xE0 == 0xE0 && validMPEGFrameHeader(head) {
		return FormatMP3, "audio/mpeg"
	}

	return FormatUnknown, ""
}

// validMPEGFrameHeader 校验 4 字节 MPEG 帧头的关键位，降低把任意 0xFF 开头数据
// 误判为 mp3 的概率（与 tcolgate/mp3 Decoder.Decode 的同步判定保持一致）。
func validMPEGFrameHeader(head []byte) bool {
	version := (head[1] >> 3) & 0x03
	layer := (head[1] >> 1) & 0x03
	bitrateIdx := (head[2] >> 4) & 0x0F
	sampleRateIdx := (head[2] >> 2) & 0x03
	// version == 1（MPEG Reserved）、layer == 0（Layer Reserved）、
	// 码率索引 0（free）/0xF（bad）、采样率索引 3（reserved）均视为非法。
	if version == 1 || layer == 0 || bitrateIdx == 0 || bitrateIdx == 0xF || sampleRateIdx == 3 {
		return false
	}
	return true
}
