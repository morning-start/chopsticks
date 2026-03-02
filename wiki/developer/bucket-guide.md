# 软件源创建指南

> 完整的软件源创建教程

---

## 1. 什么是软件源

软件源（Bucket）是 Chopsticks 中的软件源概念，用于组织和存储多个应用（软件包）。

### 1.1 核心术语

| 术语        | 说明                           |
| ----------- | ------------------------------ |
| 软件源 (Bucket)    | 软件源的内部名称               |
| 软件源      | 用户友好术语，软件源的别名         |
| 应用 (App) | 单个软件包定义                 |
| 脚本文件    | 定义安装逻辑的 JavaScript 文件 |

---

## 2. 软件源目录结构

一个标准的软件源目录结构如下：

```
my-bucket/
├── bucket.json           # 必需：软件源配置文件
├── README.md            # 可选：说明文档
├── bucket.db               # 可选：应用元数据缓存（SQLite）
├── apps/_chopsticks_.js  # 必需：类型定义（包含 App 基类）
├── apps/_example_.js     # 必需：示例应用
├── apps/_tools_.js       # 可选：共享工具函数
└── apps/              # 必需：应用目录
    └── git.js           # 必需：JS 脚本
```

---

## 3. 创建步骤

### 3.1 创建软件源配置文件 (bucket.json)

`bucket.json` 是软件源的核心配置文件：

```json
{
  "name": "my-bucket",
  "description": "我的软件源",
  "homepage": "https://github.com/username/my-bucket",
  "license": "MIT",
  "author": "Your Name",
  "keywords": ["chopsticks", "bucket"]
}
```

### 3.2 创建应用脚本

在 `apps/` 目录下创建 JavaScript 脚本文件。

**示例：git.js**

```javascript
class GitApp extends App {
  constructor() {
    super({
      name: "git",
      description: "Distributed version control system",
      homepage: "https://git-scm.com/",
      license: "GPL-2.0",
      category: "development",
      tags: ["vcs", "git", "scm"],
    });
  }

  checkVersion() {
    // 获取最新版本（同步方法）
    const response = fetch.get(
      "https://api.github.com/repos/git-for-windows/git/releases/latest",
    );
    const data = JSON.parse(response.body);
    return data.tag_name.replace(/^v/, "");
  }

  getDownloadInfo(version, arch) {
    // 返回下载信息（同步方法）
    const archMap = {
      amd64: "64-bit",
      x86: "32-bit",
    };
    const filename = `PortableGit-${version}-${archMap[arch] || arch}.7z.exe`;
    return {
      url: `https://github.com/git-for-windows/git/releases/download/v${version}.windows.1/${filename}`,
      type: "7z",
    };
  }
}

module.exports = new GitApp();
```

---

## 4. 软件源管理命令

### 4.1 添加软件源

```bash
# 添加远程软件源
chopsticks bucket add my-bucket https://github.com/username/my-bucket
# 添加本地软件源
chopsticks bucket add local-bucket /path/to/local-bucket
```

### 4.2 更新软件源

```bash
# 更新指定软件源
chopsticks bucket update my-bucket
# 更新所有软件源
chopsticks bucket update
```

### 4.3 列出软件源

```bash
# 列出所有软件源
chopsticks bucket list
```

### 4.4 删除软件源

```bash
# 删除软件源
chopsticks bucket remove my-bucket
# 删除并清理数据
chopsticks bucket remove my-bucket --purge
```

---

## 5. 高级配置

### 5.1 分支管理

```bash
# 指定分支添加
chopsticks bucket add dev-bucket https://github.com/username/my-bucket --branch develop
```

### 5.2 克隆深度

```bash
# 浅克隆（加速克隆）
chopsticks bucket add my-bucket https://github.com/username/my-bucket --depth 1
```

---

## 6. 最佳实践

### 6.1 软件源命名规范

- 使用小写字母和连字符
- 避免特殊字符
- 保持简洁明了

### 6.2 组织结构

建议按功能分类组织应用：

```
my-bucket/
├── apps/
│   ├── dev-tools/     # 开发工具
│   │   └── git.js
│   ├── editors/       # 编辑器
│   │   └── vscode.js
│   └── utils/         # 实用工具
│       └── 7zip.js
└── apps/_tools_.js    # 共享工具
```

### 6.3 定期更新

- 定期检查软件版本更新
- 及时更新数据库中的版本信息
- 保持 README 文档最新

---

## 7. 示例：创建第一个软件源

### 步骤 1：初始化目录

```bash
mkdir my-first-bucket
cd my-first-bucket
mkdir apps
```

### 步骤 2：创建 bucket.json

```json
{
  "name": "my-first-bucket",
  "description": "我的第一个 Chopsticks 软件源",
  "homepage": "https://github.com/username/my-first-bucket",
  "keywords": ["chopsticks", "bucket"]
}
```

### 步骤 3：创建应用脚本

参考 App 最佳实践文档。

### 步骤 4：测试

```bash
# 添加并测试
chopsticks bucket add test /path/to/my-first-bucket
chopsticks search <app-name>
```

---

## 8. 相关链接

- [App 最佳实践](app-best-practices.md)
- [常见问题解答](../user/faq.md)
- [API 参考](API.md)

---

_最后更新：2026-03-01_
_版本：v0.10.0-alpha_
