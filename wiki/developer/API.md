# Chopsticks JavaScript API 参考

> 应用脚本中可用的 JavaScript API 完整参考（OOP 风格）

---

## 1. 基类与生命周期

### 1.1 App 基类

```javascript
const { App } = require('@chopsticks/core');

class MyApp extends App {
    constructor() {
        super({
            name: "app",
            description: "My Application",
            homepage: "https://...",
            license: "MIT",
            bucket: "my-bucket",
        });
    }

    async checkVersion() { ... }
    async getDownloadInfo(version, arch) { ... }
}
```

### 1.2 生命周期钩子

| 钩子                   | 说明   | 参数  |
| ---------------------- | ------ | ----- |
| `onPreDownload(ctx)`   | 下载前 | `ctx` |
| `onPostDownload(ctx)`  | 下载后 | `ctx` |
| `onPreExtract(ctx)`    | 解压前 | `ctx` |
| `onPostExtract(ctx)`   | 解压后 | `ctx` |
| `onPreInstall(ctx)`    | 安装前 | `ctx` |
| `onPostInstall(ctx)`   | 安装后 | `ctx` |
| `onPreUninstall(ctx)`  | 卸载前 | `ctx` |
| `onPostUninstall(ctx)` | 卸载后 | `ctx` |

### 1.3 上下文对象

```javascript
// ctx 参数包含
{
    name: "git",           // 软件名
    version: "2.43.0",     // 版本
    arch: "amd64",        // 架构
    cookDir: "C:\\...\\apps\\git\\2.43.0",  // 安装目录
    bucket: "main",          // 来源软件源
    downloadPath: "C:\\...\\cache\\git-2.43.0.7z",  // 下载文件
}
```

---

## 2. 日志模块 (log)

```javascript
log.debug("debug message"); // 调试日志
log.info("info message"); // 信息日志
log.warn("warning message"); // 警告日志
log.error("error message"); // 错误日志
```

---

## 3. JSON 模块 (json)

```javascript
// 对象转 JSON
const str = JSON.stringify({ name: "test", version: "1.0.0" });

// JSON 转对象
const obj = JSON.parse('{"name":"test","version":"1.0.0"}');

// 简写
const str = json.encode(obj);
const obj = json.decode(str);
```

---

## 4. 路径模块 (path)

```javascript
// 连接路径
const full = path.join("dir", "subdir", "file.txt");

// 转绝对路径
const abs = path.abs("relative/path");

// 获取目录
const dir = path.dir("/path/to/file.txt");

// 获取文件名
const base = path.base("/path/to/file.txt");

// 获取扩展名
const ext = path.ext("/path/to/file.txt");

// 检查存在
const exists = path.exists("/path/to/file");

// 检查是否为目录
const isDir = path.isDir("/path/to/dir");
```

---

## 5. 执行模块 (exec)

```javascript
// 执行命令
const result = exec.exec("git", "--version");
// result.exitCode, result.stdout, result.stderr, result.success

// 执行 shell 命令
const result = exec.shell("echo hello");

// 执行 PowerShell 命令
const result = exec.powershell("Get-Process");
```

---

## 6. HTTP 模块 (fetch)

```javascript
// GET 请求
const response = fetch.get(url);
// response.status, response.ok, response.body, response.headers

// POST 请求
const response = fetch.post(url, body, "application/json");

// 下载文件
fetch.download(url, destPath);

// 带选项的请求
const response = fetch.get(url, {
  headers: { "User-Agent": "Chopsticks" },
  timeout: 30000,
});
```

---

## 7. 文件系统模块 (fs)

```javascript
// 读取文件
const content = fs.readFile("path/to/file");
const content = fs.readFile("path/to/file", "utf8");

// 写入文件
fs.writeFile("path/to/file", content);

// 复制文件
fs.copy("src", "dst");

// 删除文件
fs.remove("path/to/file");
fs.removeAll("path/to/dir");

// 创建目录
fs.mkdir("path/to/dir");
fs.mkdirAll("path/to/nested/dir");

// 读取目录
const entries = fs.readDir("path/to/dir");
// entries = ["file1", "file2", "dir1"]

// 检查
fs.exists("path/to/file");
fs.isDir("path/to/file");
fs.isFile("path/to/file");

// 获取文件信息
const info = fs.stat("path/to/file");
// info.size, info.isDirectory, info.isFile, info.mtime
```

---

## 8. 校验和模块 (checksum)

