# App 最佳实践

> 应用脚本编写指南与最佳实践

---

## 1. 概述

应用（App）是 Chopsticks 中的软件包定义，通过 JavaScript 脚本描述软件的安装、卸载和更新逻辑。

---

## 2. 基本结构

### 2.1 最小示例

```javascript
/** @type {import('./_chopsticks_')} */

class MyApp extends App {
    constructor() {
        super({
            name: "myapp",
            description: "My Application",
            homepage: "https://example.com",
            license: "MIT",
            bucket: "my-bucket",
        });
    }

    async checkVersion() {
        // 获取最新版本号
    }

    async getDownloadInfo(version, arch) {
        // 返回下载 URL 和类型
    }
}

module.exports = new MyApp();
```

---

## 3. 必需方法

### 3.1 checkVersion()

获取软件的最新版本号：

```javascript
async checkVersion() {
    // 方式 1：GitHub Releases
    const response = await fetch.get(
        "https://api.github.com/repos/owner/repo/releases/latest"
    );
    const data = JSON.parse(response.body);
    return data.tag_name.replace(/^v/, "");

    // 方式 2：网页解析
    const response = await fetch.get("https://example.com/download");
    const match = response.body.match(/version[= ](\d+\.\d+\.\d+)/);
    return match ? match[1] : null;

    // 方式 3：JSON API
    const response = await fetch.get(
        "https://example.com/api/version"
    );
    return JSON.parse(response.body).version;
}
```

### 3.2 getDownloadInfo()

返回下载信息：

```javascript
async getDownloadInfo(version, arch) {
    const archMap = {
        amd64: "x64",
        x86: "x86"
    };

    const filename = `myapp-${version}-${archMap[arch] || arch}.zip`;
    return {
        url: `https://example.com/downloads/${filename}`,
        type: "zip",  // zip, 7z, tar.gz, exe, msi
    };
}
```

---

## 4. 生命周期钩子

### 4.1 安装钩子

```javascript
async onPreDownload(ctx) {
    // 下载前执行
    log.info("开始下载...");
}

async onPostDownload(ctx) {
    // 下载完成后执行
    log.info("下载完成");
}

async onPreExtract(ctx) {
    // 解压前执行
}

async onPostExtract(ctx) {
    // 解压后执行
    log.info("解压完成");
}

async onPreInstall(ctx) {
    // 安装前执行
    log.info("开始安装...");
}

async onPostInstall(ctx) {
    // 安装后执行
    log.info("安装完成！");
    // 设置环境变量
    await chopsticks.addToPath(path.join(ctx.cookDir, "bin"));
    // 创建快捷方式
    await chopsticks.createShortcut({
        source: path.join(ctx.cookDir, "app.exe"),
        name: "My App",
        description: "My Application",
    });
}
```

### 4.2 卸载钩子

```javascript
async onPreUninstall(ctx) {
    // 卸载前执行
    log.info("开始卸载...");
}

async onPostUninstall(ctx) {
    // 卸载后执行
    // 清理注册表
    await registry.deleteKey("HKCU\\Software\\MyApp");
}
```

---

## 5. 配置函数

### 5.1 env_path()

返回需要加入 PATH 的目录：

```javascript
env_path() {
    return ["bin", "cmd"];
}
```

### 5.2 bin()

返回可执行文件列表（用于创建 shim）：

```javascript
bin() {
    return ["app.exe", "tool.exe"];
}
```

### 5.3 persist()

返回需要持久化的目录（更新时保留）：

```javascript
persist() {
    return ["config", "data"];
}
```

### 5.4 depends()

声明依赖：

```javascript
depends() {
    return ["nodejs", "python"];
}
```

### 5.5 conflicts()

声明冲突软件：

```javascript
conflicts() {
    return ["git-for-windows"];
}
```

---

## 6. 常见模式

### 6.1 绿色软件模式

适用于解压即用的绿色软件：

```javascript
class App extends App {
    constructor() {
        super({ name: "app", description: "My App" });
    }

