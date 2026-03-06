# Chopsticks 版本号处理设计

> 版本: v1.0.0  
> 最后更新: 2026-03-06

> 详细设计文档，描述版本号处理的完整实现方案

---

## 1. 设计概述

### 1.1 问题背景

Windows 软件的版本号格式极其混乱：

| 软件 | 版本号示例 | 格式 |
|------|-----------|------|
| Chrome | `133.0.6943.98` | 四段式 |
| 7-Zip | `24.09` | CalVer（年月） |
| Python | `3.13.2` | 标准 SemVer |
| Windows | `10.0.26100.3194` | 四段式 |
| 内部版本 | `build 12345` | 纯构建号 |

### 1.2 设计目标

| 目标 | 说明 |
|------|------|
| **宽容解析** | 接受各种格式，不强求标准 |
| **最佳努力比较** | 无法解析时回退到字符串比较 |
| **用户可控** | 允许脚本作者指定比较策略 |
| **缓存优化** | 解析结果缓存，避免重复计算 |

### 1.3 核心原则

```
输入版本号 → 规范化 → 类型检测 → 解析 → 比较
                ↓           ↓         ↓
            清洗输入    自动识别    分层比较
```

---

## 2. 版本号分类

### 2.1 五大类型

```
┌─────────────────────────────────────────────────────────────┐
│                    版本号分类处理流程                         │
├─────────────────────────────────────────────────────────────┤
│  输入: "v1.2.3-beta4"                                       │
│    ↓                                                        │
│  1. 规范化 → "1.2.3-beta4"                                  │
│    ↓                                                        │
│  2. 类型检测 → SemVer 类型                                   │
│    ↓                                                        │
│  3. 解析 → Version{Segments:[1,2,3], Prerelease:"beta", ...} │
│    ↓                                                        │
│  4. 比较 → 数字段 → 预发布 → 构建元数据                       │
└─────────────────────────────────────────────────────────────┘
```

| 类型 | 说明 | 示例 |
|------|------|------|
| **semver** | 语义化版本 | `1.2.3`, `1.2.3-beta4` |
| **calver** | 日历版本 | `2024.03.01`, `24.09` |
| **quad** | 四段式版本 | `10.0.26100.3194` |
| **build** | 纯构建号 | `build 12345`, `r456` |
| **custom** | 自定义格式 | 回退到字符串比较 |

### 2.2 类型检测规则

```go
// 类型检测器（按优先级排序）
var typeDetectors = []struct {
    Type    string
    Pattern *regexp.Regexp
}{
    // 1. 纯构建号: build 123, r456, rev789
    {"build", regexp.MustCompile(`^(?i)(build|b|r|rev)[\s.-]*(\d+)$`)},
    
    // 2. 四段式: 10.0.26100.3194
    {"quad", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)},
    
    // 3. CalVer: 2024.03.01 或 24.03
    {"calver", regexp.MustCompile(`^(\d{2}|\d{4})\.\d{2}(\.\d{2})?$`)},
    
    // 4. SemVer: 1.2.3 或 1.2.3-beta4
    {"semver", regexp.MustCompile(`^\d+(\.\d+)*([+-].+)?$`)},
    
    // 5. 其他: 回退到 custom
    {"custom", regexp.MustCompile(`.*`)},
}
```

---

## 3. 规范化预处理

### 3.1 规范化规则

| 原始输入 | 规范化后 | 处理 |
|----------|----------|------|
| `v1.2.3` | `1.2.3` | 去除 v/V 前缀 |
| `V1.2.3` | `1.2.3` | 同上 |
| `1.2.3_beta` | `1.2.3-beta` | 统一分隔符为 `-` |
| `1.2.3p4` | `1.2.3-patch4` | 展开缩写 |
| `  1.2.3  ` | `1.2.3` | 去除空格 |
| `build-123` | `build-123` | 保留特殊格式 |

### 3.2 规范化算法

```go
func Normalize(version string) string {
    v := strings.TrimSpace(version)
    
    // 去除前缀
    v = strings.TrimPrefix(v, "v")
    v = strings.TrimPrefix(v, "V")
    
    // 替换分隔符
    v = strings.ReplaceAll(v, "_", "-")
    
    // 展开缩写: 1.2.3p1 -> 1.2.3-patch1
    v = regexp.MustCompile(`(\d)p(\d+)$`).ReplaceAllString(v, "$1-patch$2")
    
    return strings.ToLower(v)
}
```

