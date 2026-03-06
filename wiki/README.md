# 🥢 Chopsticks

[![Version](https://img.shields.io/badge/version-v0.10.0--alpha-blue)]()

> 基于 JavaScript 的 Windows 包管理器，高性能、可扩展

---

## 📖 文档导航

### 核心文档

| 文档                            | 说明               | 目标读者       |
| ------------------------------- | ------------------ | -------------- |
| [README](README.md)             | 项目概述与快速开始 | 所有人         |
| [CHANGELOG](CHANGELOG.md)       | 变更日志           | 所有人         |
| [ROADMAP](ROADMAP.md)           | 路线图             | 所有人         |
| [ARCHITECTURE](ARCHITECTURE.md) | 系统架构设计       | 架构师、开发者 |

### 用户文档

| 文档                                 | 说明           | 目标读者 |
| ------------------------------------ | -------------- | -------- |
| [USAGE](user/USAGE.md)               | 用户使用指南   | 用户     |
| [FAQ](user/faq.md)                   | 常见问题解答   | 用户     |
| [缓存管理](user/cache-management.md) | 缓存机制与管理 | 用户     |
| [错误码](user/error-codes.md)        | 错误码参考     | 用户     |

### 开发者文档

| 文档                                            | 说明                | 目标读者 |
| ----------------------------------------------- | ------------------- | -------- |
| [DEVELOPER](developer/DEVELOPER.md)             | 开发者指南          | 开发者   |
| [API](developer/API.md)                         | JavaScript API 参考 | 开发者   |
| [STYLE](developer/STYLE.md)                     | Go 编码规范         | 开发者   |
| [Bucket 创建指南](developer/bucket-guide.md)    | 软件源创建完整教程  | 开发者   |
| [App 最佳实践](developer/app-best-practices.md) | 应用编写指南        | 开发者   |
| [命令别名](developer/command-aliases.md)        | 命令别名配置        | 开发者   |

### 设计文档

| 文档                                                           | 说明              | 目标读者         |
| -------------------------------------------------------------- | ----------------- | ---------------- |
| [REQUIREMENT](design/REQUIREMENT.md)                           | 功能需求规格      | 产品经理、开发者 |
| [DATABASE](design/DATABASE.md)                                 | 数据存储设计      | 架构师、开发者   |
| [DEPENDENCY](design/DEPENDENCY.md)                              | 依赖管理设计      | 架构师、开发者   |
| [VERSION](design/VERSION.md)                                    | 版本号处理设计    | 架构师、开发者   |
| [STATE](design/STATE.md)                                       | 状态管理设计      | 架构师、开发者   |
| [BUCKET-SCAFFOLD](design/BUCKET-SCAFFOLD.md)                   | Bucket 脚手架设计 | 开发者           |
| [PERFORMANCE-OPTIMIZATION](design/PERFORMANCE-OPTIMIZATION.md) | 性能优化设计      | 架构师、开发者   |

---

## 🥢 项目简介

Chopsticks（筷子）是一个基于 JavaScript 的 **Windows 包管理器**，采用 Go 语言开发，使用 JavaScript 作为包定义脚本语言，提供高性能、可扩展的软件包管理能力。

### 核心特性

- ⚡ **快速** - 轻量级，无需管理员权限
- 🔧 **灵活** - JavaScript 脚本定义包行为
- 📦 **开放** - 支持自定义软件源 (bucket)
- 🖥️ **优雅** - 命令行自动补全
- 🔄 **JavaScript 引擎** - 基于 Goja 的高性能脚本引擎
- 🗄️ **纯文件系统存储** - 人类可读的 JSON 格式，易于调试和备份
- 🚀 **高性能** - 并行处理，批量操作性能提升 5-6 倍
- 📊 **可观测** - 实时性能监控和诊断工具
- 🔗 **智能依赖管理** - 分类管理（runtime、tools、libraries）、引用计数、反向依赖计算
- 🏷️ **智能版本处理** - 支持多种版本号格式（SemVer、CalVer、四段式等）

### 术语对照表

| 用户友好术语     | 英文      | 说明         |
| ---------------- | --------- | ------------ |
| 软件源 (Bucket)  | Bucket    | 软件包的集合 |
| 软件包 (App)     | App       | 单个软件定义 |
| 安装 (Install)   | Install   | 部署软件     |
| 卸载 (Uninstall) | Uninstall | 移除软件     |
| 更新 (Update)    | Update    | 升级软件     |

---

## 🚀 Quick Start

5 分钟上手 Chopsticks：

```powershell
# 1. 安装 Chopsticks
iwr -useb https://get.chopsticks.dev/install.ps1 | iex

# 2. 添加官方软件源
chopsticks bucket add main https://github.com/chopsticks-bucket/main

# 3. 安装 Git
chopsticks install git

# 4. 验证安装
git --version
```

### 常用命令
```bash
# 安装软件
chopsticks install git
chopsticks i git

# 卸载软件
chopsticks uninstall git
chopsticks rm git

# 更新软件
chopsticks update git
chopsticks up git

# 搜索软件
chopsticks search vscode
chopsticks s vscode

# 管理软件源
chopsticks bucket add main https://github.com/chopsticks-bucket/main
chopsticks bucket list

# 依赖管理
chopsticks deps git --tree              # 查看依赖树
chopsticks deps git --reverse           # 查看反向依赖
chopsticks autoremove                  # 清理孤儿依赖
chopsticks cleanup-runtime             # 清理无用运行时库

# 批量安装（Go 层自动并发调度）
chopsticks install git nodejs python

# 性能监控
chopsticks perf
```

📖 [详细教程](user/USAGE.md) | 🎓 [开发者指南](developer/DEVELOPER.md)

---

## 4. ✨ 主要功能

### 软件包管理

- **安装** - 从软件源安装软件包
- **卸载** - 彻底移除已安装软件
- **更新** - 升级软件到最新版本
- **搜索** - 快速查找可用软件包

### 软件源管理 (Bucket)

- 添加/删除/列出软件源
- 支持自定义软件源
- 自动同步软件源更新

### 智能并发调度

Go 层自动实现智能任务调度和并发控制，批量操作性能提升 5-6 倍：

```bash
chopsticks install git nodejs python vscode
```

### 性能监控

内置 `perf` 命令提供实时性能监控和诊断：

```bash
chopsticks perf
```

### JavaScript API

提供 13 个模块的 JavaScript API，用于定义软件包行为：

- `fsutil` - 文件系统操作
- `fetch` - HTTP 请求
- `execx` - 命令执行
- `archive` - 压缩解压
- `checksum` - 校验和计算
- `pathx` - 路径操作
- `logx` - 日志输出
- `jsonx` - JSON 处理
- `symlink` - 符号链接
- `registry` - 注册表操作
- `semver` - 版本比较
- `chopsticksx` - 系统 API
- `utils` - 通用工具

📚 [API 参考文档](developer/API.md)

---

## 🛠️ 技术栈

| 组件     | 技术选型 | 说明             |
| -------- | -------- | ---------------- |
| 核心语言 | Go       | 高性能、并发支持 |
| 脚本引擎 | Goja     | JavaScript 引擎  |
| 数据库   | SQLite   | 关系型数据库     |
| Git 操作 | go-git   | 纯 Go Git 库     |

---

## 6. 📂 项目结构

```
chopsticks/
├── cmd/chopsticks/          # 程序入口和 CLI 命令
│   ├── main.go              # 主程序入口
│   └── cli/                 # CLI 命令实现
│
├── core/                    # 核心业务逻辑
│   ├── app/                 # 应用管理
│   ├── bucket/              # 软件源管理
│   ├── store/               # 数据存储
│   └── manifest/            # 数据结构定义
│
├── engine/                  # 脚本引擎
│   ├── engine.go            # 引擎接口定义
│   ├── js_engine.go         # JavaScript 引擎 (goja)
│   ├── script_executor.go   # 脚本执行器
│   └── ...                  # API 模块
│
├── infra/                   # 基础设施
│   ├── git/                 # Git 操作
│   └── installer/           # 安装程序处理
│
├── pkg/                     # 公共包
│   ├── config/              # 配置管理
│   ├── errors/              # 错误处理
│   ├── output/              # 输出和进度
│   └── parallel/            # 并行处理
│
├── wiki/                    # 文档中心
└── bin/                     # 编译输出
```

---

## 🔗 相关链接

- [GitHub 仓库](https://github.com/chopsticks-bucket/main)

---

_最后更新：2026-03-06_
