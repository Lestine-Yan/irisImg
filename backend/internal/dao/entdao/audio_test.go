package entdao

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/ent/apikey"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// newTestAudioDAO 在临时目录打开一个真实的 SQLite 数据库并完成迁移，构造 AudioDAO。
func newTestAudioDAO(t *testing.T) dao.AudioDAO {
	t.Helper()
	cfg := config.DatabaseConfig{
		Driver:      "sqlite",
		DSN:         filepath.Join(t.TempDir(), "test.db"),
		AutoMigrate: true,
	}
	client, err := Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	if err := Migrate(context.Background(), client, cfg); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewAudioDAO(client)
}

func sampleAudio() *model.Audio {
	return &model.Audio{
		Filename:         "song.mp3",
		StoredPath:       "2026/06/song.mp3",
		URL:              "/a/2026/06/song.mp3",
		Size:             4096,
		MimeType:         "audio/mpeg",
		Hash:             "abc123",
		Format:           "mp3",
		Title:            "Test Song",
		Artist:           "Artist",
		Album:            "Album",
		AlbumArtist:      "Album Artist",
		Composer:         "Composer",
		Genre:            "Pop",
		Year:             2026,
		TrackNumber:      3,
		DiscNumber:       1,
		DurationSeconds: 213.5,
		BitrateKbps:      320,
		SampleRate:       44100,
		Channels:         2,
		BitDepth:         0,
		EmbeddedCoverURL: "/audios/covers/cover.jpg",
	}
}

func TestAudioDAO_CreateAndGet(t *testing.T) {
	d := newTestAudioDAO(t)
	ctx := context.Background()

	created, err := d.Create(ctx, sampleAudio())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero id")
	}
	if created.CreatedAt.IsZero() {
		t.Fatalf("expected created_at to be filled")
	}
	if created.CoverURL != nil {
		t.Fatalf("expected nil cover override by default, got %v", *created.CoverURL)
	}

	got, err := d.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	// 全部元数据字段应原样落库读回。
	if got.Hash != "abc123" || got.Title != "Test Song" || got.Artist != "Artist" ||
		got.Album != "Album" || got.AlbumArtist != "Album Artist" || got.Composer != "Composer" ||
		got.Genre != "Pop" || got.Year != 2026 || got.TrackNumber != 3 || got.DiscNumber != 1 {
		t.Fatalf("tag fields mismatch: %+v", got)
	}
	if got.DurationSeconds != 213.5 || got.BitrateKbps != 320 || got.SampleRate != 44100 ||
		got.Channels != 2 || got.BitDepth != 0 {
		t.Fatalf("tech fields mismatch: %+v", got)
	}
	if got.Format != "mp3" || got.EmbeddedCoverURL != "/audios/covers/cover.jpg" {
		t.Fatalf("format/cover mismatch: %+v", got)
	}

	byHash, err := d.GetByHash(ctx, "abc123")
	if err != nil {
		t.Fatalf("get by hash: %v", err)
	}
	if byHash.ID != created.ID {
		t.Fatalf("get by hash id mismatch: %d != %d", byHash.ID, created.ID)
	}
}

func TestAudioDAO_NotFound(t *testing.T) {
	d := newTestAudioDAO(t)
	ctx := context.Background()

	if _, err := d.GetByID(ctx, 9999); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := d.GetByHash(ctx, "missing"); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := d.UpdateCoverURL(ctx, 9999, nil); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on update cover, got %v", err)
	}
}

