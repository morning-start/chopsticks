// Package manifest 定义了 Chopsticks 包管理器的数据结构。
// 这些数据结构用于在包之间共享数据，避免循环依赖。
package manifest

import "time"

// BucketConfig 包含软件源配置。
type BucketConfig struct {
	ID          string         `json:"id" toml:"id"`                   // 标识符
	Name        string         `json:"name" toml:"name"`               // 显示名称
	Author      string         `json:"author" toml:"author"`           // 作者
	Description string         `json:"description" toml:"description"` // 描述
	Homepage    string         `json:"homepage" toml:"homepage"`       // 主页
	License     string         `json:"license" toml:"license"`         // 许可证
	Repository  RepositoryInfo `json:"repository" toml:"repository"`   // 仓库
}

// RepositoryInfo 包含仓库信息。
type RepositoryInfo struct {
	Type   string `json:"type" toml:"type"`     // 类型（git, svn 等）
	URL    string `json:"url" toml:"url"`       // 地址
	Branch string `json:"branch" toml:"branch"` // 分支
}

// Bucket 表示完整的软件源信息。
type Bucket struct {
	Config      BucketConfig       // 配置
	Path        string             // 本地路径
	Apps        map[string]*AppRef // 应用索引
	LastUpdated time.Time          // 最后更新时间
}
