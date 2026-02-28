# Chopsticks Go 编码规范

> 项目遵循的 Go 语言编码规范和最佳实践

---

## 1. 设计原则

### 1.1 核心原则（按优先级）

| 原则                           | 说明                     |
| ------------------------------ | ------------------------ |
| **清晰性 (Clarity)**           | 代码意图明确，易于理解   |
| **简洁性 (Simplicity)**        | 使用最简单的方案实现目标 |
| **一致性 (Consistency)**       | 与 Go 生态保持一致       |
| **可维护性 (Maintainability)** | 易于后续修改             |

### 1.2 关键决策

- **显式优于隐式** - 避免魔法，代码行为明确
- **组合优于继承** - 使用接口和嵌入实现复用
- **依赖注入** - 通过接口解耦模块

---

## 2. 项目结构

```
chopsticks/
├── cmd/chopsticks/          # 程序入口
│   └── cli/               # CLI 命令
├── core/                    # 核心业务
│   ├── app/                # 应用管理
│   ├── bucket/             # 软件源管理
│   ├── store/              # 数据存储
│   └── manifest/           # 数据结构
├── infra/                  # 基础设施
│   ├── git/               # Git 操作
│   └── installer/         # 安装程序处理
└── engine/                  # 引擎系统（供开发者使用）
    ├── script/            # 脚本执行器
    ├── fetch/            # HTTP 请求
    ├── fsutil/           # 文件操作
    └── ...               # 其他功能模块
```

---

## 3. 包命名规范

### 3.1 包名规则

- ✅ 使用小写字母
- ✅ 简短且有意义
- ✅ 无下划线或混合格式
- ✅ 与目录名一致

```go
// Good
package app      // 目录名: app/
package install     // 目录名: install/

// Bad
package App     // 大写
package app_pkg // 下划线
package utils    // 过于笼统
```

### 3.2 包注释

```go
// Package app 提供应用（软件包）管理功能。
package app

// Package install 提供软件安装、卸载和更新功能。
package install
```

---

## 4. 命名规范

### 4.1 接口命名

| 类型 | 规则      | 示例                               |
| ---- | --------- | ---------------------------------- |
| 接口 | 名词/角色 | `Manager`, `Cooker`, `Storage`     |
| 实现 | 小写      | `manager`, `cooker`, `boltStorage` |

```go
// 接口定义
type Manager interface {
    Add(ctx context.Context, name, url string) error
    Remove(ctx context.Context, name string) error
}

// 实现
type manager struct {
    // 私有字段
}

var _ Manager = (*manager)(nil) // 编译时检查
```

### 4.2 函数命名

| 类型     | 规则           | 示例                          |
| -------- | -------------- | ----------------------------- |
| 构造函数 | `New` + 类型名 | `NewManager()`, `NewCooker()` |
| 工厂函数 | `New` + 描述   | `NewLuaEngine()`              |
| 方法     | 动词/动词短语  | `Add()`, `Remove()`           |

### 4.3 变量命名

| 类型       | 规则 | 示例                       |
| ---------- | ---- | -------------------------- |
| 导出变量   | 驼峰 | `MaxLength`, `DefaultPath` |
| 未导出变量 | 混合 | `maxLength`, `defaultPath` |
| 常量       | 驼峰 | `MaxRetries`               |

### 4.4 Receiver 命名

```go
// Good: 1-2 个字母的缩写
func (m *manager) Add(...)      {}
func (c *cooker) Cook(...)      {}

// Bad: 完整单词
func (manager *manager) Add()  {}
```

---

## 5. 接口设计

### 5.1 消费者定义接口

接口由**使用者**（消费者）定义，而非实现者：

```go
// 使用者定义接口
type AppManager interface {
    Install(ctx context.Context, spec InstallSpec) error
    Remove(ctx context.Context, name string) error
}

// 实现者提供具体实现
type appManager struct {
    storage Storage
    installer Installer
}

// 编译时接口检查
var _ AppManager = (*appManager)(nil)
```

### 5.2 接口位置

- 接口放在**使用者**包中
- 实现可以引用使用者的接口类型
- 避免循环依赖

---

## 6. 错误处理

### 6.1 Sentinel 错误

```go
var (
    ErrBucketNotFound         = errors.New("bucket not found")
    ErrAppNotFound        = errors.New("app not found")
    ErrAppAlreadyInstalled = errors.New("app already installed")
)
```

### 6.2 错误包装

```go
// 使用 %w 保留错误链
return nil, fmt.Errorf("加载失败: %w", err)

// 上下文信息
return nil, fmt.Errorf("download %s failed: %w", url, err)
```

### 6.3 错误检查

```lua
-- Good: 清晰检查
if err != nil {
    return nil, fmt.Errorf("failed: %w", err)
}

// Bad: 隐藏错误
if err != nil {
    return nil, err
}
```

---

## 7. 文档注释

### 7.1 包注释

```go
// Package bucket 提供软件源（碗）管理功能。
//
// 主要功能包括：
//   - 添加和删除软件源
//   - 更新软件源
//   - 搜索软件包
package bucket
```

### 7.2 函数注释

```go
// NewManager 创建一个新的管理器实例。
//
// 参数：
//   - storage: 存储接口实现
//
// 返回：
//   - Manager: 管理器接口
func NewManager(storage Storage) Manager {
    return &manager{storage: storage}
}
```

### 7.3 导出类型注释

```go
// App 表示一个软件包定义。
type App struct {
    Name    string // 名称
    Version string // 版本
    // ...
}
```

---

## 8. 导入规范

### 8.1 导入分组

```go
import (
    // 标准库
    "context"
    "fmt"
    "io"
    "time"

    // 第三方库
    "github.com/pkg/errors"
    "github.com/yuin/gopher-lua"

    // 项目内部包
    "chopsticks/core/store"
    "chopsticks/core/manifest"
)
```

### 8.2 导入别名

```go
// 避免冲突
import (
    lua "github.com/yuin/gopher-lua"  // 重命名为 lua
)

// 简短包名
import (
    l "github.com/yuin/gopher-lua"
)
```

---

## 9. 代码结构

### 9.1 减少嵌套

```go
// Bad: 深嵌套
for _, v := range data {
    if v.Valid {
        if err := process(v); err == nil {
            v.Send()
        }
    }
}

// Good: 早返回
for _, v := range data {
    if !v.Valid {
        continue
    }
    if err := process(v); err != nil {
        return err
    }
    v.Send()
}
```

### 9.2 避免不必要的 Else

```go
// Bad
var a int
if b {
    a = 100
} else {
    a = 10
}

// Good
a := 10
if b {
    a = 100
}
```

### 9.3 Naked Returns

```go
// 小函数可以使用 naked return
func minMax(a, b int) (min, max int) {
    if a < b {
        return a, b
    }
    return b, a
}

// 大函数使用显式 return
func processData(data []byte) (result []byte, err error) {
    // ... 复杂逻辑
    return result, nil
}
```

---

## 10. 测试规范

### 10.1 测试文件命名

```go
app.go       → app_test.go
manager.go    → manager_test.go
```

### 10.2 表驱动测试

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid", "git", "git", false},
        {"empty", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

---

## 11. 静态检查清单

### 11.1 必须通过

- [ ] `gofmt` 格式化
- [ ] `go vet` 无警告
- [ ] `go build` 编译成功

### 11.2 建议检查

- [ ] `golint` 无警告
- [ ] `staticcheck` 无问题
- [ ] 单元测试通过

---

_最后更新：2026-02-28_
_版本：v0.5.0-alpha_