```javascript
// 计算 SHA256
const result = checksum.sha256("path/to/file");
// result.success, result.hash

// 计算 MD5
const result = checksum.md5("path/to/file");
// result.success, result.hash

// 验证校验和
const result = checksum.verify("path/to/file", expectedHash, "sha256");
// result.success, result.valid

// 通用算法
const result = checksum.hash("path/to/file", "sha256");
// result.success, result.hash
```

---

## 9. 版本比较模块 (semver)

```javascript
// 比较版本
const result = semver.compare("1.2.3", "1.2.4"); // -1, 0, 1

// 大于/小于/等于
semver.gt("2.0.0", "1.9.0"); // true
semver.lt("1.0.0", "2.0.0"); // true
semver.eq("1.0.0", "1.0.0"); // true

// 大于等于/小于等于
semver.gte("1.2.3", "1.0.0"); // true
semver.lte("1.0.0", "1.2.3"); // true

// 范围判断
semver.satisfies("1.2.3", "^1.0.0"); // true
semver.satisfies("2.0.0", "^1.0.0"); // false
```

---

## 10. 压缩模块 (archive)

```javascript
// 解压 ZIP
const result = archive.extractZip("archive.zip", "dest/dir");
// result.success

// 解压 7z
const result = archive.extract7z("archive.7z", "dest/dir7z", "dest/dir");
// result.success

// 解压 tar.gz
const result = archive.extractTarGz("archive.tar.gz", "dest/dir");
// result.success

// 自动根据扩展名解压
const result = archive.extract("archive.zip", "dest/dir");
// result.success
```

---

## 11. 符号链接模块 (symlink)

```javascript
// 创建符号链接（文件）
const result = symlink.create("target/file.exe", "link/name.exe");
// result.success

// 创建目录符号链接
const result = symlink.createDir("target/dir", "link/dir");
// result.success

// 创建硬链接
const result = symlink.createHard("target/file", "link/file");
// result.success

// 创建 Windows 目录联接
const result = symlink.createJunction("target/dir", "link/dir");
// result.success

// 读取链接目标
const result = symlink.readLink("link");
// result.success, result.target

// 检查是否为链接
const result = symlink.isLink("path");
// result.success, result.isLink
```

---

## 12. Windows 注册表模块 (registry)

```javascript
// 设置字符串值
const result = registry.setValue("HKCU\\Software\\App", "Version", "1.0.0");
// result.success

// 设置 DWORD 值
const result = registry.setDword("HKCU\\Software\\App", "Count", 42);
// result.success

// 设置二进制值
const result = registry.setBinary(
  "HKCU\\Software\\App",
  "Data",
  Buffer.from([0x01, 0x02]),
);
// result.success

// 读取值
const result = registry.getValue("HKCU\\Software\\App", "Version");
// result.success, result.value

// 删除值
const result = registry.deleteValue("HKCU\\Software\\App", "Version");
// result.success

// 创建键
const result = registry.createKey("HKCU\\Software\\App");
// result.success

// 删除键
const result = registry.deleteKey("HKCU\\Software\\App");
// result.success

// 检查键是否存在
const result = registry.keyExists("HKCU\\Software\\App");
// result.success, result.exists

// 列出子键
const result = registry.listKeys("HKCU\\Software");
// result.success, result.keys

// 列出值
const result = registry.listValues("HKCU\\Software\\App");
// result.success, result.values
```

---

## 13. 安装程序模块 (installer)

```javascript
// 运行安装程序（自动检测类型）
const result = installer.run("installer.exe", ["/S", "/D=path"]);
// result.success

// 指定类型
const result = installer.runNSIS("installer.exe", ["/S"]);
// result.success

const result = installer.runMSI("msi.msi", ["/quiet", "/norestart"]);
// result.success

const result = installer.runInno("setup.exe", ["/VERYSILENT", "/SUPPRESSMSGBOXES"]);
// result.success

// 等待安装完成
const result = installer.waitForProcess(processName);
// result.success

// 检查安装程序类型
const result = installer.detectType("installer.exe");
// result.success, result.type = "nsis" | "msi" | "inno" | "autoit" | "unknown"
```

---

## 14. Chopsticks 系统模块 (chopsticks)

