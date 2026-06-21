// Package apikey 提供 API 密钥的生成、哈希与格式校验工具。
//
// 明文密钥由 32 字节加密随机数经 base64(URL-safe, 无填充) 编码而成，长度固定 43 字符；
// 数据库只保存其 SHA-256 哈希（十六进制），明文仅在创建时返回一次，无法从库中反推。
package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// keyBytes 是随机熵的字节数。
const keyBytes = 32

// KeyLength 是 base64(RawURLEncoding) 编码 32 字节后的明文长度（ceil(32*4/3)=43）。
const KeyLength = 43

// prefixLength 是用于展示识别的明文前缀长度。
const prefixLength = 8

// Generate 生成一把新密钥，返回明文、明文的 SHA-256 哈希、以及用于展示的前缀。
// 明文仅应在创建响应中返回一次，调用方不应持久化明文。
func Generate() (plaintext, hash, prefix string, err error) {
	buf := make([]byte, keyBytes)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", fmt.Errorf("generate api key: %w", err)
	}
	plaintext = base64.RawURLEncoding.EncodeToString(buf)
	return plaintext, Hash(plaintext), plaintext[:prefixLength], nil
}

// Hash 返回明文密钥的 SHA-256 哈希（十六进制小写）。
func Hash(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// IsValidFormat 校验字符串是否符合密钥格式：长度为 KeyLength 且为合法的 base64url 字符集。
func IsValidFormat(s string) bool {
	if len(s) != KeyLength {
		return false
	}
	if _, err := base64.RawURLEncoding.DecodeString(s); err != nil {
		return false
	}
	return true
}
