package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateShortTraceID 生成短跟踪标识符
func GenerateShortTraceID() string {
	b := make([]byte, 8) // 8 字节 = 16 字符 hex
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
