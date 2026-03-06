# Chopsticks

<p align="center">
  <img src="https://raw.githubusercontent.com/your-repo/chopsticks/main/docs/assets/logo.png" alt="Chopsticks Logo" width="120">
</p>

<p align="center">
  <strong>Windows 包管理器 - 开发者友好的 Scoop 替代品</strong>
</p>

<p align="center">
  <a href="https://github.com/your-repo/chopsticks/releases">
    <img src="https://img.shields.io/badge/version-v0.10.0--alpha-blue?style=flat-square" alt="Version">
  </a>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.25.6-00ADD8?style=flat-square&logo=go" alt="Go Version">
  </a>
  <a href="#">
    <img src="https://img.shields.io/badge/platform-Windows-blue?style=flat-square&logo=windows" alt="Platform">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License">
  </a>
</p>

<p align="center">
  <a href="docs/ROADMAP.md">路线图</a> •
  <a href="docs/ARCHITECTURE.md">架构设计</a> •
  <a href="docs/CHANGELOG.md">变更日志</a> •
  <a href="wiki/user/USAGE.md">使用指南</a> •
  <a href="wiki/developer/DEVELOPER.md">开发文档</a>
</p>

---

## 简介

**Chopsticks（筷子）** 是一个为 Windows 平台设计的现代化命令行包管理器。它受到 [Scoop](https://scoop.sh) 的启发，采用 Git 仓库作为软件源分发机制，并通过 JavaScript/Lua 脚本赋予开发者完全控制软件安装流程的能力。

### 为什么选择 Chopsticks?

- **开发者友好** - 使用 JavaScript/Lua 编写安装脚本，灵活控制安装流程
- **现代化架构** - Go 语言开发，单文件部署，性能优秀
- **双脚本引擎** - 同时支持 JavaScript (ES6+) 和 Lua 脚本
- **自动追踪** - 自动记录系统操作，卸载时智能清理
- **轻量存储** - SQLite 本地数据库，无需额外服务

---

## 核心特性

| 特性 | 描述 |
|------|------|
| 📦 **包管理** | 简洁的命令行界面管理应用安装、卸载、更新 |
| 📚 **Git 软件源** | 通过 Git 仓库分发软件配置，支持版本控制 |
| 🔌 **双脚本引擎** | JavaScript (Goja) + Lua (Gopher-lua) 双引擎支持 |
| 🪝 **生命周期钩子** | 下载前/后、解压前/后、安装前/后等完整钩子 |
| 🔒 **安全验证** | SHA256/MD5 校验和验证，确保软件完整性 |
| 🧹 **智能清理** | 自动追踪系统操作，卸载时精确清理 |
| 💾 **轻量存储** | SQLite 本地数据库，单文件存储 |
| ⚡ **并行处理** | 支持并发下载和软件源更新 |

---

## 快速开始

### 系统要求

- Windows 10/11
- PowerShell 5.1+ 或 CMD
- x64 或 arm64 架构

### 安装

```powershell
# 使用 PowerShell 安装
iwr -useb https://github.com/your-repo/chopsticks/releases/latest/download/install.ps1 | iex

# 或手动下载
iwr -useb https://github.com/your-repo/chopsticks/releases/latest/download/chopsticks.exe -o chopsticks.exe
```

### 基本用法

```powershell
# 添加软件源
chopsticks bucket add main https://github.com/chopsticks/bucket-main

# 搜索应用
chopsticks search nodejs

# 安装应用
chopsticks install nodejs

# 列出已安装应用
chopsticks list

# 更新应用
chopsticks update nodejs

# 卸载应用
chopsticks uninstall nodejs

# 查看帮助
chopsticks --help
```

---

## 文档

### 用户文档

- [使用指南](wiki/user/USAGE.md) - 详细的命令使用说明
- [常见问题](wiki/user/faq.md) - 常见问题解答

### 开发者文档

- [架构设计](docs/ARCHITECTURE.md) - 系统架构和技术设计
- [开发指南](wiki/developer/DEVELOPER.md) - 如何参与开发
- [API 参考](wiki/developer/API.md) - JavaScript/Lua API 文档
- [编码规范](wiki/developer/STYLE.md) - Go 代码规范
- [Bucket 创建指南](wiki/developer/bucket-guide.md) - 创建软件源指南
- [应用编写指南](wiki/developer/app-best-practices.md) - 编写应用脚本最佳实践

### 项目文档

- [路线图](docs/ROADMAP.md) - 项目规划和迭代计划
- [变更日志](docs/CHANGELOG.md) - 版本变更记录
- [功能需求](wiki/design/REQUIREMENT.md) - 详细功能规格
- [数据库设计](wiki/design/DATABASE.md) - 数据库 Schema 设计

---

## 项目结构

```
chopsticks/
├── cmd/chopsticks/          # CLI 入口点
│   ├── main.go              # 程序主入口
│   └── cli/                 # CLI 命令实现
│       ├── root.go          # 命令路由
│       ├── serve.go         # install 命令
│       ├── clear.go         # uninstall 命令
│       ├── refresh.go       # update 命令
│       ├── search.go        # search 命令
│       ├── list.go          # list 命令
│       └── bucket.go        # bucket 命令
├── core/                    # 核心业务逻辑
│   ├── app/                 # 应用管理
│   ├── bucket/              # 软件源管理
│   ├── manifest/            # 数据结构定义
│   └── store/               # 数据存储 (SQLite/BoltDB)
├── engine/                  # 脚本引擎和 API 模块
│   ├── js.go                # JavaScript 引擎 (goja)
│   ├── lua.go               # Lua 引擎 (gopher-lua)
│   ├── fsutil/              # 文件系统模块
│   ├── fetch/               # HTTP 请求模块
│   ├── execx/               # 命令执行模块
│   ├── archive/             # 压缩解压模块
│   ├── checksum/            # 校验和模块
│   ├── pathx/               # 路径模块
│   ├── logx/                # 日志模块
│   ├── jsonx/               # JSON 模块
│   ├── symlink/             # 符号链接模块
│   ├── registry/            # 注册表模块
│   ├── semver/              # 版本比较模块
│   └── chopsticksx/         # 系统 API 模块
├── infra/                   # 基础设施
│   ├── git/                 # Git 操作
│   └── installer/           # 安装程序处理
├── docs/                    # 项目文档
│   ├── ROADMAP.md           # 路线图
│   ├── ARCHITECTURE.md      # 架构设计
│   └── CHANGELOG.md         # 变更日志
├── wiki/                    # Wiki 文档
│   ├── design/              # 设计文档
│   ├── developer/           # 开发者文档
│   └── user/                # 用户文档
├── go.mod                   # Go 模块定义
├── go.sum                   # Go 依赖校验
└── README.md                # 本文件
```

---

## 技术栈

| 技术 | 用途 | 版本 |
|------|------|------|
| Go | 主开发语言 | 1.25.6 |
| Goja | JavaScript 引擎 | v0.0.0-20260106131823 |
| Gopher-lua | Lua 引擎 | v1.1.1 |
| go-git | Git 操作 | v5.11.0 |
| SQLite | 本地数据库 | v1.14.24 |

---

## 开发

### 构建

```bash
# 克隆仓库
git clone https://github.com/your-repo/chopsticks.git
cd chopsticks

# 安装依赖
go mod download

# 构建
go build -o chopsticks.exe ./cmd/chopsticks

# 运行
./chopsticks.exe --help
```

### 测试

```bash
# 运行测试
go test ./...

# 运行测试并生成覆盖率报告
go test -cover ./...
```

### 贡献

我们欢迎所有形式的贡献！请查看 [贡献指南](CONTRIBUTING.md) 了解如何参与。

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

---

## 路线图

查看 [ROADMAP.md](docs/ROADMAP.md) 了解项目的详细规划和迭代计划。

### 当前进度

- ✅ **v0.1.0-alpha** (2026-02-14) - 基础架构完成
- ✅ **v0.2.0-alpha** (2026-02-27) - 引擎 API 完善
- ✅ **v0.3.0-alpha** (2026-02-27) - 核心功能实现
- ⏳ **v1.0.0** (2026-04-25) - 正式版本发布

---

## 社区

- [GitHub Issues](https://github.com/your-repo/chopsticks/issues) - 问题反馈
- [GitHub Discussions](https://github.com/your-repo/chopsticks/discussions) - 讨论交流

---

## 许可证

[MIT](LICENSE) © Chopsticks Contributors

---

<p align="center">
  Made with ❤️ by the Chopsticks Team
</p>
