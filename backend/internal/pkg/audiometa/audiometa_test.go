package audiometa

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/dhowden/tag"
)

// id3Frame 构造一个 ID3v2.3 帧。
func id3Frame(id string, body []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString(id)
	var sz [4]byte
	binary.BigEndian.PutUint32(sz[:], uint32(len(body)))
	buf.Write(sz[:])
	buf.Write([]byte{0x00, 0x00})
	buf.Write(body)
	return buf.Bytes()
}

// makeID3Only 构造只含 ID3v2.3 标签（不含音频帧）的内容，用于纯标签解析断言。
func makeID3Only(title string, mime string, picture []byte) []byte {
	frames := [][]byte{
		id3Frame("TIT2", append([]byte{0x00}, []byte(title)...)),
	}
	if picture != nil {
		apic := []byte{0x00}
		apic = append(apic, []byte(mime)...)
		apic = append(apic, 0x00, 0x03, 0x00)
		apic = append(apic, picture...)
		frames = append(frames, id3Frame("APIC", apic))
	}

	var body []byte
	for _, f := range frames {
		body = append(body, f...)
	}
	var buf bytes.Buffer
	buf.WriteString("ID3")
	buf.Write([]byte{0x03, 0x00, 0x00})
	n := len(body)
	buf.Write([]byte{byte(n >> 21 & 0x7F), byte(n >> 14 & 0x7F), byte(n >> 7 & 0x7F), byte(n & 0x7F)})
	buf.Write(body)
	return buf.Bytes()
}

// mp4Box 构造尺寸自洽的 MP4 盒子。
func mp4Box(typ string, payload []byte) []byte {
	var buf bytes.Buffer
	var sz [4]byte
	binary.BigEndian.PutUint32(sz[:], uint32(len(payload)+8))
	buf.Write(sz[:])
	buf.WriteString(typ)
	buf.Write(payload)
	return buf.Bytes()
}

// mp4Hdlr 构造 hdlr 盒。
func mp4Hdlr(handler string) []byte {
	payload := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	payload = append(payload, []byte(handler)...)
	payload = append(payload, make([]byte, 12)...)
	return mp4Box("hdlr", payload)
}

// makeMP4 构造 ftyp + moov(mvhd + trak(mdia(hdlr))) 的最小 MP4 结构。
func makeMP4(handlers ...string) []byte {
	var minor [4]byte
	binary.BigEndian.PutUint32(minor[:], 512)
	ftypPayload := append(append([]byte("isom"), minor[:]...), []byte("isom")...)
	ftyp := mp4Box("ftyp", ftypPayload)

	mvhdPayload := make([]byte, 0, 110)
	mvhdPayload = append(mvhdPayload, 0x00, 0x00, 0x00, 0x00) // version 0
	var u32 [4]byte
	mvhdPayload = append(mvhdPayload, u32[:]...)
	mvhdPayload = append(mvhdPayload, u32[:]...)
	binary.BigEndian.PutUint32(u32[:], 1000) // timescale
	mvhdPayload = append(mvhdPayload, u32[:]...)
	binary.BigEndian.PutUint32(u32[:], 60000) // duration 60s
	mvhdPayload = append(mvhdPayload, u32[:]...)
	binary.BigEndian.PutUint32(u32[:], 0x00010000) // rate
	mvhdPayload = append(mvhdPayload, u32[:]...)
	mvhdPayload = append(mvhdPayload, 0x01, 0x00)
	mvhdPayload = append(mvhdPayload, make([]byte, 10)...)
	mvhdPayload = append(mvhdPayload, make([]byte, 36)...)
	mvhdPayload = append(mvhdPayload, make([]byte, 24)...)
	binary.BigEndian.PutUint32(u32[:], 2) // next track id
	mvhdPayload = append(mvhdPayload, u32[:]...)
	mvhd := mp4Box("mvhd", mvhdPayload)

	moovPayload := append([]byte{}, mvhd...)
	for _, h := range handlers {
		moovPayload = append(moovPayload, mp4Box("trak", mp4Box("mdia", mp4Hdlr(h)))...)
	}
	return append(append([]byte{}, ftyp...), mp4Box("moov", moovPayload)...)
}

