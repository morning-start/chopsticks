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
const result = await exec.exec("git", "--version");
// result.exitCode, result.stdout, result.stderr, result.success

// 执行 shell 命令
const result = await exec.shell("echo hello");

// 执行 PowerShell 命令
const result = await exec.powershell("Get-Process");
```

---

## 6. HTTP 模块 (fetch)

```javascript
// GET 请求
const response = await fetch.get(url);
// response.status, response.ok, response.body, response.headers

// POST 请求
const response = await fetch.post(url, body, "application/json");

// 下载文件
await fetch.download(url, destPath);

// 带选项的请求
const response = await fetch.get(url, {
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
const hash = await checksum.sha256("path/to/file");

// 计算 MD5
const hash = await checksum.md5("path/to/file");

// 验证校验和
const valid = await checksum.verify("path/to/file", expectedHash, "sha256");

// 通用算法
const hash = await checksum.hash("path/to/file", "sha256");
const hash = await checksum.hash("path/to/file", "md5");
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
await archive.extractZip("archive.zip", "dest/dir");

// 解压 7z
await archive.extract7z("archive.7z", "dest/dir7z", "dest/dir");

// 解压 tar.gz
await archive.extractTarGz("archive.tar.gz", "dest/dir");

// 自动根据扩展名解压
await archive.extract("archive.zip", "dest/dir");
await archive.extract("archive.7z", "dest/dir");
await archive.extract("archive.tar.gz", "dest/dir");
```

---

## 11. 符号链接模块 (symlink)

```javascript
// 创建符号链接（文件）
await symlink.create("target/file.exe", "link/name.exe");

// 创建目录符号链接
await symlink.createDir("target/dir", "link/dir");

// 创建硬链接
await symlink.createHard("target/file", "link/file");

// 创建 Windows 目录联接
await symlink.createJunction("target/dir", "link/dir");

// 读取链接目标
const target = symlink.readLink("link");

// 检查是否为链接
const isLink = symlink.isLink("path");
```

---

## 12. Windows 注册表模块 (registry)

```javascript
// 设置字符串值
await registry.setValue("HKCU\\Software\\App", "Version", "1.0.0");

// 设置 DWORD 值
await registry.setDword("HKCU\\Software\\App", "Count", 42);

// 设置二进制值
await registry.setBinary(
  "HKCU\\Software\\App",
  "Data",
  Buffer.from([0x01, 0x02]),
);

// 读取值
const value = await registry.getValue("HKCU\\Software\\App", "Version");

// 删除值
await registry.deleteValue("HKCU\\Software\\App", "Version");

// 创建键
await registry.createKey("HKCU\\Software\\App");

// 删除键
await registry.deleteKey("HKCU\\Software\\App");

// 检查键是否存在
const exists = await registry.keyExists("HKCU\\Software\\App");

// 列出子键
const keys = await registry.listKeys("HKCU\\Software");

// 列出值
const values = await registry.listValues("HKCU\\Software\\App");
```

---

## 13. 安装程序模块 (installer)

```javascript
// 运行安装程序（自动检测类型）
await installer.run("installer.exe", ["/S", "/D=path"]);

// 指定类型
await installer.runNSIS("installer.exe", ["/S"]);
await installer.runMSI("msi.msi", ["/quiet", "/norestart"]);
await installer.runInno("setup.exe", ["/VERYSILENT", "/SUPPRESSMSGBOXES"]);

// 等待安装完成
await installer.waitForProcess(processName);

// 检查安装程序类型
const type = await installer.detectType("installer.exe");
// type = "nsis" | "msi" | "inno" | "autoit" | "unknown"
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
await chopsticks.setEnv("VAR_NAME", "value");
const value = await chopsticks.getEnv("VAR_NAME");
await chopsticks.deleteEnv("VAR_NAME");

// PATH 管理
await chopsticks.addToPath("path/to/bin");
await chopsticks.removeFromPath("path/to/bin");
const paths = chopsticks.getPath();

// 创建 shim
await chopsticks.createShim("source.exe", "alias");

// 创建快捷方式（Windows）
await chopsticks.createShortcut({
  source: "app.exe",
  name: "My App",
  description: "Application description",
  icon: "app.ico",
  workingDir: "C:\\app",
  arguments: "--start",
});

// 持久化数据
await chopsticks.persistData("appname", ["config", "data"]);
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

  async checkVersion() {
    try {
      const response = await fetch.get(
        "https://api.github.com/repos/git-for-windows/git/releases/latest",
      );
      const data = JSON.parse(response.body);
      return data.tag_name.replace(/^v/, "");
    } catch (error) {
      log.warn("Failed to fetch version: " + error.message);
      return "2.43.0"; // fallback
    }
  }

  async getDownloadInfo(version, arch) {
    const archMap = { amd64: "64-bit", x86: "32-bit" };
    const filename = `PortableGit-${version}-${archMap[arch] || arch}.7z.exe`;
    return {
      url: `https://github.com/git-for-windows/git/releases/download/v${version}.windows.1/${filename}`,
      type: "7z",
    };
  }

  async onPostInstall(ctx) {
    log.info("Configuring Git...");
    const gitExe = path.join(ctx.cookDir, "bin", "git.exe");

    // 设置 Git 配置
    await exec.exec(gitExe, "config", "--global", "core.autocrlf", "true");
    await exec.exec(gitExe, "config", "--global", "core.longpaths", "true");

    // 添加到 PATH
    await chopsticks.addToPath(path.join(ctx.cookDir, "bin"));

    // 创建快捷方式
    await chopsticks.createShortcut({
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

## 16. 错误处理

```javascript
async checkVersion() {
    try {
        const result = await this.safeCall(async () => {
            const response = await fetch.get(url);
            return JSON.parse(response.body).version;
        });

        if (!result.success) {
            log.error("Error: " + result.error);
            return "fallback-version";
        }

        return result.value;
    } catch (error) {
        log.error("Exception: " + error.message);
        return "fallback-version";
    }
}
```

---

_最后更新：2026-02-26_
