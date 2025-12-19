# Chopsticks JS 插件系统文档

## 1. 概述

Chopsticks 采用「核心 + 插件」分离架构，核心使用 Rust 编写，负责 CLI 解析、生命周期管理、安全沙箱等基础设施；插件使用 JavaScript 编写，负责描述软件的「是什么」和「如何安装」。

本文档详细介绍 JS 插件系统中提供的核心功能：文件校验和、正则表达式、网络请求、文件读写。

## 2. 核心 API

### 2.1 文件校验和

Chopsticks 提供了多种文件校验和算法，用于验证下载文件的完整性。

#### 2.1.1 API 方法

| 方法名 | 参数 | 返回值 | 描述 |
|--------|------|--------|------|
| `checksum.sha256()` | `filePath: string` | `Promise<string>` | 计算文件的 SHA256 校验和 |
| `checksum.sha512()` | `filePath: string` | `Promise<string>` | 计算文件的 SHA512 校验和 |
| `checksum.md5()` | `filePath: string` | `Promise<string>` | 计算文件的 MD5 校验和 |
| `checksum.verify()` | `filePath: string, expectedSum: string, algorithm: string` | `Promise<boolean>` | 验证文件校验和是否匹配 |

#### 2.1.2 使用示例

```javascript
// 计算文件校验和
const sha256 = await checksum.sha256('downloaded/file.zip');
console.log(`SHA256: ${sha256}`);

// 验证文件校验和
const isValid = await checksum.verify('downloaded/file.zip', 'expected-sha256', 'sha256');
if (isValid) {
    console.log('文件校验通过');
} else {
    console.log('文件校验失败');
}
```

### 2.2 正则表达式

Chopsticks 提供了正则表达式工具，方便插件处理字符串匹配和提取。

#### 2.2.1 API 方法

| 方法名 | 参数 | 返回值 | 描述 |
|--------|------|--------|------|
| `regex.match()` | `pattern: string, text: string, flags?: string` | `Array<string> | null` | 执行正则匹配 |
| `regex.replace()` | `pattern: string, text: string, replacement: string, flags?: string` | `string` | 执行正则替换 |
| `regex.test()` | `pattern: string, text: string, flags?: string` | `boolean` | 测试正则匹配 |
| `regex.exec()` | `pattern: string, text: string, flags?: string` | `Array<string> | null` | 执行正则搜索 |

#### 2.2.2 使用示例

```javascript
// 匹配版本号
const version = regex.match(r'v(\d+\.\d+\.\d+)', 'Release v1.2.3', 'i');
console.log(`版本号: ${version[1]}`);

// 替换文本
const cleanedText = regex.replace(r'[^a-zA-Z0-9]', 'Hello@World#123', '_');
console.log(`清理后的文本: ${cleanedText}`);

// 测试匹配
const hasDigit = regex.test(r'\d', 'abc123');
console.log(`包含数字: ${hasDigit}`);
```

### 2.3 网络请求

Chopsticks 提供了安全的网络请求 API，用于插件下载文件、检查更新等。

#### 2.3.1 API 方法

| 方法名 | 参数 | 返回值 | 描述 |
|--------|------|--------|------|
| `http.get()` | `url: string, options?: object` | `Promise<Response>` | 发送 GET 请求 |
| `http.post()` | `url: string, data?: any, options?: object` | `Promise<Response>` | 发送 POST 请求 |
| `http.download()` | `url: string, destPath: string, options?: object` | `Promise<DownloadResult>` | 下载文件到指定路径 |
| `http.head()` | `url: string, options?: object` | `Promise<Headers>` | 发送 HEAD 请求获取响应头 |

#### 2.3.2 Response 对象

| 属性名 | 类型 | 描述 |
|--------|------|------|
| `status` | `number` | HTTP 状态码 |
| `headers` | `object` | 响应头 |
| `body` | `string` | 响应体 |

#### 2.3.3 使用示例

```javascript
// 发送 GET 请求
const response = await http.get('https://api.example.com/version');
const versionInfo = JSON.parse(response.body);
console.log(`最新版本: ${versionInfo.version}`);

// 下载文件
const result = await http.download('https://example.com/app.zip', 'downloads/app.zip');
console.log(`下载完成: ${result.path}, 大小: ${result.size} 字节`);

// 发送 POST 请求
const postResponse = await http.post('https://api.example.com/submit', {
    name: 'test',
    version: '1.0.0'
});
```

