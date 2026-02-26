// Package checksum 提供校验和计算与验证功能。
package checksum

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

// Algorithm 表示校验和算法类型。
type Algorithm string

const (
	// MD5 算法。
	MD5 Algorithm = "md5"
	// SHA256 算法。
	SHA256 Algorithm = "sha256"
	// SHA512 算法。
	SHA512 Algorithm = "sha512"
)

// Calculator 定义校验和计算器接口。
type Calculator interface {
	// Calculate 计算文件的校验和。
	Calculate(path string) (string, error)
	// CalculateString 计算字符串的校验和。
	CalculateString(data string) string
	// Verify 验证文件的校验和。
	Verify(path, expected string) (bool, error)
	// VerifyString 验证字符串的校验和。
	VerifyString(data, expected string) bool
}

// calculator 是 Calculator 的实现。
type calculator struct {
	alg Algorithm
}

// 编译时接口检查。
var _ Calculator = (*calculator)(nil)

// New 创建新的 Calculator。
func New(alg Algorithm) Calculator {
	return &calculator{alg: alg}
}

// Calculate 计算文件的校验和。
func (c *calculator) Calculate(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开文件: %w", err)
	}
	defer file.Close()

	h := c.newHash()
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("计算校验和: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// CalculateString 计算字符串的校验和。
func (c *calculator) CalculateString(data string) string {
	h := c.newHash()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Verify 验证文件的校验和。
func (c *calculator) Verify(path, expected string) (bool, error) {
	actual, err := c.Calculate(path)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(actual, expected), nil
}

// VerifyString 验证字符串的校验和。
func (c *calculator) VerifyString(data, expected string) bool {
	actual := c.CalculateString(data)
	return strings.EqualFold(actual, expected)
}

// newHash 根据算法创建 hash.Hash 实例。
func (c *calculator) newHash() hash.Hash {
	switch c.alg {
	case MD5:
		return md5.New()
	case SHA256:
		return sha256.New()
	case SHA512:
		return sha512.New()
	default:
		return sha256.New()
	}
}

// CalculateFile 使用指定算法计算文件校验和。
func CalculateFile(path string, alg Algorithm) (string, error) {
	return New(alg).Calculate(path)
}

// VerifyFile 验证文件的校验和。
func VerifyFile(path, expected string, alg Algorithm) (bool, error) {
	return New(alg).Verify(path, expected)
}

// CalculateBytes 计算字节切片的校验和。
func CalculateBytes(data []byte, alg Algorithm) string {
	c := New(alg)
	h := c.(*calculator).newHash()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// AutoDetectAlgorithm 从校验和字符串自动检测算法。
func AutoDetectAlgorithm(checksum string) Algorithm {
	switch len(checksum) {
	case 32:
		return MD5
	case 64:
		return SHA256
	case 128:
		return SHA512
	default:
		return SHA256
	}
}

// IsValidChecksum 检查校验和字符串格式是否有效。
func IsValidChecksum(checksum string) bool {
	if len(checksum) == 0 {
		return false
	}
	// 检查是否为十六进制字符
	for _, c := range checksum {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	// 检查长度
	length := len(checksum)
	return length == 32 || length == 64 || length == 128
}