// TestAudioDAO_ListFilterAndOrder 覆盖 List 的过滤、排序、分页（镜像 image 同名测试）。
// 为让排序可测，直接用 ent client 写入带明确 created_at 的记录。
func TestAudioDAO_ListFilterAndOrder(t *testing.T) {
	d := newTestAudioDAO(t)
	ctx := context.Background()
	impl := d.(*audioDAO)

	// audio.key_id 外键依赖 api_keys.id，先插入两把密钥。
	for _, name := range []string{"key-one", "key-two"} {
		if _, err := impl.client.ApiKey.Create().
			SetName(name).
			SetKeyHash("hash-" + name).
			SetPrefix("pfx").
			SetScope(apikey.ScopeReadwrite).
			Save(ctx); err != nil {
			t.Fatalf("create key %s: %v", name, err)
		}
	}

	base := time.Now().Truncate(time.Hour).UTC()
	records := []struct {
		hash string
		key  int
		t    time.Time
	}{
		{"h1", 1, base.Add(3 * time.Minute)},
		{"h2", 1, base.Add(1 * time.Minute)},
		{"h3", 2, base.Add(2 * time.Minute)},
	}
	for _, r := range records {
		if _, err := impl.client.Audio.Create().
			SetFilename(r.hash + ".mp3").
			SetStoredPath("p/" + r.hash).
			SetURL("/a/" + r.hash).
			SetSize(100).
			SetMimeType("audio/mpeg").
			SetHash(r.hash).
			SetNillableKeyID(&r.key).
			SetCreatedAt(r.t).
			Save(ctx); err != nil {
			t.Fatalf("create %s: %v", r.hash, err)
		}
	}

	// 全部升序：h2 → h3 → h1。
	items, total, err := d.List(ctx, model.AudioListQuery{Order: "asc"})
	if err != nil {
		t.Fatalf("list asc: %v", err)
	}
	if total != 3 || len(items) != 3 || items[0].Hash != "h2" || items[2].Hash != "h1" {
		t.Fatalf("asc: total=%d items=%+v", total, items)
	}

	// 倒序：h1 → h3 → h2。
	items, _, err = d.List(ctx, model.AudioListQuery{Order: "desc"})
	if err != nil {
		t.Fatalf("list desc: %v", err)
	}
	if items[0].Hash != "h1" || items[2].Hash != "h2" {
		t.Fatalf("desc order: %s %s %s", items[0].Hash, items[1].Hash, items[2].Hash)
	}

	// 按 key_id=1 过滤：h2、h1。
	key1 := 1
	items, total, err = d.List(ctx, model.AudioListQuery{KeyID: &key1, Order: "asc"})
	if err != nil {
		t.Fatalf("list key1: %v", err)
	}
	if total != 2 || len(items) != 2 || items[0].Hash != "h2" || items[1].Hash != "h1" {
		t.Fatalf("key1: total=%d items=%+v", total, items)
	}

	// 分页：升序 offset=1, limit=1 命中 h3。
	items, total, err = d.List(ctx, model.AudioListQuery{Order: "asc", Offset: 1, Limit: 1})
	if err != nil {
		t.Fatalf("list paged: %v", err)
	}
	if total != 3 || len(items) != 1 || items[0].Hash != "h3" {
		t.Fatalf("paged: total=%d items=%+v", total, items)
	}
}

// TestAudioDAO_UpdateCoverURL 覆盖改封面的三种情形：设置覆盖 → 清除覆盖（nil）→ 记录不存在。
func TestAudioDAO_UpdateCoverURL(t *testing.T) {
	d := newTestAudioDAO(t)
	ctx := context.Background()

	created, err := d.Create(ctx, sampleAudio())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// 设置覆盖。
	override := "https://cdn.example.com/cover.jpg"
	got, err := d.UpdateCoverURL(ctx, created.ID, &override)
	if err != nil {
		t.Fatalf("set cover: %v", err)
	}
	if got.CoverURL == nil || *got.CoverURL != override {
		t.Fatalf("expected override set, got %+v", got.CoverURL)
	}

	// 清除覆盖：SetNillable(nil) 是「不操作」，必须走 ClearCoverURL 才能回到 NULL。
	got, err = d.UpdateCoverURL(ctx, created.ID, nil)
	if err != nil {
		t.Fatalf("clear cover: %v", err)
	}
	if got.CoverURL != nil {
		t.Fatalf("expected override cleared, got %v", *got.CoverURL)
	}
	// 内嵌封面不受影响。
	if got.EmbeddedCoverURL != "/audios/covers/cover.jpg" {
		t.Fatalf("embedded cover changed unexpectedly: %q", got.EmbeddedCoverURL)
	}
}

