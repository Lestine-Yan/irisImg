package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/storage"
)

// memAudioDAO 是 dao.AudioDAO 的内存实现，供 service 单测使用（镜像 memImageDAO）。
type memAudioDAO struct {
	byHash map[string]*model.Audio
	byID   map[int]*model.Audio
	nextID int
}

func newMemAudioDAO() *memAudioDAO {
	return &memAudioDAO{
		byHash: map[string]*model.Audio{},
		byID:   map[int]*model.Audio{},
	}
}

func (m *memAudioDAO) Create(_ context.Context, a *model.Audio) (*model.Audio, error) {
	m.nextID++
	a.ID = m.nextID
	a.CreatedAt = time.Unix(0, 0)
	cp := *a
	m.byHash[cp.Hash] = &cp
	m.byID[cp.ID] = &cp
	return &cp, nil
}

func (m *memAudioDAO) GetByID(_ context.Context, id int) (*model.Audio, error) {
	if v, ok := m.byID[id]; ok {
		return v, nil
	}
	return nil, dao.ErrNotFound
}

func (m *memAudioDAO) GetByHash(_ context.Context, hash string) (*model.Audio, error) {
	if v, ok := m.byHash[hash]; ok {
		return v, nil
	}
	return nil, dao.ErrNotFound
}

func (m *memAudioDAO) List(_ context.Context, _ model.AudioListQuery) ([]*model.Audio, int, error) {
	return nil, 0, errors.New("not used")
}

// ListByIDs 按主键集合批量查询现存的音频记录（不存在的 ID 静默跳过，内存实现）。
func (m *memAudioDAO) ListByIDs(_ context.Context, ids []int) ([]*model.Audio, error) {
	items := make([]*model.Audio, 0, len(ids))
	for _, id := range ids {
		if v, ok := m.byID[id]; ok {
			items = append(items, v)
		}
	}
	return items, nil
}

// DeleteByIDs 按主键集合批量删除音频记录，返回实际删除条数（不存在的 ID 静默跳过）。
func (m *memAudioDAO) DeleteByIDs(_ context.Context, ids []int) (int, error) {
	n := 0
	for _, id := range ids {
		if v, ok := m.byID[id]; ok {
			delete(m.byID, id)
			delete(m.byHash, v.Hash)
			n++
		}
	}
	return n, nil
}

// ListByKeyID 返回指定密钥关联的全部音频（内存实现，供删除密钥级联测试使用）。
func (m *memAudioDAO) ListByKeyID(_ context.Context, keyID int) ([]*model.Audio, error) {
	var items []*model.Audio
	for _, a := range m.byID {
		if a.KeyID != nil && *a.KeyID == keyID {
			items = append(items, a)
		}
	}
	return items, nil
}

// DeleteByKeyID 批量删除指定密钥关联的全部音频记录，返回删除条数。
func (m *memAudioDAO) DeleteByKeyID(_ context.Context, keyID int) (int, error) {
	n := 0
	for id, a := range m.byID {
		if a.KeyID != nil && *a.KeyID == keyID {
			delete(m.byID, id)
			delete(m.byHash, a.Hash)
			n++
		}
	}
	return n, nil
}

// UpdateCoverURL 更新用户覆盖的封面：nil 清除覆盖，非 nil 设置值（内存实现）。
func (m *memAudioDAO) UpdateCoverURL(_ context.Context, id int, coverURL *string) (*model.Audio, error) {
	v, ok := m.byID[id]
	if !ok {
		return nil, dao.ErrNotFound
	}
	v.CoverURL = coverURL
	cp := *v
	return &cp, nil
}

// 编译期断言：memAudioDAO 满足 dao.AudioDAO。
var _ dao.AudioDAO = (*memAudioDAO)(nil)

func newTestAudioService(t *testing.T) (*AudioService, *memAudioDAO) {
	t.Helper()
	cfg := config.StorageConfig{
		RootDir:         filepath.Join(t.TempDir(), "imgs"),
		AudioRootDir:    filepath.Join(t.TempDir(), "audios"),
		AudioMaxUploadSizeMB: 1,
		AudioAllowedMimeTypes: []string{
			"audio/mpeg", "audio/mp4", "audio/x-m4a", "audio/flac",
		},
	}
	saver, err := storage.NewAudioSaver(cfg)
	if err != nil {
		t.Fatalf("new audio saver: %v", err)
	}
	mem := newMemAudioDAO()
	return NewAudioService(mem, saver, cfg), mem
}

