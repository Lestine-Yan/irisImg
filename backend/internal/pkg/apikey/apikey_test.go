package apikey

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	plaintext, hash, prefix, err := Generate()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(plaintext) != KeyLength {
		t.Fatalf("expected plaintext length %d, got %d", KeyLength, len(plaintext))
	}
	if !IsValidFormat(plaintext) {
		t.Fatalf("generated key should pass format validation: %q", plaintext)
	}
	if len(hash) != 64 { // sha256 十六进制为 64 字符
		t.Fatalf("expected hash length 64, got %d", len(hash))
	}
	if !strings.HasPrefix(plaintext, prefix) || len(prefix) != prefixLength {
		t.Fatalf("prefix %q should be the first %d chars of %q", prefix, prefixLength, plaintext)
	}

	// 两次生成应不同。
	other, _, _, _ := Generate()
	if other == plaintext {
		t.Fatalf("two generated keys should differ")
	}
}

func TestHashStable(t *testing.T) {
	const s = "some-plaintext-key"
	h1 := Hash(s)
	h2 := Hash(s)
	if h1 != h2 {
		t.Fatalf("hash should be deterministic")
	}
	if h1 == Hash(s+"x") {
		t.Fatalf("different inputs should hash differently")
	}
}

func TestIsValidFormat(t *testing.T) {
	good, _, _, _ := Generate()
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"valid", good, true},
		{"too short", "abc", false},
		{"empty", "", false},
		{"wrong length", strings.Repeat("a", KeyLength-1), false},
		{"bad charset", strings.Repeat("*", KeyLength), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsValidFormat(tc.in); got != tc.want {
				t.Fatalf("IsValidFormat(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