```javascript
// 获取安装目录
const cookDir = chopsticks.getCookDir("git", "2.43.0");

// 获取缓存目录
const cacheDir = chopsticks.getCacheDir();

// 获取用户配置目录
const configDir = chopsticks.getConfigDir();

// 环境变量操作
const result = chopsticks.setEnv("VAR_NAME", "value");
// result.success

const result = chopsticks.getEnv("VAR_NAME");
// result.success, result.value

const result = chopsticks.deleteEnv("VAR_NAME");
// result.success

// PATH 管理
const result = chopsticks.addToPath("path/to/bin");
// result.success

const result = chopsticks.removeFromPath("path/to/bin");
// result.success

const paths = chopsticks.getPath();

// 创建 shim（命令快捷方式）
// shim 会创建在 %USERPROFILE%\.chopsticks\shim\ 目录下
// 该目录已自动添加到 PATH，用户可直接在命令行调用
const result = chopsticks.createShim("source.exe", "alias");
// result.success

// 获取 shim 目录
const shimDir = chopsticks.getShimDir();

// 获取 persist 目录（持久化数据目录）
// persist 目录用于存储更新时需要保留的用户配置和数据
const persistDir = chopsticks.getPersistDir();

// 创建快捷方式（Windows）
const result = chopsticks.createShortcut({
  source: "app.exe",
  name: "My App",
  description: "Application description",
  icon: "app.ico",
  workingDir: "C:\\app",
  arguments: "--start",
});
// result.success

// 持久化数据
const result = chopsticks.persistData("appname", ["config", "data"]);
// result.success
```

---

## 15. 完整示例

```javascript
const { App } = require("@chopsticks/core");

class GitApp extends App {
  constructor() {
    super({
      name: "git",
      description: "Distributed version control system",
      homepage: "https://git-scm.com/",
      license: "GPL-2.0",
      bucket: "main",
    });
  }

  checkVersion() {
    try {
      const response = fetch.get(
        "https://api.github.com/repos/git-for-windows/git/releases/latest",
      );
      if (!response.success) {
        log.warn("Failed to fetch version");
        return "2.43.0"; // fallback
      }
      const data = JSON.parse(response.body);
      return data.tag_name.replace(/^v/, "");
    } catch (error) {
      log.warn("Failed to fetch version: " + error.message);
      return "2.43.0"; // fallback
    }
  }

  getDownloadInfo(version, arch) {
    const archMap = { amd64: "64-bit", x86: "32-bit" };
    const filename = `PortableGit-${version}-${archMap[arch] || arch}.7z.exe`;
    return {
      url: `https://github.com/git-for-windows/git/releases/download/v${version}.windows.1/${filename}`,
      type: "7z",
    };
  }

  onPostInstall(ctx) {
    log.info("Configuring Git...");
    const gitExe = path.join(ctx.cookDir, "bin", "git.exe");

    // 设置 Git 配置
    exec.exec(gitExe, "config", "--global", "core.autocrlf", "true");
    exec.exec(gitExe, "config", "--global", "core.longpaths", "true");

    // 添加到 PATH
    chopsticks.addToPath(path.join(ctx.cookDir, "bin"));

    // 创建快捷方式
    chopsticks.createShortcut({
      source: path.join(ctx.cookDir, "git-bash.exe"),
      name: "Git Bash",
      description: "Git Bash - Command line interface",
      icon: path.join(
        ctx.cookDir,
        "mingw64",
        "share",
        "git",
        "git-for-windows.ico",
      ),
    });

    log.info("Git installed successfully!");
  }
}

module.exports = new GitApp();
```

---

## 16. 输出模块 (output)

### 16.1 彩色输出

```javascript
// 成功消息（绿色）
output.success("安装成功");
output.successf("%s 安装完成", "git");
output.successln("✓ 操作完成");

// 错误消息（红色）
output.error("安装失败");
output.errorf("错误: %s", err.message);
output.errorln("✗ 操作失败");

// 警告消息（黄色）
output.warning("注意: 配置文件已存在");
output.warningf("警告: %s", message);
output.warningln("⚠ 请检查配置");

// 信息消息（蓝色）
output.info("正在下载...");
output.infof("下载进度: %d%%", 50);
output.infoln("ℹ 提示信息");

// 高亮消息（青色）
output.highlight("重要: 请备份数据");
output.highlightf("→ 下一步: %s", "配置环境变量");
output.highlightln("→ 开始安装");

// 暗淡消息（灰色）
output.dim("详细信息...");
output.dimf("路径: %s", path);
output.dimln("(可选)");

// 带图标的输出
output.successCheck("安装完成"); // ✓
output.errorCross("安装失败"); // ✗
output.warningSign("配置警告"); // ⚠
output.infoSign("提示信息"); // ℹ
output.arrow("下一步"); // →
```

### 16.2 颜色控制

```javascript
// 禁用颜色输出
output.disableColor();

// 启用颜色输出
output.enableColor();