// TestAudioDAO_ListAndDeleteByKeyID 覆盖按密钥批量查询与删除（级联清理用）。
func TestAudioDAO_ListAndDeleteByKeyID(t *testing.T) {
	d := newTestAudioDAO(t)
	ctx := context.Background()
	impl := d.(*audioDAO)

	for _, name := range []string{"k1", "k2"} {
		if _, err := impl.client.ApiKey.Create().
			SetName(name).
			SetKeyHash("hash-" + name).
			SetPrefix("pfx").
			SetScope(apikey.ScopeReadwrite).
			Save(ctx); err != nil {
			t.Fatalf("create key %s: %v", name, err)
		}
	}

	mk := func(hash string, keyID int) {
		a := sampleAudio()
		a.Hash = hash
		a.StoredPath = "p/" + hash
		a.KeyID = &keyID
		if _, err := d.Create(ctx, a); err != nil {
			t.Fatalf("create %s: %v", hash, err)
		}
	}
	mk("a1", 1)
	mk("a2", 1)
	mk("b1", 2)

	items, err := d.ListByKeyID(ctx, 1)
	if err != nil || len(items) != 2 {
		t.Fatalf("expected 2 audios for key 1, got %d (err %v)", len(items), err)
	}

	n, err := d.DeleteByKeyID(ctx, 1)
	if err != nil || n != 2 {
		t.Fatalf("expected 2 removed, got %d (err %v)", n, err)
	}
	if left, err := d.ListByKeyID(ctx, 1); err != nil || len(left) != 0 {
		t.Fatalf("expected key1 empty, got %d (err %v)", len(left), err)
	}
	if left, err := d.ListByKeyID(ctx, 2); err != nil || len(left) != 1 {
		t.Fatalf("expected key2 still 1, got %d (err %v)", len(left), err)
	}
}

// TestAudioDAO_ListAndDeleteByIDs 覆盖按主键集合批量查询与删除（音频中心批量删除）。
func TestAudioDAO_ListAndDeleteByIDs(t *testing.T) {
	d := newTestAudioDAO(t)
	ctx := context.Background()

	var ids []int
	for i := 0; i < 3; i++ {
		a := sampleAudio()
		a.Hash = "ids-hash-" + string(rune('a'+i))
		a.StoredPath = "p/" + a.Hash
		created, err := d.Create(ctx, a)
		if err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
		ids = append(ids, created.ID)
	}

	// 混入不存在的 ID：只返回现存的 2 条。
	items, err := d.ListByIDs(ctx, []int{ids[0], ids[1], 99999})
	if err != nil {
		t.Fatalf("list by ids: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// 删除其中 1 条（混入不存在 ID）：计数为 1。
	n, err := d.DeleteByIDs(ctx, []int{ids[0], 99999})
	if err != nil || n != 1 {
		t.Fatalf("expected 1 removed, got %d (err %v)", n, err)
	}
	if left, err := d.ListByIDs(ctx, ids); err != nil || len(left) != 2 {
		t.Fatalf("expected 2 left, got %d (err %v)", len(left), err)
	}

	// 空 ids 防御。
	if empty, err := d.ListByIDs(ctx, nil); err != nil || len(empty) != 0 {
		t.Fatalf("empty ListByIDs = %v (err %v), want empty", empty, err)
	}
	if n, err := d.DeleteByIDs(ctx, nil); err != nil || n != 0 {
		t.Fatalf("empty DeleteByIDs = %d (err %v), want 0", n, err)
	}
}
