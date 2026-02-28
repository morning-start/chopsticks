// Package registry 提供 Windows 注册表操作功能。
package registry

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type Key = registry.Key

const (
	HKEY_CLASSES_ROOT   = registry.CLASSES_ROOT
	HKEY_CURRENT_USER   = registry.CURRENT_USER
	HKEY_LOCAL_MACHINE  = registry.LOCAL_MACHINE
	HKEY_USERS          = registry.USERS
	HKEY_CURRENT_CONFIG = registry.CURRENT_CONFIG
)

func ParseKey(path string) (registry.Key, string, error) {
	parts := splitKeyPath(path)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("无效的注册表路径: %s", path)
	}

	var key registry.Key
	switch normalizeKey(parts[0]) {
	case "HKEY_CLASSES_ROOT", "HKCR":
		key = registry.CLASSES_ROOT
	case "HKEY_CURRENT_USER", "HKCU":
		key = registry.CURRENT_USER
	case "HKEY_LOCAL_MACHINE", "HKLM":
		key = registry.LOCAL_MACHINE
	case "HKEY_USERS", "HKU":
		key = registry.USERS
	case "HKEY_CURRENT_CONFIG", "HKCC":
		key = registry.CURRENT_CONFIG
	default:
		return 0, "", fmt.Errorf("未知的注册表根键: %s", parts[0])
	}

	return key, parts[1], nil
}