### 2.4 文件读写

Chopsticks 提供了安全的文件读写 API，限制在指定目录内操作，确保系统安全。

#### 2.4.1 API 方法

| 方法名 | 参数 | 返回值 | 描述 |
|--------|------|--------|------|
| `fs.readFile()` | `path: string, options?: object` | `Promise<string | Buffer>` | 读取文件内容 |
| `fs.writeFile()` | `path: string, data: string | Buffer, options?: object` | `Promise<void>` | 写入文件内容 |
| `fs.appendFile()` | `path: string, data: string | Buffer, options?: object` | `Promise<void>` | 追加文件内容 |
| `fs.readDir()` | `path: string, options?: object` | `Promise<Array<string>>` | 读取目录内容 |
| `fs.exists()` | `path: string` | `Promise<boolean>` | 检查文件/目录是否存在 |
| `fs.mkdir()` | `path: string, options?: object` | `Promise<void>` | 创建目录 |
| `fs.copy()` | `src: string, dest: string, options?: object` | `Promise<void>` | 复制文件/目录 |
| `fs.move()` | `src: string, dest: string, options?: object` | `Promise<void>` | 移动文件/目录 |
| `fs.remove()` | `path: string, options?: object` | `Promise<void>` | 删除文件/目录 |

#### 2.4.2 使用示例

```javascript
// 读取文件
const content = await fs.readFile('config.json', { encoding: 'utf8' });
const config = JSON.parse(content);

// 写入文件
await fs.writeFile('output.txt', 'Hello, Chopsticks!', { encoding: 'utf8' });

// 读取目录
const files = await fs.readDir('downloads');
console.log(`目录文件: ${files.join(', ')}`);

// 检查文件是否存在
const exists = await fs.exists('app.exe');
console.log(`文件存在: ${exists}`);
```

## 3. 插件开发最佳实践

1. **安全性优先**：只使用提供的 API，避免尝试访问系统资源
2. **错误处理**：始终使用 try-catch 处理异步操作
3. **性能优化**：避免在插件中执行复杂计算，将耗时操作交给 Rust 核心
4. **版本兼容**：插件应检查 API 版本，确保兼容性
5. **测试充分**：为插件编写单元测试，确保功能正确性

## 4. 安全沙箱限制

为确保系统安全，Chopsticks 对 JS 插件施加了以下限制：

- 所有文件操作限制在 Chopsticks 目录结构内
- 网络请求需经过核心验证，防止恶意请求
- 不允许执行系统命令
- 内存使用有限制
- 执行时间有限制

## 5. 插件示例

```javascript
class Software {
    constructor(name, version, updateScript, installScript) {
        this.name = name;
        this.version = version;
        this.updateScript = updateScript;
        this.installScript = installScript;
    }
    
    async checkForUpdates(domain, oldVersion) {
        try {
            // 发送网络请求获取最新版本
            const response = await http.get(`https://${domain}/api/version`);
            const latestVersion = response.body.trim();
            
            // 使用正则验证版本号格式
            if (!regex.test(r'^\d+\.\d+\.\d+$', latestVersion)) {
                throw new Error('无效的版本号格式');
            }
            
            return latestVersion;
        } catch (error) {
            console.error('检查更新失败:', error.message);
            return oldVersion;
        }
    }
    
    async downloadAndVerify(url, destPath, expectedSha256) {
        // 下载文件
        await http.download(url, destPath);
        
        // 验证文件校验和
        const isValid = await checksum.verify(destPath, expectedSha256, 'sha256');
        if (!isValid) {
            throw new Error('文件校验失败');
        }
        
        return destPath;
    }
}
```

## 6. 总结

Chopsticks 的 JS 插件系统提供了丰富的 API，让插件开发者能够轻松实现各种功能，同时保持系统安全性。通过合理使用这些 API，开发者可以创建强大、灵活的软件包管理插件，为用户提供更好的体验。

核心设计理念：
- 保持 Scoop 的简洁与约定
- 赋予 manifest 动态能力
- 让插件强大，而核心保持轻量
- 安全第一，沙箱隔离

Chopsticks 不是重造轮子，而是用现代技术为 Scoop 的理念注入新活力——让包管理 manifest 从「静态配置」走向「可编程逻辑」，同时坚守简洁、安全、可靠的原则。