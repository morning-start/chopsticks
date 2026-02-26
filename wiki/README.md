# 🥢 Chopsticks Wiki

> Chopsticks（筷子）- Windows 包管理器知识中心

---

## 📖 文档导航

### 核心文档

| 文档                      | 说明               | 目标读者 |
| ------------------------- | ------------------ | -------- |
| [README](README.md)       | 项目概述与快速开始 | 所有人   |
| [CHANGELOG](CHANGELOG.md) | 变更日志           | 所有人   |
| [ROADMAP](ROADMAP.md)     | 路线图             | 所有人   |

### 用户文档

| 文档                   | 说明         | 目标读者 |
| ---------------------- | ------------ | -------- |
| [USAGE](user/USAGE.md) | 用户使用指南 | 用户     |
| [FAQ](user/faq.md)     | 常见问题解答 | 用户     |

### 开发者文档

| 文档                                            | 说明                | 目标读者 |
| ----------------------------------------------- | ------------------- | -------- |
| [DEVELOPER](developer/DEVELOPER.md)             | 开发者指南          | 开发者   |
| [API](developer/API.md)                         | JavaScript API 参考 | 开发者   |
| [STYLE](developer/STYLE.md)                     | Go 编码规范         | 开发者   |
| [Bucket 创建指南](developer/bucket-guide.md)    | 软件源创建完整教程  | 开发者   |
| [App 最佳实践](developer/app-best-practices.md) | 应用编写指南        | 开发者   |

### 设计文档

| 文档                                         | 说明              | 目标读者         |
| -------------------------------------------- | ----------------- | ---------------- |
| [ARCHITECTURE](design/ARCHITECTURE.md)       | 系统架构设计      | 架构师、开发者   |
| [REQUIREMENT](design/REQUIREMENT.md)         | 功能需求规格      | 产品经理、开发者 |
| [DATABASE](design/DATABASE.md)               | 数据库设计        | 架构师、开发者   |
| [STATE](design/STATE.md)                     | 状态管理设计      | 架构师、开发者   |
| [BUCKET-SCAFFOLD](design/BUCKET-SCAFFOLD.md) | Bucket 脚手架设计 | 开发者           |

---

## 🥢 项目简介

Chopsticks（筷子）是一个受 [Scoop](https://scoop.sh/) 启发的 **Windows 包管理器**，采用 Go 语言开发，使用 JavaScript/Lua 作为包定义脚本语言。

### 核心特性

- ⚡ **快速** - 轻量级，无需管理员权限
- 🔧 **灵活** - JavaScript/Lua 脚本定义包行为
- 📦 **开放** - 支持自定义软件源
- 🖥️ **优雅** - 命令行自动补全

### 术语对照表

| 用户友好术语     | 英文      | 说明         |
| ---------------- | --------- | ------------ |
| 软件源 (Source)  | Bucket    | 软件包的集合 |
| 软件包 (Package) | App       | 单个软件定义 |
| 安装 (Install)   | Install   | 部署软件     |
| 卸载 (Uninstall) | Uninstall | 移除软件     |
| 更新 (Update)    | Update    | 升级软件     |

---

## 🚀 快速开始

### 安装

```powershell
git clone https://github.com/chopsticks-bows/main.git
cd main
go build -o chopsticks.exe ./cmd/chopsticks
```

### 基本命令

```bash
# 安装软件
chopsticks install git

# 卸载软件
chopsticks uninstall git

# 更新软件
chopsticks update --all

# 搜索软件
chopsticks search vscode

# 管理软件源
chopsticks bucket add main https://github.com/chopsticks-bows/main
```

---

## 📂 项目结构

```
chopsticks/
├── cmd/chopsticks/          # 程序入口和 CLI 命令
│   ├── main.go              # 主程序入口
│   ├── install.go           # 安装命令
│   ├── uninstall.go         # 卸载命令
│   ├── update.go            # 更新命令
│   ├── search.go            # 搜索命令
│   ├── list.go              # 列表命令
│   └── bucket.go            # 软件源管理命令
│
├── core/                    # 核心业务逻辑
│   ├── app/                 # 应用管理
│   │   ├── manager.go      # 应用管理器
│   │   ├── install.go      # 安装流程
│   │   ├── uninstall.go    # 卸载流程
│   │   └── updater.go      # 更新流程
│   ├── bucket/              # 软件源管理
│   │   ├── manager.go      # 软件源管理器
│   │   └── bucket.go       # 软件源操作
│   ├── store/               # 数据存储
│   │   ├── storage.go      # 存储接口
│   │   └── sqlite.go       # SQLite 实现
│   └── manifest/            # 数据结构定义
│
├── engine/                  # 脚本引擎
│   ├── engine.go            # 引擎接口
│   ├── js.go               # JavaScript 引擎 (goja)
│   ├── lua.go              # Lua 引擎 (gopher-lua)
│   ├── fsutil/             # 文件系统 API
│   ├── fetch/              # HTTP 请求 API
│   ├── execx/              # 命令执行 API
│   ├── archive/            # 压缩解压 API
│   ├── checksum/           # 校验和 API
│   ├── pathx/              # 路径操作 API
│   ├── logx/               # 日志 API
│   ├── jsonx/              # JSON 处理 API
│   ├── symlink/            # 符号链接 API
│   ├── registry/           # 注册表 API
│   ├── semver/             # 版本比较 API
│   └── chopsticksx/        # 系统 API
│
├── infra/                   # 基础设施
│   ├── git/                # Git 操作 (go-git)
│   └── installer/          # 安装程序处理
│
├── wiki/                   # Wiki 文档
└── bin/                    # 编译输出
```

---

## 📊 技术栈

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| 核心语言 | Go 1.25.6 | 高性能、并发支持 |
| 脚本引擎 | JavaScript (goja) / Lua (gopher-lua) | 双引擎支持 |
| 数据库 | SQLite | 关系型，支持复杂查询 |
| Git | go-git | 纯 Go Git 库 |

---

## 🔗 相关链接

- [GitHub 仓库](https://github.com/chopsticks-bows/main)

---

_最后更新：2026-02-27_