func splitKeyPath(path string) []string {
	var parts []string
	var current string
	for _, c := range path {
		if c == '\\' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func normalizeKey(key string) string {
	result := ""
	for _, c := range key {
		if c >= 'A' && c <= 'Z' {
			result += string(c)
		} else if c >= 'a' && c <= 'z' {
			result += string(c - 32)
		} else {
			result += string(c)
		}
	}
	return result
}

func CreateKey(path string) (registry.Key, error) {
	parts := splitKeyPath(path)
	if len(parts) < 2 {
		return 0, fmt.Errorf("无效的注册表路径: %s", path)
	}

	var rootKey registry.Key
	switch normalizeKey(parts[0]) {
	case "HKEY_CLASSES_ROOT", "HKCR":
		rootKey = registry.CLASSES_ROOT
	case "HKEY_CURRENT_USER", "HKCU":
		rootKey = registry.CURRENT_USER
	case "HKEY_LOCAL_MACHINE", "HKLM":
		rootKey = registry.LOCAL_MACHINE
	case "HKEY_USERS", "HKU":
		rootKey = registry.USERS
	case "HKEY_CURRENT_CONFIG", "HKCC":
		rootKey = registry.CURRENT_CONFIG
	default:
		return 0, fmt.Errorf("未知的注册表根键: %s", parts[0])
	}

	key, _, err := registry.CreateKey(rootKey, parts[1], registry.WRITE|registry.CREATE_SUB_KEY)
	if err != nil {
		return 0, fmt.Errorf("创建注册表键: %w", err)
	}
	return key, nil
}

func OpenKey(path string) (registry.Key, error) {
	parts := splitKeyPath(path)
	if len(parts) < 2 {
		return 0, fmt.Errorf("无效的注册表路径: %s", path)
	}

	var rootKey registry.Key
	switch normalizeKey(parts[0]) {
	case "HKEY_CLASSES_ROOT", "HKCR":
		rootKey = registry.CLASSES_ROOT
	case "HKEY_CURRENT_USER", "HKCU":
		rootKey = registry.CURRENT_USER
	case "HKEY_LOCAL_MACHINE", "HKLM":
		rootKey = registry.LOCAL_MACHINE
	case "HKEY_USERS", "HKU":
		rootKey = registry.USERS
	case "HKEY_CURRENT_CONFIG", "HKCC":
		rootKey = registry.CURRENT_CONFIG
	default:
		return 0, fmt.Errorf("未知的注册表根键: %s", parts[0])
	}

	key, err := registry.OpenKey(rootKey, parts[1], registry.READ)
	if err != nil {
		return 0, fmt.Errorf("打开注册表键: %w", err)
	}
	return key, nil
}

func CloseKey(key registry.Key) error {
	return key.Close()
}

func SetStringValue(key registry.Key, name string, value string) error {
	return key.SetStringValue(name, value)
}

func GetStringValue(key registry.Key, name string) (string, error) {
	val, _, err := key.GetStringValue(name)
	return val, err
}

func SetDWordValue(key registry.Key, name string, value uint32) error {
	return key.SetDWordValue(name, value)
}

func GetDWordValue(key registry.Key, name string) (uint32, error) {
	val, _, err := key.GetIntegerValue(name)
	return uint32(val), err
}

func DeleteValue(key registry.Key, name string) error {
	return key.DeleteValue(name)
}

func DeleteKey(path string) error {
	parts := splitKeyPath(path)
	if len(parts) < 2 {
		return fmt.Errorf("无效的注册表路径: %s", path)
	}

	var rootKey registry.Key
	switch normalizeKey(parts[0]) {
	case "HKEY_CLASSES_ROOT", "HKCR":
		rootKey = registry.CLASSES_ROOT
	case "HKEY_CURRENT_USER", "HKCU":
		rootKey = registry.CURRENT_USER
	case "HKEY_LOCAL_MACHINE", "HKLM":
		rootKey = registry.LOCAL_MACHINE
	case "HKEY_USERS", "HKU":
		rootKey = registry.USERS
	case "HKEY_CURRENT_CONFIG", "HKCC":
		rootKey = registry.CURRENT_CONFIG
	default:
		return fmt.Errorf("未知的注册表根键: %s", parts[0])
	}

	return registry.DeleteKey(rootKey, parts[1])
}

// KeyExists 检查注册表键是否存在
func KeyExists(path string) (bool, error) {
	parts := splitKeyPath(path)
	if len(parts) < 2 {
		return false, fmt.Errorf("无效的注册表路径: %s", path)
	}

	var rootKey registry.Key
	switch normalizeKey(parts[0]) {
	case "HKEY_CLASSES_ROOT", "HKCR":
		rootKey = registry.CLASSES_ROOT
	case "HKEY_CURRENT_USER", "HKCU":
		rootKey = registry.CURRENT_USER
	case "HKEY_LOCAL_MACHINE", "HKLM":
		rootKey = registry.LOCAL_MACHINE
	case "HKEY_USERS", "HKU":
		rootKey = registry.USERS
	case "HKEY_CURRENT_CONFIG", "HKCC":
		rootKey = registry.CURRENT_CONFIG
	default:
		return false, fmt.Errorf("未知的注册表根键: %s", parts[0])
	}

	key, err := registry.OpenKey(rootKey, parts[1], registry.READ)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	key.Close()
	return true, nil
}

// ListKeys 列出指定键下的所有子键
func ListKeys(path string) ([]string, error) {
	parts := splitKeyPath(path)
	if len(parts) < 2 {
		return nil, fmt.Errorf("无效的注册表路径: %s", path)
	}

	var rootKey registry.Key
	switch normalizeKey(parts[0]) {
	case "HKEY_CLASSES_ROOT", "HKCR":
		rootKey = registry.CLASSES_ROOT
	case "HKEY_CURRENT_USER", "HKCU":
		rootKey = registry.CURRENT_USER
	case "HKEY_LOCAL_MACHINE", "HKLM":
		rootKey = registry.LOCAL_MACHINE
	case "HKEY_USERS", "HKU":
		rootKey = registry.USERS
	case "HKEY_CURRENT_CONFIG", "HKCC":
		rootKey = registry.CURRENT_CONFIG
	default:
		return nil, fmt.Errorf("未知的注册表根键: %s", parts[0])
	}

	key, err := registry.OpenKey(rootKey, parts[1], registry.READ|registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, fmt.Errorf("打开注册表键: %w", err)
	}
	defer key.Close()

	keys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("读取子键: %w", err)
	}

	return keys, nil
}

// ListValues 列出指定键下的所有值名称
func ListValues(path string) ([]string, error) {
	parts := splitKeyPath(path)
	if len(parts) < 2 {
		return nil, fmt.Errorf("无效的注册表路径: %s", path)
	}

	var rootKey registry.Key
	switch normalizeKey(parts[0]) {
	case "HKEY_CLASSES_ROOT", "HKCR":
		rootKey = registry.CLASSES_ROOT
	case "HKEY_CURRENT_USER", "HKCU":
		rootKey = registry.CURRENT_USER
	case "HKEY_LOCAL_MACHINE", "HKLM":
		rootKey = registry.LOCAL_MACHINE
	case "HKEY_USERS", "HKU":
		rootKey = registry.USERS
	case "HKEY_CURRENT_CONFIG", "HKCC":
		rootKey = registry.CURRENT_CONFIG
	default:
		return nil, fmt.Errorf("未知的注册表根键: %s", parts[0])
	}

	key, err := registry.OpenKey(rootKey, parts[1], registry.READ)
	if err != nil {
		return nil, fmt.Errorf("打开注册表键: %w", err)
	}
	defer key.Close()

	values, err := key.ReadValueNames(-1)
	if err != nil {
		return nil, fmt.Errorf("读取值: %w", err)
	}

	return values, nil
}
