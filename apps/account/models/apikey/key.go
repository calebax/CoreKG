package apikey

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSecretKey 生成随机密钥
func GenerateSecretKey() string {
	// 固定前缀
	prefix := "yg-"

	// 随机生成16字节的标识符
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	// 转换为十六进制字符串
	return prefix + hex.EncodeToString(randomBytes)
}