---

## 4. 版本号结构

### 4.1 Version 结构体

```go
type Version struct {
    Raw          string   // 原始字符串: "1.2.3-beta4+build.567"
    Normalized   string   // 规范化后: "1.2.3-beta4"
    Type         string   // 类型: "semver"
    
    // 核心字段
    Segments     []int    // 数字段: [1, 2, 3]
    
    // 预发布版本
    Prerelease   string   // 标识: "beta"
    PrereleaseNum int     // 编号: 4
    
    // 构建元数据
    Build        string   // "build.567"
    
    // 状态
    Comparable   bool     // 是否可比较
}
```

### 4.2 JSON 表示

```json
{
  "version": {
    "raw": "1.2.3-beta4+build.567",
    "normalized": "1.2.3-beta4",
    "type": "semver",
    "segments": [1, 2, 3],
    "prerelease": {
      "tag": "beta",
      "number": 4
    },
    "build": "build.567",
    "comparable": true
  }
}
```

---

## 5. 比较算法

### 5.1 标准比较流程

```
比较 v1 和 v2:
  1. 如果类型不同且都不是 custom:
     - 优先比较数字段多的（更精确）
     - 否则按类型优先级: semver > calver > quad > build
  
  2. 同类型比较:
     a. 逐段比较数字部分
     b. 比较预发布版本（如果有）
     c. 构建元数据不参与比较
  
  3. 任一版本为 custom:
     - 回退到字符串比较
```

### 5.2 数字段比较（通用）

```go
func compareSegments(s1, s2 []int) int {
    maxLen := max(len(s1), len(s2))
    for i := 0; i < maxLen; i++ {
        n1 := getOrZero(s1, i)
        n2 := getOrZero(s2, i)
        if n1 != n2 {
            return n1 - n2
        }
    }
    return 0
}
```

### 5.3 预发布版本比较

**预发布优先级表**：

| 标识符 | 优先级 | 说明 |
|--------|--------|------|
| `dev` | 0 | 开发版 |
| `snapshot` | 0 | 快照版 |
| `alpha` / `a` | 1 | 内测版 |
| `beta` / `b` | 2 | 公测版 |
| `preview` / `pre` | 3 | 预览版 |
| `rc` | 4 | 候选版 |
| (无) | 5 | 正式版 |

**比较逻辑**：

```go
var prereleasePriority = map[string]int{
    "dev":       0,
    "snapshot": 0,
    "alpha":     1,
    "a":         1,
    "beta":      2,
    "b":         2,
    "preview":   3,
    "pre":       3,
    "rc":        4,
}

func comparePrerelease(v1, v2 *Version) int {
    // 有预发布的版本 < 无预发布的版本
    if v1.Prerelease != "" && v2.Prerelease == "" {
        return -1
    }
    if v1.Prerelease == "" && v2.Prerelease != "" {
        return 1
    }
    
    // 都有预发布版本，比较优先级
    if v1.Prerelease != "" && v2.Prerelease != "" {
        p1 := prereleasePriority[v1.Prerelease]
        p2 := prereleasePriority[v2.Prerelease]
        if p1 != p2 {
            return p1 - p2
        }
        
        // 预发布标识相同，比较数字
        return v1.PrereleaseNum - v2.PrereleaseNum
    }
    
    return 0
}
```

**比较示例**：

```
1.0.0-alpha < 1.0.0-beta < 1.0.0-rc < 1.0.0
1.0.0-alpha1 < 1.0.0-alpha2  （数字比较）
1.0.0-alpha10 > 1.0.0-alpha2  （数字比较，非字符串）
```

---

## 6. 实际案例处理

### 6.1 Chrome 版本（四段式）

```
输入: "133.0.6943.98"
检测: quad 类型
解析: segments = [133, 0, 6943, 98]
比较: 逐段数字比较

133.0.6943.98 > 133.0.6943.97
133.0.6943.98 < 133.0.6944.0
```

### 6.2 混合版本（SemVer + 预发布）

