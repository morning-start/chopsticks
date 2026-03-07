// Package manifest 定义了 Chopsticks 包管理器的数据结构。
// 这些数据结构用于在包之间共享数据，避免循环依赖。
package manifest

import "time"

// App 表示一个完整的软件包。
type App struct {
	Script *AppScript // 脚本信息（来自 .lua/.js 文件）
	Meta   *AppMeta   // 元数据（来自 .meta.json 文件）
	Ref    *AppRef    // 引用信息
}

// AppScript 定义脚本信息（不变部分）。
type AppScript struct {
	Name         string               // 软件名称
	Description  string               // 描述
	Homepage     string               // 主页 URL
	License      string               // 许可证
	Category     string               // 分类
	Tags         []string             // 标签
	Maintainer   string               // 维护者
	Bucket       string               // 所属软件源
	Dependencies []Dependency         // 依赖列表（扁平结构，向后兼容）
	Resources    *ResourceDeclaration // 资源声明（端口、环境变量、注册表等）
}

// Dependency 表示应用依赖
type Dependency struct {
	Name       string            // 依赖应用名称
	Version    string            // 版本约束（如 ">=1.0.0", "^2.0.0"）
	Optional   bool              // 是否为可选依赖
	Conditions map[string]string // 安装条件（如 {"os": "windows"}）
}

// Dependencies 表示完整的依赖声明
type Dependencies struct {
	Runtime   []Dependency // 运行时库（全局共享，引用计数）
	Tools     []Dependency // 工具软件（全局共享，检查 installed_on_request）
	Libraries []Dependency // 库文件（不共享，随主软件卸载）
	Conflicts []string     // 冲突软件（不允许同时存在）
}

// AppMeta 定义元数据（动态部分）。
type AppMeta struct {
	Version  string                 // 当前版本
	Versions map[string]VersionInfo // 所有版本信息
}

// VersionInfo 包含版本信息。
type VersionInfo struct {
	Version    string                  // 版本号
	ReleasedAt time.Time               // 发布时间
	Downloads  map[string]DownloadInfo // 各架构下载信息
}

// DownloadInfo 包含下载信息。
type DownloadInfo struct {
	URL  string // 下载地址
	Hash string // 校验和
	Size int64  // 文件大小
	Type string // 压缩类型（zip, tar.gz, 7z）
}

// AppRef 是应用的引用（用于索引）。
type AppRef struct {
	Name        string   // 名称
	Description string   // 描述
	Version     string   // 最新版本
	Category    string   // 分类
	Tags        []string // 标签
	ScriptPath  string   // 脚本文件路径
	MetaPath    string   // 元数据文件路径
}

// InstalledApp 表示已安装的应用。
type InstalledApp struct {
	Name        string    // 名称
	Version     string    // 版本
	Bucket      string    // 所属软件源
	InstallDir  string    // 安装目录
	InstalledAt time.Time // 安装时间
	UpdatedAt   time.Time // 更新时间
}

// AppInfo 包含应用信息（用于展示）。
type AppInfo struct {
	Name             string
	Description      string
	Homepage         string
	License          string
	Category         string
	Tags             []string
	Version          string
	Bucket           string
	Installed        bool
	InstalledVersion string
}

// RuntimeInfo 表示运行时库信息
type RuntimeInfo struct {
	Version     string    // 版本号
	InstalledAt time.Time // 安装时间
	RequiredBy  []string  // 依赖此运行时库的软件列表
	RefCount    int       // 引用计数
	Size        int64     // 占用字节数
}

// Orphans 表示孤儿依赖
type Orphans struct {
	Runtime []string // 孤儿运行时库
	Tools   []string // 孤儿工具软件
}

// ResourceDeclaration 定义应用的资源声明。
// 用于声明应用运行所需的系统资源，如端口、环境变量、注册表项等。
type ResourceDeclaration struct {
	Ports    []PortDeclaration     `json:"ports,omitempty"`    // 端口声明列表
	EnvVars  []EnvVarDeclaration   `json:"env_vars,omitempty"` // 环境变量声明列表
	Registry []RegistryDeclaration `json:"registry,omitempty"` // 注册表声明列表
}

// PortDeclaration 定义端口资源声明。
// 用于声明应用需要使用的网络端口，包括端口号、协议和用途说明。
type PortDeclaration struct {
	Port        int    `json:"port"`                  // 端口号（1-65535）
	Protocol    string `json:"protocol,omitempty"`    // 协议类型（tcp/udp，默认为 tcp）
	Description string `json:"description,omitempty"` // 端口用途说明
	Required    bool   `json:"required,omitempty"`    // 是否为必需端口（默认为 true）
}

// EnvVarDeclaration 定义环境变量声明。
// 用于声明应用需要设置或读取的环境变量，包括变量名、默认值和说明。
type EnvVarDeclaration struct {
	Name        string `json:"name"`                  // 环境变量名称
	Value       string `json:"value,omitempty"`       // 默认值（可选）
	Description string `json:"description,omitempty"` // 变量用途说明
	Required    bool   `json:"required,omitempty"`    // 是否为必需变量（默认为 false）
}

// RegistryDeclaration 定义 Windows 注册表声明。
// 用于声明应用需要创建或修改的注册表项，包括键路径、值名称和值数据。
type RegistryDeclaration struct {
	Hive        string `json:"hive"`                  // 根键（HKCU/HKLM/HKCR/HKCC/HKU）
	Key         string `json:"key"`                   // 注册表键路径（如 "Software\\MyApp"）
	ValueName   string `json:"value_name,omitempty"`  // 值名称（空字符串表示默认值）
	ValueType   string `json:"value_type,omitempty"`  // 值类型（STRING/EXPAND_SZ/BINARY/DWORD/QWORD/MULTI_SZ，默认为 STRING）
	ValueData   string `json:"value_data,omitempty"`  // 值数据
	Description string `json:"description,omitempty"` // 注册表项用途说明
	Required    bool   `json:"required,omitempty"`    // 是否为必需项（默认为 true）
}
