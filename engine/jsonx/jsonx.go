// Package jsonx 提供 JSON 编码/解码功能。
package jsonx

import (
	"encoding/json"
	"fmt"
)

// Encode 将 Go 值编码为 JSON 字符串。
func Encode(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encode json: %w", err)
	}
	return string(data), nil
}

// EncodeIndent 将 Go 值编码为格式化的 JSON 字符串。
func EncodeIndent(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode json: %w", err)
	}
	return string(data), nil
}

// Decode 将 JSON 字符串解码为 Go 值。
func Decode(s string) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

// DecodeTo 将 JSON 字符串解码到指定类型的值。
func DecodeTo(s string, v interface{}) error {
	if err := json.Unmarshal([]byte(s), v); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

// Module 为脚本引擎提供 json 注册。
type Module struct{}

// NewModule 创建新的 json 模块。
func NewModule() *Module {
	return &Module{}
}