// TestParse_ID3TagsAndPicture 覆盖标签 + 封面解析：标题、封面字节与扩展名推断
//（MIME 为 image/png 时 Ext 应为 png——标签不提供 Ext，走 MIME 映射）。
func TestParse_ID3TagsAndPicture(t *testing.T) {
	pic := append([]byte{0x89, 'P', 'N', 'G'}, []byte("pngdata")...)
	info := Parse(makeID3Only("Hello", "image/png", pic), FormatMP3, 1024)

	if info.Title != "Hello" {
		t.Fatalf("title = %q, want Hello", info.Title)
	}
	if !bytes.Equal(info.PictureData, pic) {
		t.Fatalf("picture data mismatch: %d bytes", len(info.PictureData))
	}
	if info.PictureExt != "png" {
		t.Fatalf("picture ext = %q, want png", info.PictureExt)
	}
}

// TestParse_M4ADuration 覆盖 mvhd 时长解析（timescale 1000 / duration 60000 = 60s）。
func TestParse_M4ADuration(t *testing.T) {
	info := Parse(makeMP4("soun"), FormatM4A, 60000)
	if math.Abs(info.DurationSeconds-60.0) > 1e-9 {
		t.Fatalf("duration = %v, want 60", info.DurationSeconds)
	}
	// 平均码率 = 60000*8/60/1000 = 8kbps。
	if info.BitrateKBPS != 8 {
		t.Fatalf("bitrate = %d, want 8", info.BitrateKBPS)
	}
}

// TestParse_NeverFails 覆盖「永不因解析失败返回错误/panic」：
// 任何格式传入垃圾字节都应返回零值 Info 而非 panic/error。
func TestParse_NeverFails(t *testing.T) {
	garbage := []byte("definitely not a media file at all")
	cases := []struct {
		name string
		f    Format
	}{
		{"garbage as mp3", FormatMP3},
		{"garbage as m4a", FormatM4A},
		{"garbage as flac", FormatFLAC},
		{"unknown format", FormatUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			info := Parse(garbage, c.f, 33)
			if info.Title != "" || info.DurationSeconds != 0 || info.BitrateKBPS != 0 ||
				info.SampleRate != 0 || info.PictureData != nil {
				t.Fatalf("expected zero-value Info, got %+v", info)
			}
		})
	}
}

// TestHasAudioTrack 覆盖 MP4 结构校验：纯音轨 true、含视频轨 false、非 MP4 内容 false。
func TestHasAudioTrack(t *testing.T) {
	if !HasAudioTrack(makeMP4("soun")) {
		t.Fatalf("expected audio-only mp4 to pass")
	}
	if HasAudioTrack(makeMP4("soun", "vide")) {
		t.Fatalf("expected mp4 with video track to be rejected")
	}
	if HasAudioTrack([]byte("not an mp4")) {
		t.Fatalf("expected garbage to be rejected")
	}
}

// TestPictureExtChain 覆盖封面扩展名推断的优先级链：标签 Ext > MIME > 字节魔数。
func TestPictureExtChain(t *testing.T) {
	// 标签自带 Ext 优先（jpeg 归一化为 jpg）。
	if got := pictureExt(&tag.Picture{Ext: "JPEG"}); got != "jpg" {
		t.Fatalf("ext-from-tag = %q, want jpg", got)
	}
	// 无 Ext 时按 MIME。
	if got := pictureExt(&tag.Picture{MIMEType: "image/webp"}); got != "webp" {
		t.Fatalf("ext-from-mime = %q, want webp", got)
	}
	// Ext 与 MIME 都缺时按字节魔数（WebP：RIFF....WEBP——DetectContentType 不识别，须手工判定）。
	if got := pictureExt(&tag.Picture{Data: append([]byte("RIFF\x00\x00\x00\x00WEBP"), []byte("webpdata")...)}); got != "webp" {
		t.Fatalf("ext-from-magic = %q, want webp", got)
	}
	// 全部未知返回空串（调用方跳过落盘）。
	if got := pictureExt(&tag.Picture{Data: []byte("\x00\x01\x02\x03unknown")}); got != "" {
		t.Fatalf("unknown picture = %q, want empty", got)
	}
}
