package audiometa

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestSniff 用表驱动覆盖魔数嗅探的命中与拒绝分支。
func TestSniff(t *testing.T) {
	ftypHeader := func() []byte {
		var buf bytes.Buffer
		var sz [4]byte
		binary.BigEndian.PutUint32(sz[:], 20)
		buf.Write(sz[:])
		buf.WriteString("ftyp")
		buf.WriteString("isom")
		var minor [4]byte
		binary.BigEndian.PutUint32(minor[:], 512)
		buf.Write(minor[:])
		buf.WriteString("mp41")
		return buf.Bytes()
	}()

	cases := []struct {
		name     string
		head     []byte
		wantFmt  Format
		wantMime string
	}{
		{"flac magic", []byte("fLaC\x00\x00\x00\x22"), FormatFLAC, "audio/flac"},
		{"ftyp box (isom brand, ffmpeg m4a)", ftypHeader, FormatM4A, "audio/mp4"},
		{"id3 header", []byte("ID3\x03\x00\x00\x00\x00"), FormatMP3, "audio/mpeg"},
		{"raw mpeg frame sync", []byte{0xFF, 0xFB, 0x90, 0x64, 0x00}, FormatMP3, "audio/mpeg"},
		{"jpeg bytes not mp3", append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, []byte("jpeg...")...), FormatUnknown, ""},
		{"mp3 bad bitrate idx", []byte{0xFF, 0xFB, 0xF0, 0x64}, FormatUnknown, ""},
		{"mp3 reserved version", []byte{0xFF, 0xEA, 0x90, 0x64}, FormatUnknown, ""},
		{"mp3 reserved samplerate idx", []byte{0xFF, 0xFB, 0x9C, 0x64}, FormatUnknown, ""},
		{"plain text", []byte("plain text, not audio"), FormatUnknown, ""},
		{"empty", nil, FormatUnknown, ""},
		{"two bytes", []byte{0xFF, 0xFB}, FormatUnknown, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, mime := Sniff(c.head)
			if f != c.wantFmt || mime != c.wantMime {
				t.Fatalf("Sniff = (%q, %q), want (%q, %q)", f, mime, c.wantFmt, c.wantMime)
			}
		})
	}
}
