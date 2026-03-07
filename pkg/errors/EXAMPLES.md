# 错误处理系统使用示例

本文档展示如何使用 Chopsticks 增强的错误处理系统。

## 目录

1. [基础使用](#基础使用)
2. [创建结构化错误](#创建结构化错误)
3. [错误包装](#错误包装)
4. [错误恢复建议](#错误恢复建议)
5. [错误序列化](#错误序列化)
6. [错误格式化输出](#错误格式化输出)

---

## 基础使用

### 使用预定义的错误码

```go
package main

import (
    "fmt"
    "chopsticks/pkg/errors"
)

func main() {
    // 创建软件未找到错误
    err := errors.NewAppNotFound("git")
    fmt.Println(err.Error())
    // 输出：[CHP-4001] 软件未找到：git
    
    // 创建下载失败错误
    err = errors.NewDownloadFailed("https://example.com/file.zip", nil)
    fmt.Println(err.Error())
    // 输出：[CHP-2003] 下载失败：https://example.com/file.zip
}
```

---

## 创建结构化错误

### 使用 NewStructured 创建错误

```go
// 创建基础结构化错误
err := errors.NewStructured(
    errors.ErrAppNotFound,
    "软件未找到：git",
)

// 添加上下文信息
err.WithContext("app_name", "git").
    WithContext("bucket", "main").
    WithOperation("install")

// 设置是否可重试
err.WithRetryable(false)

// 设置是否可恢复
err.WithRecoverable(true)
```

### 使用专用构造函数

```go
// 安装错误
err := errors.NewAppAlreadyInstalled("git", "2.30.0")

// 卸载错误
err = errors.NewUninstallFailed("git", someError)

// 依赖错误
err = errors.NewDependencyConflict("nodejs", "版本不匹配")

// 网络错误
err = errors.NewNetworkConnection("github.com", networkErr)

// 配置错误
err = errors.NewConfigInvalid("config.yaml", "YAML 格式错误")
```

---

## 错误包装

### 包装现有错误

```go
import (
    "net/http"
    "chopsticks/pkg/errors"
)

func downloadFile(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        // 包装为标准结构化错误
        return errors.WrapStructured(
            err,
            errors.ErrDownloadFailed,
            "download package",
        )
    }
    defer resp.Body.Close()
    
    // ... 处理响应
    return nil
}
```

### 使用格式化包装

```go
func installApp(name, version string) error {
    err := doInstall(name, version)
    if err != nil {
        return errors.WrapStructuredf(
            err,
            errors.ErrInstallFailed,
            "install %s version %s",
            name,
            version,
        )
    }
    return nil
}
```

### 错误链

```go
// 创建错误链
cause := errors.New("network timeout")
err1 := errors.WrapStructured(cause, errors.ErrNetworkConnection, "connect")
err2 := errors.WrapStructured(err1, errors.ErrDownloadFailed, "download")

// 可以解包获取根本原因
unwrapped := errors.Unwrap(err2) // 返回 err1
```

---

## 错误恢复建议

### 获取恢复建议

```go
func handleError(err error) {
    suggestions := errors.GetRecoverySuggestions(err)
    
    fmt.Println("错误信息：", err.Error())
    fmt.Println("\n恢复建议:")
    
    for i, sug := range suggestions {
        fmt.Printf("%d. %s\n", i+1, sug.Title)
        fmt.Printf("   %s\n", sug.Description)
        
        if len(sug.Commands) > 0 {
            fmt.Println("   推荐命令:")
            for _, cmd := range sug.Commands {
                fmt.Printf("     > %s\n", cmd)
            }
        }
        
        if sug.AutoFixable {
            fmt.Println("   [可自动修复]")
        }
    }
}
```

### 自定义恢复建议

```go
err := errors.NewStructured(
    errors.ErrDownloadFailed,
    "下载失败",
).WithSuggestion(errors.RecoverySuggestion{
    Title:       "检查网络连接",
    Description: "确保网络正常连接",
    Commands:    []string{"ping github.com"},
    AutoFixable: false,
}).WithSuggestion(errors.RecoverySuggestion{
    Title:       "使用镜像源",
    Description: "尝试使用更快的镜像源",
    Commands:    []string{"chopsticks bucket add mirror https://mirror.example.com"},
    AutoFixable: true,
})
```

---

## 错误序列化

### 转换为 JSON

```go
err := errors.NewAppNotFound("git").
    WithContext("bucket", "main").
    WithOperation("install")

// 序列化为 JSON
jsonStr, err := errors.ToJSON(err)
if err != nil {
    log.Fatal(err)
}
fmt.Println(jsonStr)

// 输出类似：
// {
//   "code": "CHP-4001",
//   "message": "软件未找到：git",
//   "category": "install",
//   "suggestions": [...],
//   "details": {
//     "timestamp": "2026-03-06T00:00:00Z",
//     "operation": "install",
//     "context": {
//       "bucket": "main"
//     }
//   },
//   "recoverable": true,
//   "retryable": false
// }
```

### 从 JSON 解析

```go
jsonData := `{
    "code": "CHP-4001",
    "message": "软件未找到",
    "category": "install"
}`

parsedErr, err := errors.FromJSON([]byte(jsonData))
if err != nil {
    log.Fatal(err)
}

fmt.Println("错误码：", parsedErr.Code)
fmt.Println("错误消息：", parsedErr.Message)
```

---

## 错误格式化输出

### 简洁格式

```go
err := errors.NewAppNotFound("git")
fmt.Println(errors.FormatError(err, false))

// 输出：
// 错误代码：CHP-4001
// 错误分类：install
// 错误消息：软件未找到：git
// 
// 恢复建议:
//   1. 搜索软件
//      使用关键词搜索可用的软件
//      推荐命令:
//        > chopsticks search {keyword}
//   2. 查看所有可用软件
//      列出所有可安装的软件
//      推荐命令:
//        > chopsticks list --all
```

### 详细格式

```go
err := errors.NewAppNotFound("git").
    WithContext("bucket", "main").
    WithOperation("install check")

fmt.Println(errors.FormatError(err, true))

// 输出：
// 错误代码：CHP-4001
// 错误分类：install
// 错误消息：软件未找到：git
// 执行操作：install check
// 上下文信息:
//   - bucket: main
// 
// 恢复建议:
//   ...
```

---

## 错误分析

### 分析多个错误

```go
errors := []error{
    errors.NewAppNotFound("git"),
    errors.NewAppNotFound("nodejs"),
    errors.NewNetworkConnection("github.com", nil),
    errors.NewBucketNotFound("main"),
}

summary := errors.AnalyzeErrors(errors)

fmt.Printf("总错误数：%d\n", summary.TotalErrors)
fmt.Printf("可恢复：%d\n", summary.Recoverable)
fmt.Printf("不可恢复：%d\n", summary.Unrecoverable)

fmt.Println("按分类统计:")
for category, count := range summary.ByCategory {
    fmt.Printf("  %s: %d\n", category, count)
}

fmt.Println("最常见错误:")
for _, code := range summary.MostFrequent {
    fmt.Printf("  %s\n", code)
}
```

---

## 错误码参考

### 1xxx - 系统错误
- `CHP-1001` - 权限不足
- `CHP-1002` - 磁盘空间不足
- `CHP-1003` - 文件不存在
- `CHP-1004` - 文件已存在

### 2xxx - 网络错误
- `CHP-2001` - 网络连接失败
- `CHP-2002` - 下载超时
- `CHP-2003` - 下载失败
- `CHP-2004` - URL 无效

### 3xxx - 软件源错误
- `CHP-3001` - 软件源未找到
- `CHP-3002` - 软件源已存在
- `CHP-3003` - 软件源加载失败
- `CHP-3006` - 清单文件未找到

### 4xxx - 安装错误
- `CHP-4001` - 软件未找到
- `CHP-4002` - 软件已安装
- `CHP-4003` - 软件未安装
- `CHP-4004` - 安装失败
- `CHP-4005` - 卸载失败
- `CHP-4007` - 版本未找到
- `CHP-4009` - 校验和不匹配

### 5xxx - 依赖错误
- `CHP-5001` - 依赖冲突
- `CHP-5002` - 依赖未找到
- `CHP-5003` - 循环依赖
- `CHP-5004` - 依赖版本不匹配

### 6xxx - 配置错误
- `CHP-6001` - 配置文件不存在
- `CHP-6002` - 配置文件无效
- `CHP-6003` - 读取配置失败
- `CHP-6005` - 配置值无效

### 9xxx - 其他错误
- `CHP-9001` - 未知错误
- `CHP-9002` - 内部错误
- `CHP-9003` - 操作已取消
- `CHP-9004` - 操作超时

---

## 最佳实践

1. **使用结构化错误**：始终使用 `NewStructured` 或专用构造函数
2. **添加上下文**：使用 `WithContext` 添加有助于调试的信息
3. **包装而非替换**：使用 `WrapStructured` 包装底层错误
4. **提供恢复建议**：为用户创建有用的恢复建议
5. **标记可重试错误**：网络相关错误应该标记为 `Retryable: true`
6. **使用错误码**：通过错误码便于日志分析和监控

---

_最后更新：2026-03-06_
_版本：v0.11.0_
