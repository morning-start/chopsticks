# 变更日志 (Changelog)

> 所有 notable 变更都将记录在此文件中。
>
> 格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
> 版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

---

## [0.6.0-alpha] - 2026-02-28

### Added

- **日志持久化**: 使用 slog + lumberjack 实现
  - 结构化日志记录（JSON 格式）
  - 日志文件自动轮转（按大小和时间）
  - 可配置的日志级别和输出路径
  - 支持同时输出到文件和控制台
- **批量操作**: 多应用同时安装/卸载/更新
  - 支持多个应用名作为参数
  - 并发执行支持
  - 批量进度显示
  - 部分失败处理机制
- **配置管理**: 新增 `config` 子命令
  - `config get` - 获取配置项
  - `config set` - 设置配置项
  - `config list` - 列出所有配置
  - 支持全局和本地配置
- **依赖解析**: 自动安装应用依赖
  - 解析应用依赖树
  - 自动安装缺失依赖
  - 循环依赖检测
  - 依赖版本约束支持

---

## [0.5.0-alpha] - 2026-02-27

### Added

- **CLI 框架重构**: 使用 urfave/cli/v2 替代 cobra
  - 声明式命令定义
  - 自动帮助生成
  - 类型安全的 Flag 解析
  - 优雅的子命令支持
  - Shell 自动补全功能
- **进度显示功能**: 使用 mpb/v8 实现
  - 下载进度显示（带速度、剩余时间）
  - 安装多阶段进度显示
  - 并发操作支持
- **彩色输出功能**: 使用 fatih/color 实现
  - 标准颜色主题（成功、错误、警告、信息）
  - 自动检测终端颜色能力
  - 支持 --no-color 选项和 NO_COLOR 环境变量
  - 带图标的输出（✓ ✗ ⚠ ℹ →）
- **搜索功能**: 实现真实数据搜索
  - 跨软件源应用搜索
  - 支持模糊匹配
- **Bucket 管理**: 实现 bucket 子命令
  - 软件源添加、删除、列表
  - Git 克隆和更新

### Changed

- **CLI 命令重构**: 命令文件重命名
  - `serve.go` → `install.go`
  - `clear.go` → `uninstall.go`
  - `refresh.go` → `update.go`
  - `root.go` + `commands.go` → `app.go`
- **数据库迁移**: 替换 go-sqlite3 为 modernc.org/sqlite
  - 纯 Go 实现，无需 CGO
  - 更好的跨平台支持

---

## [0.4.0-alpha] - 2026-02-27

### Added

- **Shell 自动补全**: 支持 Bash、Zsh、PowerShell、Fish
- **并行处理**: 多文件并发下载和软件源并行更新
- **结构化错误处理**: 新增 `pkg/errors` 包
  - 基础错误类型 (ErrNotFound, ErrAlreadyExists 等)
  - 应用相关错误 (ErrAppNotFound, ErrInstallFailed 等)
  - 软件源相关错误 (ErrBucketNotFound, ErrBucketLoadFailed 等)
  - 错误分类 `ErrorKind` (KindNotFound, KindIO, KindNetwork 等)
- **单元测试**: 核心模块测试覆盖
  - `pkg/errors` - 错误类型和包装功能
  - `pkg/config` - 配置加载和验证
  - `pkg/parallel` - 并行任务处理
  - `core/store` - 数据库操作
  - `core/app` - 应用管理逻辑
  - `core/bucket` - 软件源管理
  - `engine/*` - 各引擎 API 模块
- **ChopsticksX 模块**: 系统级 API 支持
  - 获取安装目录、配置目录、缓存目录
  - 环境变量管理
  - 创建 shim 快捷方式

### Changed

- **代码重构**: 统一文件命名规范
  - `js.go` → `js_engine.go`
  - `lua.go` → `lua_engine.go`
  - `script.go` → `script_executor.go`
  - `execx/execx.go` → `execx/exec.go`
  - `jsonx/jsonx.go` → `jsonx/json.go`
  - `logx/logx.go` → `logx/log.go`
  - `pathx/pathx.go` → `pathx/path.go`
  - `chopsticksx/chopsticksx.go` → `chopsticksx/chopsticks.go`
- **引擎 API 统一**: 所有模块使用一致的注册和调用方式

### Documentation

- 更新架构设计文档
- 更新 API 参考文档

---

## [0.3.0-alpha] - 2026-02-27

### Added

