package template

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:bucket-js
var templates embed.FS

// GetTemplateFS 返回模板文件的嵌入文件系统
func GetTemplateFS() embed.FS {
	return templates
}

// CopyTemplateDir 将嵌入的模板目录复制到目标目录
func CopyTemplateDir(templateType, dst string) error {
	srcDir := fmt.Sprintf("bucket-%s", templateType)

	return fs.WalkDir(templates, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 计算目标路径
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			// 创建目录
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return fmt.Errorf("create directory %s: %w", dstPath, err)
			}
		} else {
			// 读取嵌入文件
			data, err := templates.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read embedded file %s: %w", path, err)
			}

			// 写入目标文件
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return fmt.Errorf("write file %s: %w", dstPath, err)
			}
		}

		return nil
	})
}

// ReadTemplateFile 读取嵌入的模板文件内容
// templateType: 模板类型（如 "js" 或 "lua"）
// filePath: 模板内的文件路径（如 "apps/_example_.js"）
func ReadTemplateFile(templateType, filePath string) ([]byte, error) {
	fullPath := fmt.Sprintf("bucket-%s/%s", templateType, filePath)
	return templates.ReadFile(fullPath)
}

// ReadTemplateFileByName 通过完整模板名称读取文件
// templateName: 完整模板名称（如 "bucket-js"）
// filePath: 模板内的文件路径
func ReadTemplateFileByName(templateName, filePath string) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/%s", templateName, filePath)
	return templates.ReadFile(fullPath)
}