```
输入: "1.2.3-beta4"
检测: semver 类型
解析: 
  - segments = [1, 2, 3]
  - prerelease = "beta"
  - prerelease_num = 4

比较:
  1.2.3-beta4 < 1.2.3-beta5
  1.2.3-beta4 < 1.2.3
  1.2.3-beta4 < 1.2.3-rc1
```

### 6.3 7-Zip（CalVer）

```
输入: "24.09"
检测: calver 类型
解析: segments = [24, 9]

比较:
  24.09 > 23.01  (年份优先)
  24.09 > 24.08  (月份其次)
```

### 6.4 Windows 系统版本

```
输入: "10.0.26100.3194"
检测: quad 类型
解析: segments = [10, 0, 26100, 3194]

比较:
  10.0.26100.3194 > 10.0.26100.3193
  10.0.26100.3194 < 10.0.26101.0
```

### 6.5 纯构建号

```
输入: "build 12345"
检测: build 类型
解析: segments = [0, 0, 0, 12345]

比较:
  build 12345 < build 12346
```

### 6.6 非标准版本（回退）

```
输入: "Mar 15 2024"
检测: custom 类型
处理: 字符串比较 "Mar 15 2024" vs "Feb 28 2024"
提示: 警告用户版本号格式不规范
```

---

## 7. 版本约束

### 7.1 约束语法

在 `manifest.json` 中支持多种约束写法：

```json
{
  "dependencies": {
    "tools": [
      {
        "name": "nodejs",
        "version": "18.x",
        "constraint": "semver"
      },
      {
        "name": "python",
        "version": ">=3.9,<3.13",
        "constraint": "semver"
      },
      {
        "name": "7zip",
        "version": ">=24.0",
        "constraint": "calver"
      },
      {
        "name": "vcredist",
        "version": "14.x",
        "constraint": "loose"
      }
    ]
  }
}
```

### 7.2 约束类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `semver` | 标准语义化版本 | `^1.2.3`, `>=1.0.0,<2.0.0` |
| `calver` | 日历版本 | `>=2024.01`, `2024.x` |
| `loose` | 宽松匹配 | `14.x` 匹配 `14.38.33135` |
| `exact` | 精确匹配 | `1.2.3` |
| `any` | 任意版本 | 已安装即可 |

### 7.3 约束解析

```go
func Satisfies(version *Version, constraint string) bool {
    // 解析约束
    // >=1.0.0,<2.0.0 -> [>=1.0.0, <2.0.0]
    // ^1.2.3 -> >=1.2.3,<2.0.0
    // ~1.2.3 -> >=1.2.3,<1.3.0
    // 18.x -> >=18.0.0,<19.0.0
    
    ranges := parseConstraint(constraint)
    
    for _, r := range ranges {
        if !r.Contains(version) {
            return false
        }
    }
    
    return true
}
```

---

## 8. 版本冲突解决

### 8.1 冲突场景

```bash
$ chopsticks install myapp

检测到版本冲突：
  myapp 需要 python >=3.9,<3.13
  但已安装 python 3.8.10
```

### 8.2 解决方案

```bash
解决方案：
  [1] 升级 python 到 3.12.x（推荐）
  [2] 隔离安装 myapp（独立 python 环境）
  [3] 强制安装（可能导致 myapp 无法正常工作）
  [4] 取消安装

请选择 [1/2/3/4]: 
```

### 8.3 隔离安装

隔离安装时，依赖安装到软件私有目录：

```
apps/
└── myapp/
    └── 1.0.0/
        ├── bin/
        └── deps/              # 隔离依赖
            └── python/
                └── 3.12.0/
```

---

## 9. 实现架构

### 9.1 核心接口

```go
package version

// Parser 版本号解析器接口
type Parser interface {
    Parse(string) (*Version, error)
    Compare(*Version, *Version) int
    Satisfies(*Version, string) bool
}

// 内置解析器
var Parsers = map[string]Parser{
    "semver": &SemverParser{},
    "calver": &CalverParser{},
    "quad":   &QuadParser{},
    "build":  &BuildParser{},
    "custom": &CustomParser{},
}

// SmartParse 智能解析（自动检测类型）
func SmartParse(v string) (*Version, error) {
    v = Normalize(v)
    vType := DetectType(v)
    return Parsers[vType].Parse(v)
}
```

### 9.2 解析器实现

