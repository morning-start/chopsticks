# Bucket 脚手架设计文档

> 版本: v0.10.0-alpha  
> 最后更新: 2026-03-01

> Bucket 脚手架工具设计方案，支持快速创建标准 Bucket 目录结构和 JavaScript 开发
>
> 本文档为开发者文档，详细描述脚手架工具的设计与实现

---

## 1. 需求分析

### 1.1 背景

当前 Chopsticks 的 Bucket（软件源）需要手动创建目录结构和配置文件，对开发者不够友好。

### 1.2 用户痛点

- 不知道标准的 Bucket 目录结构
- 手动创建容易遗漏必要文件
- 不熟悉 App 脚本的 API

### 1.3 目标

- 提供 `chopsticks bucket init` 命令创建标准 Bucket
- 支持 JavaScript 脚本
- 自动配置开发环境

---

## 2. 方案设计

### 2.1 脚手架命令

```bash
# 创建新 Bucket（JavaScript 版本）
chopsticks bucket init my-bucket

# 指定目录创建
chopsticks bucket init my-bucket --dir ./buckets
```

### 2.2 生成的目录结构

#### JavaScript 版本

```
my-bucket/
├── bucket.json                 # 配置
├── bucket.db                   # 可选：元数据缓存（SQLite）
├── apps/                       # 应用目录
│   └── example.js              # 示例应用
├── tools.js                    # 共享工具（可选）
└── .gitignore                  # 忽略文件
```

### 2.3 JavaScript JSDoc 类型提示

```javascript
// my-bucket/apps/example.js
/**
 * @typedef {Object} AppMetadata
 * @property {string} name
 * @property {string} [description]
 * @property {string} [homepage]
 * @property {string} [license]
 * @property {string} bucket
 */

/**
 * @typedef {Object} DownloadInfo
 * @property {string} url
 * @property {'zip'|'7z'|'tar'|'tar.gz'|'tar.xz'|'exe'|'msi'} type
 * @property {string} [filename]
 * @property {{type: 'sha256'|'md5', value: string}} [checksum]
 */

/**
 * @typedef {Object} InstallContext
 * @property {string} name
 * @property {string} version
 * @property {'amd64'|'x86'|'arm64'} arch
 * @property {string} cookDir
 * @property {string} bucket
 * @property {string} downloadPath
 */

const { App } = require("@chopsticks/core");

/**
 * @extends {App}
 */
class ExampleApp extends App {
  constructor() {
    super({
      name: "example",
      description: "Example Application",
      homepage: "https://example.com",
      license: "MIT",
      bucket: "my-bucket",
    });
  }

  /**
   * @returns {Promise<string>}
   */
  async checkVersion() {
    return "1.0.0";
  }

  /**
   * @param {string} version
   * @param {string} arch
   * @returns {Promise<DownloadInfo>}
   */
  async getDownloadInfo(version, arch) {
    return {
      url: `https://example.com/download/${version}/app-${arch}.zip`,
      type: "zip",
    };
  }

  /**
   * @param {InstallContext} ctx
   * @returns {Promise<void>}
   */
  async onPostInstall(ctx) {
    log.info("安装完成！");
  }
}

module.exports = new ExampleApp();
```

### 2.4 bucket.json 配置

```json
{
  "name": "my-bucket",
  "description": "My custom Chopsticks Bucket",
  "homepage": "https://github.com/username/my-bucket",
  "license": "MIT",
  "author": "Your Name",
  "keywords": ["chopsticks", "bucket"]
}
```

---

## 3. 实现计划

### 3.1 CLI 命令实现

| 命令                              | 说明              | 优先级 |
| --------------------------------- | ----------------- | ------ |
| `chopsticks bucket init`          | 初始化 Bucket 目录 | P0     |
| `chopsticks bucket create <name>` | 创建单个 App      | P1     |
| `chopsticks bucket validate`      | 验证 Bucket 配置   | P1     |

### 3.2 运行时包

| 包名               | 说明           | 优先级 |
| ------------------ | -------------- | ------ |
| `@chopsticks/core` | JS 运行时核心  | P0     |

---

## 4. 使用流程

### 4.1 创建新 Bucket

```bash
# 1. 创建 Bucket 目录
chopsticks bucket init my-software

# 2. 进入目录
cd my-software

# 3. 安装依赖（如需要）
npm install

# 4. 创建应用
chopsticks bucket create git

# 5. 开发应用
# 编辑 apps/git.js

# 6. 测试
chopsticks install git --bucket my-software
```

### 4.2 开发工作流

```
┌─────────────────────────────────────────────────────────┐
│                     开发流程                             │
├─────────────────────────────────────────────────────────┤
│  1. chopsticks bucket init my-bucket   创建 Bucket      │
│  2. cd my-bucket                       进入目录         │
│  3. chopsticks bucket create app       创建应用         │
│  4. 编辑 apps/app.js                   编写代码         │
│  5. chopsticks install app             测试安装         │
│  6. chopsticks uninstall app           测试卸载         │
│  7. git add . && git commit            提交             │
└─────────────────────────────────────────────────────────┘
```

---

## 5. 附录

### 5.1 JavaScript 应用模板

```javascript
// apps/example.js
const { App } = require("@chopsticks/core");

/**
 * Example App
 *
 * 一个示例应用，展示基本的开发模式
 */
class ExampleApp extends App {
  constructor() {
    super({
      name: "example",
      description: "Example Application",
      homepage: "https://example.com",
      license: "MIT",
      bucket: "my-bucket",
    });
  }

  /**
   * 检查最新版本
   * @returns {Promise<string>}
   */
  async checkVersion() {
    try {
      const response = await fetch.get("https://api.example.com/version");
      const data = json.parse(response.body);
      return data.version;
    } catch (error) {
      log.warn("获取版本失败，使用默认版本");
      return "1.0.0";
    }
  }

  /**
   * 获取下载信息
   * @param {string} version
   * @param {string} arch
   * @returns {Promise<{url: string, type: string, checksum?: {type: string, value: string}}>}
   */
  async getDownloadInfo(version, arch) {
    const archMap = {
      amd64: "x64",
      x86: "x86",
      arm64: "arm64",
    };

    return {
      url: `https://example.com/download/${version}/app-${archMap[arch] || arch}.zip`,
      type: "zip",
      checksum: {
        type: "sha256",
        value: "abc123...",
      },
    };
  }

  /**
   * 安装后处理
   * @param {Object} ctx
   * @returns {Promise<void>}
   */
  async onPostInstall(ctx) {
    log.info(`正在配置 ${this.metadata.name}...`);

    await chopsticks.addToPath(path.join(ctx.cookDir, "bin"));

    await chopsticks.createShortcut({
      source: path.join(ctx.cookDir, "app.exe"),
      name: this.metadata.name,
      description: this.metadata.description,
    });

    log.info(`${this.metadata.name} 安装完成！`);
  }
}

module.exports = new ExampleApp();
```

---

## 6. 更新记录

| 日期       | 版本   | 变更                             |
| ---------- | ------ | -------------------------------- |
| 2026-03-01 | v1.2.0 | 更新版本号至 v0.10.0-alpha       |
| 2026-02-28 | v1.1.0 | 仅保留 JavaScript 支持           |
| 2026-02-25 | v1.0.0 | 初始版本                         |

---

_最后更新：2026-03-01_  
_版本：v0.10.0-alpha_