// mustUploadAudio 上传一个音频并在失败时 Fatal，返回落库后的记录。
func mustUploadAudio(t *testing.T, svc *AudioService, filename string, content []byte) *model.Audio {
	t.Helper()
	a, err := svc.Upload(context.Background(), &model.UploadAudioInput{Filename: filename, Content: content})
	if err != nil {
		t.Fatalf("upload %s: %v", filename, err)
	}
	return a
}

// ---- 最小音频 fixture（手造字节，不落 testdata 二进制文件） ----

// id3Frame 构造一个 ID3v2.3 帧（4 字节 ID + 4 字节大端长度 + 2 字节标志 + 负载）。
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

// makeID3v23 构造 ID3v2.3 标签头（标签体长度用 syncsafe 编码）。
func makeID3v23(frames ...[]byte) []byte {
	var body []byte
	for _, f := range frames {
		body = append(body, f...)
	}
	var buf bytes.Buffer
	buf.WriteString("ID3")
	buf.Write([]byte{0x03, 0x00, 0x00}) // v2.3，无标志位
	n := len(body)
	buf.Write([]byte{byte(n >> 21 & 0x7F), byte(n >> 14 & 0x7F), byte(n >> 7 & 0x7F), byte(n & 0x7F)})
	buf.Write(body)
	return buf.Bytes()
}

// makeMP3Frame 构造一个 MPEG1 Layer3 128kbps/44100 双声道帧（417 字节），
// 头部 0xFF 0xFB 0x90 0x64，时长 = 1152/44100 ≈ 26.1ms。
func makeMP3Frame() []byte {
	frame := make([]byte, 417)
	frame[0], frame[1], frame[2], frame[3] = 0xFF, 0xFB, 0x90, 0x64
	return frame
}

// makeMP3WithID3 构造「ID3v2.3 标签（可选 TIT2 标题 / APIC 封面）+ 一个 MPEG 帧」的 mp3。
func makeMP3WithID3(title string, picture []byte) []byte {
	var frames [][]byte
	if title != "" {
		// TIT2：编码 0（ISO-8859-1）+ 文本。
		frames = append(frames, id3Frame("TIT2", append([]byte{0x00}, []byte(title)...)))
	}
	if picture != nil {
		// APIC：编码 0 + MIME（null 结尾）+ 图片类型 0x03（封面）+ 空描述 + 图片字节。
		apic := []byte{0x00}
		apic = append(apic, []byte("image/jpeg")...)
		apic = append(apic, 0x00, 0x03, 0x00)
		apic = append(apic, picture...)
		frames = append(frames, id3Frame("APIC", apic))
	}
	tag := makeID3v23(frames...)
	return append(tag, makeMP3Frame()...)
}

// jpegMagic 返回带 JPEG 魔数的最小封面字节（足以让 pictureExt 按 MIME 命中 jpg）。
func jpegMagic() []byte {
	return append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, []byte("fake-jpeg-body")...)
}