// 检查颜色是否启用
const enabled = output.isColorEnabled();
```

### 16.3 进度显示

```javascript
// 创建进度管理器
const pm = output.newProgressManager();

// 添加下载进度条
const bar = pm.addDownloadBar("nodejs.zip", fileSize);
// 显示: nodejs.zip 12.5 MB / 50.0 MB [25%] 2.5 MB/s  ETA 15s

// 更新进度
bar.incrBy(bytesRead);

// 添加安装进度条（多阶段）
const stages = ["下载", "解压", "安装", "配置"];
const installBar = pm.addInstallBar("nodejs", stages);
// 显示: nodejs [下载] 25% 1/4

// 设置当前阶段
installBar.setStage(0); // 下载阶段
installBar.completeStage(); // 完成当前阶段，自动进入下一阶段

// 设置阶段内进度 (0-100)
installBar.setProgress(50); // 当前阶段完成 50%

// 标记完成
installBar.complete();

// 添加批量操作进度条
const batchBar = pm.addBatchBar(totalApps);
// 显示: [3/10] 当前应用名

// 进入下一项
batchBar.nextItem("git");
batchBar.nextItem("nodejs");

// 标记完成
batchBar.complete();

// 等待所有进度条完成
pm.wait();
```

---

## 17. 错误处理

```javascript
checkVersion() {
    const response = fetch.get(url);
    
    if (!response.success) {
        log.error("Error: " + response.error);
        return "fallback-version";
    }
    
    try {
        return JSON.parse(response.body).version;
    } catch (error) {
        log.error("Exception: " + error.message);
        return "fallback-version";
    }
}
```

---

## 18. 设备同步 API

### 18.1 执行同步

```javascript
// 执行完整同步
const result = chopsticks.sync();
// result.success, result.syncedDevices, result.conflicts

// 指定设备同步
const result = chopsticks.sync({
  device: "laptop",
  force: false,
  dryRun: false,
  configOnly: false
});
// result.success, result.syncedDevices, result.conflicts
```

**参数说明**:

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `device` | string | null | 目标设备名称，null 表示所有设备 |
| `force` | boolean | false | 是否强制覆盖冲突 |
| `dryRun` | boolean | false | 是否模拟运行 |
| `configOnly` | boolean | false | 是否仅同步配置 |

**返回值**:

```javascript
{
  success: true,
  syncedDevices: ["laptop", "desktop"],
  conflicts: [],
  timestamp: "2026-02-28T10:30:00Z"
}
```

### 18.2 获取同步状态

```javascript
const result = chopsticks.syncStatus();
// result.success, result.lastSync, result.pendingChanges, result.connectedDevices
```

**返回值**:

```javascript
{
  success: true,
  lastSync: "2026-02-28T10:30:00Z",
  pendingChanges: 5,
  connectedDevices: ["laptop", "desktop", "server"],
  syncEnabled: true
}
```

### 18.3 解决冲突

```javascript
const result = chopsticks.resolveConflict({
  conflictId: "conflict-001",
  resolution: "local"  // 'local', 'remote', 'merge'
});
// result.success
```

### 18.4 获取同步历史

```javascript
const result = chopsticks.getSyncHistory({
  limit: 10,
  device: "laptop"
});
// result.success, result.history
```

---

## 附录 A: API 设计说明

### A.1 同步 vs 异步

Chopsticks 的 JavaScript/Lua API 采用**同步设计**，原因如下：

1. **本地操作为主**：大多数 API 是文件系统、注册表等本地操作，同步调用更直观
2. **Lua 兼容性**：Lua 引擎原生不支持 Promise/异步
3. **简化使用**：脚本编写者无需处理异步复杂性
4. **性能足够**：本地操作耗时通常在毫秒级

### A.2 网络请求

虽然 `fetch` 等网络 API 是同步的，但内部使用 Go 的 HTTP 客户端，
对于需要异步处理的场景，建议使用 Go 协程配合回调。

### A.3 返回值格式

所有 API 统一返回以下格式的结果对象：

**JavaScript：**
```javascript
// 成功
{
    success: true,
    data: <返回数据>,      // 可选
    error: null
}

// 失败
{
    success: false,
    data: null,
    error: "错误信息"
}
```

**Lua：**
```lua
-- 成功
return <数据>, nil

-- 失败
return nil, "错误信息"
```

### A.4 未来规划

后续版本可能引入可选的异步 API（如 `fetch.asyncGet()`），
但同步 API 将始终保持兼容。

---

_最后更新：2026-02-28_
