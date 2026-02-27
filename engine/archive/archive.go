// Package archive 提供压缩文件解压功能。
package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

// Type 表示压缩文件类型。
type Type string

const (
	// ZIP 格式。
	ZIP Type = "zip"
	// TAR 格式。
	TAR Type = "tar"
	// TARGZ 格式（tar.gz）。
	TARGZ Type = "tar.gz"
	// TARXZ 格式（tar.xz）。
	TARXZ Type = "tar.xz"
	// TARBZ2 格式（tar.bz2）。
	TARBZ2 Type = "tar.bz2"
	// SevenZ 格式（7z）。
	SevenZ Type = "7z"
)

// Extractor 定义解压器接口。
type Extractor interface {
	// Extract 解压压缩文件到目标目录。
	Extract(src, dest string) error
	// ExtractWithProgress 带进度回调的解压。
	ExtractWithProgress(src, dest string, progress func(current, total int64)) error
	// List 列出压缩文件中的内容。
	List(src string) ([]FileInfo, error)
}

// FileInfo 表示压缩文件中的文件信息。
type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	IsDir   bool
	ModTime int64
}

// DetectType 从文件扩展名检测压缩类型。
func DetectType(path string) Type {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))

	switch {
	case ext == ".zip":
		return ZIP
	case ext == ".tar":
		return TAR
	case strings.HasSuffix(base, ".tar.gz") || strings.HasSuffix(base, ".tgz"):
		return TARGZ
	case strings.HasSuffix(base, ".tar.xz") || strings.HasSuffix(base, ".txz"):
		return TARXZ
	case strings.HasSuffix(base, ".tar.bz2") || strings.HasSuffix(base, ".tbz2"):
		return TARBZ2
	case ext == ".7z":
		return SevenZ
	default:
		return ""
	}
}

// Extract 自动检测类型并解压。
func Extract(src, dest string) error {
	typ := DetectType(src)
	if typ == "" {
		return fmt.Errorf("无法检测压缩类型: %s", src)
	}

	switch typ {
	case ZIP:
		return ExtractZip(src, dest)
	case TAR:
		return ExtractTar(src, dest)
	case TARGZ:
		return ExtractTarGz(src, dest)
	case TARXZ:
		return ExtractTarXz(src, dest)
	case TARBZ2:
		return ExtractTarBz2(src, dest)
	case SevenZ:
		return Extract7z(src, dest)
	default:
		return fmt.Errorf("不支持的压缩类型: %s", typ)
	}
}

// ExtractZip 解压 ZIP 文件。
func ExtractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("打开 zip 文件: %w", err)
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("创建目标目录: %w", err)
	}

	for _, f := range r.File {
		if err := extractZipFile(f, dest); err != nil {
			return fmt.Errorf("解压文件 %s: %w", f.Name, err)
		}
	}

	return nil
}

func extractZipFile(f *zip.File, dest string) error {
	path := filepath.Join(dest, f.Name)

	// 安全检查：防止路径穿越
	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("非法文件路径: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(path, f.Mode())
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// ExtractTar 解压 TAR 文件。
func ExtractTar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开 tar 文件: %w", err)
	}
	defer file.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("创建目标目录: %w", err)
	}

	tr := tar.NewReader(file)
	return extractTar(tr, dest)
}

// ExtractTarGz 解压 tar.gz 文件。
func ExtractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开 tar.gz 文件: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("创建 gzip 读取器: %w", err)
	}
	defer gzr.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("创建目标目录: %w", err)
	}

	tr := tar.NewReader(gzr)
	return extractTar(tr, dest)
}

func extractTar(tr *tar.Reader, dest string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dest, header.Name)

		// 安全检查：防止路径穿越
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法文件路径: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
	return nil
}

// ListZip 列出 ZIP 文件内容。
func ListZip(src string) ([]FileInfo, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("打开 zip 文件: %w", err)
	}
	defer r.Close()

	var files []FileInfo
	for _, f := range r.File {
		files = append(files, FileInfo{
			Name:    f.Name,
			Size:    int64(f.UncompressedSize64),
			Mode:    f.Mode(),
			IsDir:   f.FileInfo().IsDir(),
			ModTime: f.Modified.Unix(),
		})
	}
	return files, nil
}

// ListTar 列出 TAR 文件内容。
func ListTar(src string) ([]FileInfo, error) {
	file, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("打开 tar 文件: %w", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)
	var files []FileInfo

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		files = append(files, FileInfo{
			Name:    header.Name,
			Size:    header.Size,
			Mode:    os.FileMode(header.Mode),
			IsDir:   header.Typeflag == tar.TypeDir,
			ModTime: header.ModTime.Unix(),
		})
	}
	return files, nil
}

// IsArchive 检查文件是否为支持的压缩格式。
func IsArchive(path string) bool {
	return DetectType(path) != ""
}

func ExtractTarXz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开 tar.xz 文件: %w", err)
	}
	defer file.Close()

	xzr, err := xz.NewReader(file)
	if err != nil {
		return fmt.Errorf("创建 xz 读取器: %w", err)
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("创建目标目录: %w", err)
	}

	tr := tar.NewReader(xzr)
	return extractTar(tr, dest)
}

func ExtractTarBz2(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开 tar.bz2 文件: %w", err)
	}
	defer file.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("创建目标目录: %w", err)
	}

	bzr := bzip2.NewReader(file)
	tr := tar.NewReader(bzr)
	return extractTar(tr, dest)
}

func Extract7z(src, dest string) error {
	// 尝试使用系统安装的 7z 命令
	// 首先检查常见的 7z 安装路径
	sevenZipPaths := []string{
		`C:\Program Files\7-Zip\7z.exe`,
		`C:\Program Files (x86)\7-Zip\7z.exe`,
	}

	var sevenZipPath string
	for _, path := range sevenZipPaths {
		if _, err := os.Stat(path); err == nil {
			sevenZipPath = path
			break
		}
	}

	// 如果在常见路径找不到，尝试从 PATH 环境变量查找
	if sevenZipPath == "" {
		if path, err := exec.LookPath("7z"); err == nil {
			sevenZipPath = path
		}
	}

	if sevenZipPath == "" {
		return fmt.Errorf("未找到 7z 命令，请确保已安装 7-Zip 并添加到 PATH")
	}

	// 创建目标目录
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("创建目标目录: %w", err)
	}

	// 使用 7z 命令解压
	cmd := exec.Command(sevenZipPath, "x", "-y", "-o"+dest, src)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("7z 解压失败: %w\n输出: %s", err, string(output))
	}

	return nil
}