// makeFLAC 构造「fLaC 魔数 + STREAMINFO 元数据块（无帧数据）」的最小 flac，
// 技术参数由入参决定。go-flac 的 ParseMetadata 只读元数据块，不需要帧数据。
func makeFLAC(sampleRate, channels, bitDepth int, totalSamples int64) []byte {
	var buf bytes.Buffer
	buf.WriteString("fLaC")
	// 块头：last=1 + type=0(STREAMINFO) + 24 位长度 34。
	buf.Write([]byte{0x80, 0x00, 0x00, 0x22})
	// 最小/最大块大小（16+16 位）、最小/最大帧大小（24+24 位），内容无关紧要。
	buf.Write([]byte{0x10, 0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	// 64 位打包：采样率 20 位 + 声道数-1 3 位 + 位深-1 5 位 + 总样本数 36 位。
	v := uint64(sampleRate)<<44 | uint64(channels-1)<<41 | uint64(bitDepth-1)<<36 | uint64(totalSamples)
	var b8 [8]byte
	binary.BigEndian.PutUint64(b8[:], v)
	buf.Write(b8[:])
	// MD5（16 字节，置零）。
	buf.Write(make([]byte, 16))
	return buf.Bytes()
}

// mp4Box 构造一个尺寸自洽的 MP4 盒子（4 字节大端 size + 4 字节类型 + 负载）。
func mp4Box(typ string, payload []byte) []byte {
	var buf bytes.Buffer
	var sz [4]byte
	binary.BigEndian.PutUint32(sz[:], uint32(len(payload)+8))
	buf.Write(sz[:])
	buf.WriteString(typ)
	buf.Write(payload)
	return buf.Bytes()
}

// mp4Hdlr 构造 hdlr 盒（handlerType 为 "soun" 或 "vide"，无 name 字段）。
func mp4Hdlr(handler string) []byte {
	payload := []byte{0x00, 0x00, 0x00, 0x00} // version + flags
	payload = append(payload, 0x00, 0x00, 0x00, 0x00) // predefined
	payload = append(payload, []byte(handler)...)
	payload = append(payload, make([]byte, 12)...) // reserved
	return mp4Box("hdlr", payload)
}

// makeM4A 构造最小 m4a：ftyp + moov(mvhd + trak(mdia(hdlr "soun"))) + 空 mdat。
// withVideo=true 时额外塞一个 vide trak，模拟视频 mp4。
func makeM4A(withVideo bool) []byte {
	// ftyp：major brand isom + minor 512 + compatible isom/iso2/mp41。
	var minor [4]byte
	binary.BigEndian.PutUint32(minor[:], 512)
	ftypPayload := append(append([]byte("isom"), minor[:]...), []byte("isomiso2mp41")...)
	ftyp := mp4Box("ftyp", ftypPayload)

	// mvhd（version 0）：timescale 1000、duration 213000 -> 213 秒。
	mvhdPayload := make([]byte, 0, 110)
	mvhdPayload = append(mvhdPayload, 0x00, 0x00, 0x00, 0x00) // version 0 + flags
	var u32 [4]byte
	mvhdPayload = append(mvhdPayload, u32[:]...)             // creation time
	mvhdPayload = append(mvhdPayload, u32[:]...)             // modification time
	binary.BigEndian.PutUint32(u32[:], 1000)                  // timescale
	mvhdPayload = append(mvhdPayload, u32[:]...)
	binary.BigEndian.PutUint32(u32[:], 213000)                // duration
	mvhdPayload = append(mvhdPayload, u32[:]...)
	binary.BigEndian.PutUint32(u32[:], 0x00010000)            // rate 1.0
	mvhdPayload = append(mvhdPayload, u32[:]...)
	mvhdPayload = append(mvhdPayload, 0x01, 0x00)            // volume
	mvhdPayload = append(mvhdPayload, make([]byte, 10)...)   // reserved
	mvhdPayload = append(mvhdPayload, make([]byte, 36)...)   // matrix
	mvhdPayload = append(mvhdPayload, make([]byte, 24)...)   // predefined
	binary.BigEndian.PutUint32(u32[:], 2)                    // next track id
	mvhdPayload = append(mvhdPayload, u32[:]...)
	mvhd := mp4Box("mvhd", mvhdPayload)

	audioTrak := mp4Box("trak", mp4Box("mdia", mp4Hdlr("soun")))
	moovPayload := append(append([]byte{}, mvhd...), audioTrak...)
	if withVideo {
		videoTrak := mp4Box("trak", mp4Box("mdia", mp4Hdlr("vide")))
		moovPayload = append(moovPayload, videoTrak...)
	}
	moov := mp4Box("moov", moovPayload)
	mdat := mp4Box("mdat", make([]byte, 8))

	out := append([]byte{}, ftyp...)
	out = append(out, moov...)
	return append(out, mdat...)
}

// ---- Upload ----

// TestAudioService_Upload_MP3WithTagsAndCover 覆盖带 ID3 标签与内嵌封面的 mp3 上传：
// 标题/技术参数来自标签与帧解析，封面抽取落盘 covers/ 并写入 embedded_cover_url。
func TestAudioService_Upload_MP3WithTagsAndCover(t *testing.T) {
	svc, _ := newTestAudioService(t)
	pic := jpegMagic()
	content := makeMP3WithID3("Test Title", pic)

	keyID := 7
	got, err := svc.Upload(context.Background(), &model.UploadAudioInput{
		Filename: "song.mp3",
		Content:  content,
		KeyID:    &keyID,
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if got.ID == 0 {
		t.Fatalf("expected non-zero id")
	}
	if got.Format != "mp3" || got.MimeType != "audio/mpeg" {
		t.Fatalf("format/mime = %q/%q", got.Format, got.MimeType)
	}
	if got.Title != "Test Title" {
		t.Fatalf("title = %q, want %q", got.Title, "Test Title")
	}
	if got.SampleRate != 44100 || got.Channels != 2 {
		t.Fatalf("samplerate/channels = %d/%d", got.SampleRate, got.Channels)
	}
	// 一个 MPEG1 Layer3 帧 = 1152 样本 / 44100 Hz ≈ 0.02612s。
	wantDur := 1152.0 / 44100.0
	if math.Abs(got.DurationSeconds-wantDur) > 1e-9 {
		t.Fatalf("duration = %v, want %v", got.DurationSeconds, wantDur)
	}
	if got.KeyID == nil || *got.KeyID != 7 {
		t.Fatalf("key id not propagated: %+v", got.KeyID)
	}
	if got.BitrateKbps == 0 {
		t.Fatalf("expected non-zero avg bitrate for %d bytes / %vs", got.Size, got.DurationSeconds)
	}

	// 封面应落盘到 covers/<sha256>.jpg 并写回 embedded_cover_url。
	sum := sha256.Sum256(pic)
	wantCover := "/audios/covers/" + hex.EncodeToString(sum[:]) + ".jpg"
	if got.EmbeddedCoverURL != wantCover {
		t.Fatalf("embedded cover url = %q, want %q", got.EmbeddedCoverURL, wantCover)
	}
	coverAbs := filepath.Join(svc.saver.RootDir(), "covers", hex.EncodeToString(sum[:])+".jpg")
	if _, err := os.Stat(coverAbs); err != nil {
		t.Fatalf("expected cover file on disk: %v", err)
	}

	// 音频文件本体应落盘且 URL 指向 /audios 前缀。
	audioAbs := filepath.Join(svc.saver.RootDir(), filepath.FromSlash(got.StoredPath))
	if _, err := os.Stat(audioAbs); err != nil {
		t.Fatalf("expected audio file on disk: %v", err)
	}
	if len(got.URL) == 0 || got.URL[:8] != "/audios/" {
		t.Fatalf("url should start with /audios/, got %q", got.URL)
	}
}

// TestAudioService_Upload_TitleFallsBackToFilename 覆盖无标签 mp3：
// 标题回退为文件名去扩展名，封面为空。
func TestAudioService_Upload_TitleFallsBackToFilename(t *testing.T) {
	svc, mem := newTestAudioService(t)
	content := makeMP3Frame()

	got := mustUploadAudio(t, svc, "my-song.mp3", content)
	if got.Title != "my-song" {
		t.Fatalf("title = %q, want %q", got.Title, "my-song")
	}
	if got.EmbeddedCoverURL != "" {
		t.Fatalf("expected empty embedded cover, got %q", got.EmbeddedCoverURL)
	}
	if _, ok := mem.byHash[got.Hash]; !ok {
		t.Fatalf("expected hash to be indexed in dao")
	}
}

// TestAudioService_Upload_FLACTechnical 覆盖 FLAC 的 STREAMINFO 技术参数解析
//（44.1kHz / 2 声道 / 16bit / 88200 样本 = 2 秒）。
func TestAudioService_Upload_FLACTechnical(t *testing.T) {
	svc, _ := newTestAudioService(t)
	content := makeFLAC(44100, 2, 16, 88200)

	got := mustUploadAudio(t, svc, "lossless.flac", content)
	if got.Format != "flac" || got.MimeType != "audio/flac" {
		t.Fatalf("format/mime = %q/%q", got.Format, got.MimeType)
	}
	if got.SampleRate != 44100 {
		t.Fatalf("samplerate = %d, want 44100", got.SampleRate)
	}
	if got.Channels != 2 {
		t.Fatalf("channels = %d, want 2", got.Channels)
	}
	if got.BitDepth != 16 {
		t.Fatalf("bit depth = %d, want 16", got.BitDepth)
	}
	if math.Abs(got.DurationSeconds-2.0) > 1e-9 {
		t.Fatalf("duration = %v, want 2.0", got.DurationSeconds)
	}
}

// TestAudioService_Upload_M4A 覆盖最小 m4a：格式/时长来自 mvhd，结构校验（soun 音轨）通过。
func TestAudioService_Upload_M4A(t *testing.T) {
	svc, _ := newTestAudioService(t)
	content := makeM4A(false)

	got := mustUploadAudio(t, svc, "chapter.m4a", content)
	if got.Format != "m4a" || got.MimeType != "audio/mp4" {
		t.Fatalf("format/mime = %q/%q", got.Format, got.MimeType)
	}
	if math.Abs(got.DurationSeconds-213.0) > 1e-9 {
		t.Fatalf("duration = %v, want 213.0 (mvhd 213000/1000)", got.DurationSeconds)
	}
}

// TestAudioService_Upload_RejectsVideoMP4 覆盖视频拒收：带 vide 音轨的 mp4
// 嗅探为 m4a 家族（ftyp），但结构校验发现视频轨，按不支持的类型拒收。
func TestAudioService_Upload_RejectsVideoMP4(t *testing.T) {
	svc, _ := newTestAudioService(t)
	_, err := svc.Upload(context.Background(), &model.UploadAudioInput{
		Filename: "clip.mp4",
		Content:  makeM4A(true),
	})
	if !errors.Is(err, ErrUnsupportedMime) {
		t.Fatalf("expected ErrUnsupportedMime for video mp4, got %v", err)
	}
}

func TestAudioService_Upload_DedupSecondTime(t *testing.T) {
	svc, mem := newTestAudioService(t)
	content := makeMP3WithID3("Same", nil)

	first := mustUploadAudio(t, svc, "a.mp3", content)
	second := mustUploadAudio(t, svc, "different-name.mp3", content)
	if second.ID != first.ID {
		t.Fatalf("expected dedup hit; first=%d second=%d", first.ID, second.ID)
	}
	if mem.nextID != 1 {
		t.Fatalf("expected only 1 row, got nextID=%d", mem.nextID)
	}
}

// TestAudioService_Upload_AdminNilKey 覆盖后台 JWT 直传：KeyID 保持 nil。
func TestAudioService_Upload_AdminNilKey(t *testing.T) {
	svc, mem := newTestAudioService(t)
	content := makeFLAC(48000, 2, 24, 96000)

	got := mustUploadAudio(t, svc, "hi-res.flac", content)
	if got.KeyID != nil {
		t.Fatalf("expected nil key_id for admin upload, got %v", *got.KeyID)
	}
	if stored, ok := mem.byHash[got.Hash]; !ok || stored.KeyID != nil {
		t.Fatalf("expected nil key_id in dao record")
	}
	if got.BitDepth != 24 {
		t.Fatalf("bit depth = %d, want 24", got.BitDepth)
	}
}

func TestAudioService_Upload_Errors(t *testing.T) {
	svc, _ := newTestAudioService(t)
	ctx := context.Background()

	t.Run("empty file", func(t *testing.T) {
		_, err := svc.Upload(ctx, &model.UploadAudioInput{Filename: "a.mp3", Content: nil})
		if !errors.Is(err, ErrEmptyFile) {
			t.Fatalf("expected ErrEmptyFile, got %v", err)
		}
	})

	t.Run("too large", func(t *testing.T) {
		big := make([]byte, svc.MaxBytes()+1)
		copy(big, "ID3")
		_, err := svc.Upload(ctx, &model.UploadAudioInput{Filename: "big.mp3", Content: big})
		if !errors.Is(err, ErrFileTooLarge) {
			t.Fatalf("expected ErrFileTooLarge, got %v", err)
		}
	})

	t.Run("unsupported mime (plain text)", func(t *testing.T) {
		_, err := svc.Upload(ctx, &model.UploadAudioInput{
			Filename: "a.mp3",
			Content:  []byte("plain text content, not audio at all"),
		})
		if !errors.Is(err, ErrUnsupportedMime) {
			t.Fatalf("expected ErrUnsupportedMime, got %v", err)
		}
	})
}

// ---- BatchDelete ----

// TestAudioService_BatchDelete 覆盖批量删除：
//   - 成功路径：音频文件删除、共享封面刻意保留、记录删除；
//   - 混入不存在的 ID：静默跳过；
//   - 空列表：零值返回。
func TestAudioService_BatchDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("success deletes audio but keeps shared cover", func(t *testing.T) {
		svc, mem := newTestAudioService(t)
		a := mustUploadAudio(t, svc, "a.mp3", makeMP3WithID3("A", jpegMagic()))
		b := mustUploadAudio(t, svc, "b.mp3", makeMP3WithID3("B", jpegMagic()))

		// 两条记录的封面字节相同 -> 内容寻址落同一文件。
		if a.EmbeddedCoverURL != b.EmbeddedCoverURL {
			t.Fatalf("expected shared cover path, got %q vs %q", a.EmbeddedCoverURL, b.EmbeddedCoverURL)
		}

		deleted, ids, err := svc.BatchDelete(ctx, []int{a.ID, b.ID})
		if err != nil {
			t.Fatalf("batch delete: %v", err)
		}
		if deleted != 2 || len(ids) != 2 {
			t.Fatalf("expected 2 deleted (ids=%v), got %d %v", ids, deleted, ids)
		}
		// 音频文件应已删除。
		for _, item := range []*model.Audio{a, b} {
			abs := filepath.Join(svc.saver.RootDir(), filepath.FromSlash(item.StoredPath))
			if _, statErr := os.Stat(abs); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("expected audio file deleted: %s (statErr=%v)", abs, statErr)
			}
		}
		// 共享封面应刻意保留。
		rel := "covers/" + filepath.Base(filepath.FromSlash(a.EmbeddedCoverURL))
		coverAbs := filepath.Join(svc.saver.RootDir(), filepath.FromSlash(rel))
		if _, statErr := os.Stat(coverAbs); statErr != nil {
			t.Fatalf("expected shared cover to survive delete: %v", statErr)
		}
		if _, ok := mem.byID[a.ID]; ok {
			t.Fatalf("expected record %d removed", a.ID)
		}
	})

	t.Run("missing ids are skipped", func(t *testing.T) {
		svc, mem := newTestAudioService(t)
		a := mustUploadAudio(t, svc, "a.flac", makeFLAC(44100, 2, 16, 44100))

		deleted, ids, err := svc.BatchDelete(ctx, []int{a.ID, 99999})
		if err != nil {
			t.Fatalf("batch delete: %v", err)
		}
		if deleted != 1 || len(ids) != 1 || ids[0] != a.ID {
			t.Fatalf("expected 1 deleted [%d], got %d %v", a.ID, deleted, ids)
		}
		if _, ok := mem.byID[a.ID]; ok {
			t.Fatalf("expected record %d removed", a.ID)
		}
	})

	t.Run("empty ids", func(t *testing.T) {
		svc, _ := newTestAudioService(t)
		deleted, ids, err := svc.BatchDelete(ctx, nil)
		if err != nil {
			t.Fatalf("batch delete: %v", err)
		}
		if deleted != 0 || ids != nil {
			t.Fatalf("expected zero result, got deleted=%d ids=%v", deleted, ids)
		}
	})
}

// ---- UpdateCover / GetByID ----

// TestAudioService_UpdateCover 覆盖改封面：设置覆盖 / 空串清除 / 不存在的 ID。
func TestAudioService_UpdateCover(t *testing.T) {
	svc, _ := newTestAudioService(t)
	ctx := context.Background()
	a := mustUploadAudio(t, svc, "song.mp3", makeMP3WithID3("S", jpegMagic()))

	// 设置覆盖。
	got, err := svc.UpdateCover(ctx, a.ID, " https://cdn.example.com/c.jpg ")
	if err != nil {
		t.Fatalf("update cover: %v", err)
	}
	if got.CoverURL == nil || *got.CoverURL != "https://cdn.example.com/c.jpg" {
		t.Fatalf("expected trimmed cover override, got %+v", got.CoverURL)
	}

	// 空串清除覆盖：恢复展示内嵌封面（cover_url 回到 nil）。
	got, err = svc.UpdateCover(ctx, a.ID, "")
	if err != nil {
		t.Fatalf("clear cover: %v", err)
	}
	if got.CoverURL != nil {
		t.Fatalf("expected cover override cleared, got %v", *got.CoverURL)
	}
	if got.EmbeddedCoverURL == "" {
		t.Fatalf("embedded cover should remain unchanged after clearing override")
	}

	// 不存在的 ID。
	if _, err := svc.UpdateCover(ctx, 99999, "x"); !errors.Is(err, ErrAudioNotFound) {
		t.Fatalf("expected ErrAudioNotFound, got %v", err)
	}
	if _, err := svc.GetByID(ctx, 99999); !errors.Is(err, ErrAudioNotFound) {
		t.Fatalf("expected ErrAudioNotFound from GetByID, got %v", err)
	}
}
