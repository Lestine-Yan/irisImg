package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
)

// fixed 是一个固定时间，便于按年月断言路径。
var fixed = time.Date(2026, 6, 23, 15, 30, 0, 0, time.UTC)

func newSaver(t *testing.T, baseURL string) *Saver {
	t.Helper()
	root := filepath.Join(t.TempDir(), "imgs")
	s, err := NewSaver(config.StorageConfig{RootDir: root, PublicBaseURL: baseURL})
	if err != nil {
		t.Fatalf("new saver: %v", err)
	}
	return s
}

func TestRelPath(t *testing.T) {
	got := RelPath("abc", "PNG", fixed)
	want := "2026/06/abc.png"
	if got != want {
		t.Fatalf("rel path = %q, want %q", got, want)
	}

	// 空扩展名兜底
	if got := RelPath("abc", "", fixed); got != "2026/06/abc.bin" {
		t.Fatalf("empty ext fallback = %q", got)
	}

	// 允许带点的扩展名
	if got := RelPath("abc", ".Jpg", fixed); got != "2026/06/abc.jpg" {
		t.Fatalf("dotted ext = %q", got)
	}
}

func TestSaver_Save_WritesFile(t *testing.T) {
	s := newSaver(t, "")
	rel, err := s.Save([]byte("hello"), "abc", "png", fixed)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if rel != "2026/06/abc.png" {
		t.Fatalf("rel = %q", rel)
	}

	abs := filepath.Join(s.RootDir(), "2026", "06", "abc.png")
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("content = %q", string(data))
	}
}

func TestSaver_Save_IdempotentOnSecondWrite(t *testing.T) {
	s := newSaver(t, "")

	rel, err := s.Save([]byte("v1"), "abc", "png", fixed)
	if err != nil {
		t.Fatalf("first save: %v", err)
	}

	// 二次写：即使内容不同，路径已存在则直接复用，不报错也不覆盖。
	// 这模拟「同 hash 已存在」的秒传分支兜底；service 上层会先查库，
	// 真实路径不会走到这一步，但 Saver 自己也得保证幂等。
	rel2, err := s.Save([]byte("ignored"), "abc", "png", fixed)
	if err != nil {
		t.Fatalf("second save: %v", err)
	}
	if rel != rel2 {
		t.Fatalf("rel mismatch: %q vs %q", rel, rel2)
	}

	abs := filepath.Join(s.RootDir(), "2026", "06", "abc.png")
	data, _ := os.ReadFile(abs)
	if string(data) != "v1" {
		t.Fatalf("file overwritten unexpectedly: %q", string(data))
	}
}

func TestSaver_PublicURL(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
		rel     string
		want    string
	}{
		{"empty base url", "", "2026/06/abc.png", "/imgs/2026/06/abc.png"},
		{"absolute base", "https://img.example.com", "2026/06/abc.png", "https://img.example.com/2026/06/abc.png"},
		{"trim trailing slash", "https://img.example.com/", "2026/06/abc.png", "https://img.example.com/2026/06/abc.png"},
		{"leading slash on rel", "https://img.example.com", "/2026/06/abc.png", "https://img.example.com/2026/06/abc.png"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newSaver(t, c.baseURL)
			if got := s.PublicURL(c.rel); got != c.want {
				t.Fatalf("url = %q, want %q", got, c.want)
			}
		})
	}
}

func TestNewSaver_EmptyRootDir(t *testing.T) {
	_, err := NewSaver(config.StorageConfig{RootDir: ""})
	if err == nil || !strings.Contains(err.Error(), "root_dir") {
		t.Fatalf("expected root_dir error, got %v", err)
	}
}