```go
// SemverParser 语义化版本解析器
type SemverParser struct{}

func (p *SemverParser) Parse(v string) (*Version, error) {
    // 正则解析: 1.2.3-beta4+build.567
    re := regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([a-zA-Z]+)(\d+)?)?(?:\+(.+))?$`)
    matches := re.FindStringSubmatch(v)
    
    // 构建 Version 结构体
    version := &Version{
        Raw: v,
        Type: "semver",
        // ...
    }
    
    return version, nil
}

func (p *SemverParser) Compare(v1, v2 *Version) int {
    // 1. 比较数字段
    if r := compareSegments(v1.Segments, v2.Segments); r != 0 {
        return r
    }
    
    // 2. 比较预发布版本
    return comparePrerelease(v1, v2)
}
```

### 9.3 缓存机制

```go
// 版本号解析缓存（LRU）
var parseCache = lru.New(1000)

func ParseWithCache(v string) (*Version, error) {
    if cached, ok := parseCache.Get(v); ok {
        return cached.(*Version), nil
    }
    
    version, err := SmartParse(v)
    if err != nil {
        return nil, err
    }
    
    parseCache.Add(v, version)
    return version, nil
}
```

---

## 10. 参考库

| 语言 | 库名 | 用途 |
|------|------|------|
| Go | `github.com/Masterminds/semver/v3` | SemVer 完整实现 |
| Go | `golang.org/x/mod/semver` | Go 官方 SemVer |
| Go | `github.com/hashicorp/go-version` | 通用版本比较 |
| Python | `packaging.version` | Python 标准版本解析 |
| Node.js | `semver` | npm 使用的版本库 |

---

## 11. 测试用例

### 11.1 比较测试

```go
var compareTests = []struct {
    v1       string
    v2       string
    expected int // -1: v1<v2, 0: v1==v2, 1: v1>v2
}{
    // 标准 SemVer
    {"1.0.0", "1.0.1", -1},
    {"1.0.0", "1.0.0", 0},
    {"1.1.0", "1.0.0", 1},
    
    // 预发布版本
    {"1.0.0-alpha", "1.0.0", -1},
    {"1.0.0-alpha", "1.0.0-beta", -1},
    {"1.0.0-beta", "1.0.0-rc", -1},
    {"1.0.0-rc", "1.0.0", -1},
    
    // 预发布版本号
    {"1.0.0-alpha1", "1.0.0-alpha2", -1},
    {"1.0.0-alpha10", "1.0.0-alpha2", 1}, // 数字比较
    
    // 四段式版本
    {"1.2.3.4", "1.2.3.5", -1},
    {"10.0.26100.3194", "10.0.26100.3195", -1},
    
    // CalVer
    {"2024.03.01", "2024.02.15", 1},
    {"24.03", "23.12", 1},
    
    // 前缀
    {"v1.2.3", "1.2.3", 0},
    {"V1.2.3", "v1.2.3", 0},
    
    // 特殊格式
    {"1.2.3p1", "1.2.3", 1},
    {"build 123", "build 124", -1},
}
```

### 11.2 约束测试

```go
var constraintTests = []struct {
    version    string
    constraint string
    expected   bool
}{
    {"1.2.3", "^1.0.0", true},
    {"1.2.3", ">=1.0.0,<2.0.0", true},
    {"2.0.0", "^1.0.0", false},
    {"1.2.3", "~1.2.0", true},
    {"1.3.0", "~1.2.0", false},
    {"18.5.0", "18.x", true},
    {"19.0.0", "18.x", false},
}
```

---

## 12. 总结

本方案的核心优势：

1. **宽容性**：接受各种版本号格式，不强求标准
2. **智能检测**：自动识别版本号类型
3. **分层比较**：数字段 → 预发布 → 构建元数据
4. **可扩展**：新增类型只需添加检测器和解析器
5. **性能**：缓存机制避免重复解析

对于 Windows 软件这种版本号混乱的场景，这种**"尽最大努力解析，无法解析时优雅降级"**的策略是最实用的。

---

## 13. 相关文档

- [REQUIREMENT.md](REQUIREMENT.md) - 功能需求规格
- [DEPENDENCY.md](DEPENDENCY.md) - 依赖管理设计
- [DATABASE.md](DATABASE.md) - 数据存储设计

---

_最后更新：2026-03-06_  
_版本：v1.0.0_