    async getDownloadInfo(version, arch) {
        return {
            url: `https://example.com/app-${version}.zip`,
            type: "zip",
        };
    }

    async onPostInstall(ctx) {
        // 创建快捷方式
        await chopsticks.createShortcut({
            source: path.join(ctx.cookDir, "app.exe"),
            name: "My App",
        });
    }
}
```

### 6.2 安装程序模式

适用于需要运行安装程序的软件：

```javascript
class App extends App {
    constructor() {
        super({ name: "app", description: "My App" });
    }

    async getDownloadInfo(version, arch) {
        return {
            url: `https://example.com/app-${version}-setup.exe`,
            type: "installer",
        };
    }

    async onPostExtract(ctx) {
        // 运行安装程序
        const installer = path.join(ctx.cookDir, "app-setup.exe");
        await installer.run(installer, ["/S", "/D=" + ctx.cookDir]);
    }
}
```

### 6.3 多架构支持

```javascript
class App extends App {
    constructor() {
        super({ name: "app", description: "My App" });
    }

    async getDownloadInfo(version, arch) {
        const archMap = {
            amd64: "x64",
            x86: "x86",
            arm64: "arm64"
        };

        return {
            url: `https://example.com/app-${version}-${archMap[arch]}.zip`,
            type: "zip",
        };
    }
}
```

### 6.4 多版本支持

```javascript
class NodeJS extends App {
    constructor() {
        super({ name: "nodejs", description: "Node.js" });
    }

    env_path() {
        return [path.join("node-v" + ctx.version, "bin")];
    }

    bin() {
        return ["node.exe", "npm.exe", "npx.exe"];
    }
}
```

---

## 7. 错误处理

### 7.1 安全调用

```javascript
async checkVersion() {
    const result = await this.safeCall(async () => {
        const response = await fetch.get(url);
        return JSON.parse(response.body).version;
    });

    if (!result.success) {
        log.error("获取版本失败: " + result.error);
        return "fallback-version";
    }

    return result.value;
}
```

### 7.2 重试机制

```javascript
async fetchWithRetry(url, retries = 3) {
    for (let i = 0; i < retries; i++) {
        try {
            const response = await fetch.get(url);
            if (response.ok) return response;
        } catch (e) {
            log.warn(`重试 ${i + 1}/${retries}: ${e.message}`);
            await new Promise(r => setTimeout(r, 1000));
        }
    }
    throw new Error("重试失败");
}
```

---

## 8. 调试技巧

### 8.1 日志输出

```javascript
async onPostInstall(ctx) {
    log.debug("安装目录: " + ctx.cookDir);
    log.info("安装完成");
    log.warn("警告信息");
    log.error("错误信息");
}
```

### 8.2 调试模式

```bash
# 启用详细输出
chopsticks install myapp --verbose

# 启用调试模式
chopsticks install myapp --debug
```

---

## 9. 最佳实践清单

### 9.1 必需项

- [ ] 实现 `checkVersion()` 方法
- [ ] 实现 `getDownloadInfo()` 方法
- [ ] 正确设置 `name` 属性
- [ ] 添加适当的 `description`

### 9.2 推荐项

- [ ] 使用 `onPostInstall` 创建快捷方式
- [ ] 配置 `env_path()` 添加到 PATH
- [ ] 配置 `bin()` 创建 shim
- [ ] 使用 `persist()` 保留用户数据

### 9.3 注意事项

- 始终使用异步操作（async/await）
- 正确处理错误和异常
- 清理安装过程中的临时文件
- 遵循版本号规范（语义化版本）

---

## 10. 完整示例

查看 [API 参考文档](../wiki/API.md) 中的完整示例。

---

## 11. 相关链接

- [软件源创建指南](bucket-guide.md)
- [常见问题解答](faq.md)
- [API 参考](../wiki/API.md)

---

_最后更新：2026-02-26_