- **应用生命周期管理**:
  - 完整的安装流程（下载、校验、解压、钩子执行）
  - 完整的卸载流程（钩子执行、目录清理）
  - 完整的更新流程（备份、更新、迁移）
- **Bucket 管理**: 软件源基础管理
- **Git 集成**: 仓库克隆和拉取更新
- **列表功能**: 已安装应用列表
- **JavaScript require**: 实现 JS 引擎的 require 函数
- **7z 解压支持**: archive 模块支持 7z 格式

### Changed

- **存储层重构**: 从 BoltDB 迁移到 SQLite
  - 更好的查询能力
  - 更简单的数据模型
- 改进应用管理器实现

---

## [0.2.0-alpha] - 2026-02-27

### Added

- **JavaScript 引擎**: 基于 Goja 的完整 JavaScript 运行时支持
- **Lua 引擎**: 基于 Gopher-lua 的 Lua 运行时支持
- **文件系统模块 (fsutil)**: 文件读写、目录操作 API
- **HTTP 模块 (fetch)**: HTTP 请求、文件下载 API
- **命令执行模块 (execx)**: 执行系统命令和 PowerShell 脚本
- **压缩解压模块 (archive)**: 支持 zip/7z/tar/tar.gz/tar.xz 格式
- **校验和模块 (checksum)**: SHA256/MD5 计算和验证
- **路径模块 (pathx)**: 路径操作和转换
- **日志模块 (logx)**: 分级日志记录
- **JSON 模块 (jsonx)**: JSON 解析和序列化
- **符号链接模块 (symlink)**: 创建符号链接、硬链接、目录联接
- **注册表模块 (registry)**: Windows 注册表读写操作
- **版本模块 (semver)**: 语义化版本比较
- **系统 API 模块 (chopsticksx)**: 获取安装目录、创建 shim 等
- **CLI 命令框架**: 基于 cobra 的命令行框架
- **数据存储层**: SQLite 存储支持
- **Bucket 模板**: JavaScript 和 Lua 两种模板

### Changed

- 重构项目目录结构，采用更清晰的分层架构
- 更新 Go 版本要求到 1.25.6

### Documentation

- 添加详细的功能需求规格文档
- 添加数据库设计文档
- 添加开发者指南和 API 参考
- 添加用户使用指南

---

## [0.1.0-alpha] - 2026-02-14

### Added

- 项目初始化
- 基础架构搭建
- CLI 框架集成 (cobra)
- 核心接口定义 (Manager, Installer, Storage, Engine)
- 数据结构设计 (App, Bucket, InstalledApp)
- 基础文档结构

---

## 版本说明

### 版本号格式

版本号格式：`主版本号.次版本号.修订号`

- **主版本号 (MAJOR)**: 不兼容的 API 修改
- **次版本号 (MINOR)**: 向下兼容的功能新增
- **修订号 (PATCH)**: 向下兼容的问题修复

### 版本标签

| 标签     | 含义                       |
| -------- | -------------------------- |
| `-alpha` | 内部测试版本，功能不完整   |
| `-beta`  | 公开测试版本，功能基本完整 |
| `-rc`    | 发布候选版本，准备正式发布 |

### 变更类型

| 类型         | 说明           |
| ------------ | -------------- |
| `Added`      | 新功能         |
| `Changed`    | 现有功能的变更 |
| `Deprecated` | 即将移除的功能 |
| `Removed`    | 移除的功能     |
| `Fixed`      | 问题修复       |
| `Security`   | 安全相关的修复 |

---

## 发布计划

| 版本        | 预计日期   | 状态      | 主要特性      |
| ----------- | ---------- | --------- | ------------- |
| 0.1.0-alpha | 2026-02-14 | ✅ 已发布 | 基础架构      |
| 0.2.0-alpha | 2026-02-27 | ✅ 已发布 | 引擎 API 完善 |
| 0.3.0-alpha | 2026-03-14 | ✅ 已发布 | 核心功能实现  |
| 0.4.0-alpha | 2026-02-27 | ✅ 已发布 | 质量提升      |
| 0.5.0-alpha | 2026-02-27 | ✅ 已发布 | 体验优化      |
| 0.6.0-alpha | 2026-02-28 | ✅ 已发布 | 功能完善      |
| 0.7.0-beta  | 2026-03-14 | ⏳ 计划中 | 稳定化        |
| 1.0.0       | 2026-04-11 | ⏳ 计划中 | 正式版本      |

---

_最后更新: 2026-02-28_
