# Chopsticks JavaScript API 参考

> 应用脚本中可用的 JavaScript API 完整参考

---

## 目录

1. [fsutil - 文件操作](#1-fsutil---文件操作)
2. [fetch - HTTP 请求](#2-fetch---http-请求)
3. [execx - 命令执行](#3-execx---命令执行)
4. [archive - 压缩解压](#4-archive---压缩解压)
5. [checksum - 校验和](#5-checksum---校验和)
6. [chopsticksx - Chopsticks 核心](#6-chopsticksx---chopsticks-核心)
7. [jsonx - JSON 处理](#7-jsonx---json-处理)
8. [logx - 日志](#8-logx---日志)
9. [pathx - 路径操作](#9-pathx---路径操作)
10. [registry - 注册表操作](#10-registry---注册表操作)
11. [semver - 版本控制](#11-semver---版本控制)
12. [symlink - 符号链接](#12-symlink---符号链接)
13. [installerx - 安装程序](#13-installerx---安装程序)

---

## 1. fsutil - 文件操作

文件系统操作模块，提供文件的读写、目录管理等功能。

### 函数列表

| 函数                        | 说明                  |
| --------------------------- | --------------------- |
| `readFile(path, encoding?)` | 读取文件内容          |
| `writeFile(path, content)`  | 写入文件内容          |
| `append(path, content)`     | 追加内容到文件        |
| `mkdir(path)`               | 创建目录              |
| `rmdir(path)`               | 删除空目录            |
| `remove(path)`              | 删除文件              |
| `exists(path)`              | 检查文件/目录是否存在 |
| `isdir(path)`               | 检查路径是否为目录    |
| `readDir(path)`             | 读取目录内容          |
| `copy(src, dst)`            | 复制文件              |
| `removeAll(path)`           | 递归删除目录及其内容  |
| `mkdirAll(path)`            | 递归创建目录          |
| `isFile(path)`              | 检查路径是否为文件    |
| `stat(path)`                | 获取文件/目录信息     |

### readFile(path, encoding?)

读取文件内容。

**参数：**

| 参数       | 类型   | 必填 | 说明                      |
| ---------- | ------ | ---- | ------------------------- |
| `path`     | string | 是   | 文件路径                  |
| `encoding` | string | 否   | 编码格式，默认为 `"utf8"` |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `content` | string  | 文件内容（成功时） |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fsutil.readFile("config.json");
if (result.success) {
  logx.info("文件内容: " + result.content);
}

// 读取二进制文件
const result = fsutil.readFile("data.bin", "binary");
```

### writeFile(path, content)

写入文件内容。

**参数：**

| 参数      | 类型   | 必填 | 说明     |
| --------- | ------ | ---- | -------- |
| `path`    | string | 是   | 文件路径 |
| `content` | string | 是   | 文件内容 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fsutil.writeFile("config.json", '{"name": "test"}');
if (result.success) {
  logx.info("文件写入成功");
}
```

### append(path, content)

追加内容到文件末尾。

**参数：**

| 参数      | 类型   | 必填 | 说明         |
| --------- | ------ | ---- | ------------ |
| `path`    | string | 是   | 文件路径     |
| `content` | string | 是   | 要追加的内容 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
fsutil.append("log.txt", "新的日志行\n");
```

### mkdir(path)

创建单个目录。

**参数：**

| 参数   | 类型   | 必填 | 说明     |
| ------ | ------ | ---- | -------- |
| `path` | string | 是   | 目录路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
fsutil.mkdir("new_folder");
```

### rmdir(path)

删除空目录。

**参数：**

| 参数   | 类型   | 必填 | 说明     |
| ------ | ------ | ---- | -------- |
| `path` | string | 是   | 目录路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
fsutil.rmdir("empty_folder");
```

### remove(path)

删除文件。

**参数：**

| 参数   | 类型   | 必填 | 说明     |
| ------ | ------ | ---- | -------- |
| `path` | string | 是   | 文件路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
fsutil.remove("old_file.txt");
```

### exists(path)

检查文件或目录是否存在。

**参数：**

| 参数   | 类型   | 必填 | 说明 |
| ------ | ------ | ---- | ---- |
| `path` | string | 是   | 路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `exists`  | boolean | 是否存在           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fsutil.exists("config.json");
if (result.success && result.exists) {
  logx.info("文件存在");
}
```

### isdir(path)

检查路径是否为目录。

**参数：**

| 参数   | 类型   | 必填 | 说明 |
| ------ | ------ | ---- | ---- |
| `path` | string | 是   | 路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `isdir`   | boolean | 是否为目录         |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fsutil.isdir("my_folder");
if (result.success && result.isdir) {
  logx.info("这是一个目录");
}
```

### readDir(path)

读取目录内容。

**参数：**

| 参数   | 类型   | 必填 | 说明     |
| ------ | ------ | ---- | -------- |
| `path` | string | 是   | 目录路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `entries` | array   | 目录项列表         |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fsutil.readDir("my_folder");
if (result.success) {
  for (const entry of result.entries) {
    logx.info("条目: " + entry);
  }
}
```

### copy(src, dst)

复制文件。

**参数：**

| 参数  | 类型   | 必填 | 说明         |
| ----- | ------ | ---- | ------------ |
| `src` | string | 是   | 源文件路径   |
| `dst` | string | 是   | 目标文件路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
fsutil.copy("source.txt", "backup.txt");
```

### removeAll(path)

递归删除目录及其所有内容。

**参数：**

| 参数   | 类型   | 必填 | 说明     |
| ------ | ------ | ---- | -------- |
| `path` | string | 是   | 目录路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
fsutil.removeAll("temp_folder");
```

### mkdirAll(path)

递归创建目录（自动创建所有父目录）。

**参数：**

| 参数   | 类型   | 必填 | 说明     |
| ------ | ------ | ---- | -------- |
| `path` | string | 是   | 目录路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
fsutil.mkdirAll("path/to/nested/folder");
```

### isFile(path)

检查路径是否为文件。

**参数：**

| 参数   | 类型   | 必填 | 说明 |
| ------ | ------ | ---- | ---- |
| `path` | string | 是   | 路径 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `isFile`  | boolean | 是否为文件         |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fsutil.isFile("document.txt");
if (result.success && result.isFile) {
  logx.info("这是一个文件");
}
```

### stat(path)

获取文件或目录的详细信息。

**参数：**

| 参数   | 类型   | 必填 | 说明 |
| ------ | ------ | ---- | ---- |
| `path` | string | 是   | 路径 |

**返回值：**

| 字段          | 类型    | 说明               |
| ------------- | ------- | ------------------ |
| `success`     | boolean | 是否成功           |
| `size`        | number  | 文件大小（字节）   |
| `isDirectory` | boolean | 是否为目录         |
| `isFile`      | boolean | 是否为文件         |
| `mtime`       | string  | 最后修改时间       |
| `error`       | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fsutil.stat("file.txt");
if (result.success) {
  logx.info("大小: " + result.size + " 字节");
  logx.info("修改时间: " + result.mtime);
}
```

---

## 2. fetch - HTTP 请求

HTTP 请求模块，提供下载、GET/POST 请求等功能。

### 函数列表

| 函数                                      | 说明                     |
| ----------------------------------------- | ------------------------ |
| `download(url, destPath, options?)`       | 下载文件到指定路径       |
| `get(url, options?)`                      | 发送 GET 请求            |
| `post(url, body, contentType?, options?)` | 发送 POST 请求           |
| `request(method, url, options?)`          | 发送通用 HTTP 请求       |
| `downloadFile(url, destPath, options?)`   | 下载文件（别名）         |
| `parseURL(url)`                           | 解析 URL                 |
| `buildURL(base, params)`                  | 构建带查询参数的 URL     |
| `getJSON(url, options?)`                  | 发送 GET 请求并解析 JSON |
| `postJSON(url, data, options?)`           | 发送 JSON POST 请求      |
| `newClient(options?)`                     | 创建新的 HTTP 客户端     |
| `setDefaultTimeout(timeout)`              | 设置默认超时时间         |

### download(url, destPath, options?)

下载文件到指定路径。

**参数：**

| 参数       | 类型   | 必填 | 说明         |
| ---------- | ------ | ---- | ------------ |
| `url`      | string | 是   | 下载地址     |
| `destPath` | string | 是   | 目标保存路径 |
| `options`  | object | 否   | 下载选项     |

**options 选项：**

| 选项      | 类型   | 说明             |
| --------- | ------ | ---------------- |
| `headers` | object | 请求头           |
| `timeout` | number | 超时时间（毫秒） |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fetch.download(
  "https://example.com/file.zip",
  "downloads/file.zip",
  { timeout: 60000 },
);
if (result.success) {
  logx.info("下载完成");
}
```

### get(url, options?)

发送 GET 请求。

**参数：**

| 参数      | 类型   | 必填 | 说明     |
| --------- | ------ | ---- | -------- |
| `url`     | string | 是   | 请求地址 |
| `options` | object | 否   | 请求选项 |

**返回值：**

| 字段      | 类型    | 说明                   |
| --------- | ------- | ---------------------- |
| `success` | boolean | 是否成功               |
| `status`  | number  | HTTP 状态码            |
| `ok`      | boolean | 请求是否成功 (200-299) |
| `body`    | string  | 响应体                 |
| `headers` | object  | 响应头                 |
| `error`   | string  | 错误信息（失败时）     |

**示例：**

```javascript
const result = fetch.get("https://api.example.com/data");
if (result.success && result.ok) {
  logx.info("响应: " + result.body);
}

// 带请求头
const result = fetch.get("https://api.example.com/data", {
  headers: { "User-Agent": "Chopsticks/1.0" },
});
```

### post(url, body, contentType?, options?)

发送 POST 请求。

**参数：**

| 参数          | 类型   | 必填 | 说明                                                 |
| ------------- | ------ | ---- | ---------------------------------------------------- |
| `url`         | string | 是   | 请求地址                                             |
| `body`        | string | 是   | 请求体                                               |
| `contentType` | string | 否   | 内容类型，默认 `"application/x-www-form-urlencoded"` |
| `options`     | object | 否   | 请求选项                                             |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `status`  | number  | HTTP 状态码        |
| `ok`      | boolean | 请求是否成功       |
| `body`    | string  | 响应体             |
| `headers` | object  | 响应头             |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fetch.post(
  "https://api.example.com/submit",
  "name=test&value=123",
  "application/x-www-form-urlencoded",
);
```

### request(method, url, options?)

发送通用 HTTP 请求。

**参数：**

| 参数      | 类型   | 必填 | 说明                               |
| --------- | ------ | ---- | ---------------------------------- |
| `method`  | string | 是   | HTTP 方法 (GET/POST/PUT/DELETE 等) |
| `url`     | string | 是   | 请求地址                           |
| `options` | object | 否   | 请求选项                           |

**options 选项：**

| 选项      | 类型   | 说明     |
| --------- | ------ | -------- |
| `body`    | string | 请求体   |
| `headers` | object | 请求头   |
| `timeout` | number | 超时时间 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `status`  | number  | HTTP 状态码        |
| `body`    | string  | 响应体             |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fetch.request("PUT", "https://api.example.com/update", {
  body: '{"status": "active"}',
  headers: { "Content-Type": "application/json" },
});
```

### parseURL(url)

解析 URL 为组成部分。

**参数：**

| 参数  | 类型   | 必填 | 说明     |
| ----- | ------ | ---- | -------- |
| `url` | string | 是   | URL 地址 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `scheme`  | string  | 协议               |
| `host`    | string  | 主机               |
| `path`    | string  | 路径               |
| `query`   | string  | 查询字符串         |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fetch.parseURL("https://example.com/path?key=value");
if (result.success) {
  logx.info("主机: " + result.host);
  logx.info("路径: " + result.path);
}
```

### buildURL(base, params)

构建带查询参数的 URL。

**参数：**

| 参数     | 类型   | 必填 | 说明         |
| -------- | ------ | ---- | ------------ |
| `base`   | string | 是   | 基础 URL     |
| `params` | object | 是   | 查询参数对象 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `url`     | string  | 构建后的 URL       |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fetch.buildURL("https://api.example.com/search", {
  q: "keyword",
  page: 1,
});
// 结果: https://api.example.com/search?q=keyword&page=1
```

### getJSON(url, options?)

发送 GET 请求并自动解析 JSON 响应。

**参数：**

| 参数      | 类型   | 必填 | 说明     |
| --------- | ------ | ---- | -------- |
| `url`     | string | 是   | 请求地址 |
| `options` | object | 否   | 请求选项 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `data`    | object  | 解析后的 JSON 对象 |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fetch.getJSON(
  "https://api.github.com/repos/git/git/releases/latest",
);
if (result.success) {
  logx.info("版本: " + result.data.tag_name);
}
```

### postJSON(url, data, options?)

发送 JSON POST 请求。

**参数：**

| 参数      | 类型   | 必填 | 说明             |
| --------- | ------ | ---- | ---------------- |
| `url`     | string | 是   | 请求地址         |
| `data`    | object | 是   | 要发送的数据对象 |
| `options` | object | 否   | 请求选项         |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `data`    | object  | 响应 JSON 对象     |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fetch.postJSON("https://api.example.com/users", {
  name: "John",
  email: "john@example.com",
});
```

### newClient(options?)

创建新的 HTTP 客户端实例。

**参数：**

| 参数      | 类型   | 必填 | 说明           |
| --------- | ------ | ---- | -------------- |
| `options` | object | 否   | 客户端默认选项 |

**返回值：**

| 字段      | 类型    | 说明               |
| --------- | ------- | ------------------ |
| `success` | boolean | 是否成功           |
| `client`  | object  | HTTP 客户端实例    |
| `error`   | string  | 错误信息（失败时） |

**示例：**

```javascript
const result = fetch.newClient({
  timeout: 30000,
  headers: { "User-Agent": "MyApp/1.0" },
});
if (result.success) {
  const client = result.client;
  const response = client.get("https://api.example.com/data");
}
```

### setDefaultTimeout(timeout)

设置默认请求超时时间。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `timeout` | number | 是 | 超时时间（毫秒） |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
fetch.setDefaultTimeout(60000); // 60秒超时
```

---

## 3. execx - 命令执行

命令执行模块，提供执行外部命令、Shell 脚本等功能。

### 函数列表

| 函数 | 说明 |
|------|------|
| `exec(command, args?, options?)` | 执行命令 |
| `shell(command, options?)` | 执行 Shell 命令 |
| `powershell(command, options?)` | 执行 PowerShell 命令 |

### exec(command, args?, options?)

执行外部命令。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `command` | string | 是 | 命令名称 |
| `args` | array/string | 否 | 命令参数 |
| `options` | object | 否 | 执行选项 |

**options 选项：**

| 选项 | 类型 | 说明 |
|------|------|------|
| `cwd` | string | 工作目录 |
| `env` | object | 环境变量 |
| `timeout` | number | 超时时间（毫秒） |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功执行 |
| `exitCode` | number | 退出码 |
| `stdout` | string | 标准输出 |
| `stderr` | string | 标准错误 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
// 简单执行
const result = execx.exec("git", ["--version"]);
if (result.success) {
    logx.info("版本: " + result.stdout);
}

// 带选项
const result = execx.exec("npm", ["install"], {
    cwd: "D:\\Projects\\MyApp",
    timeout: 120000
});
```

### shell(command, options?)

执行 Shell 命令（cmd.exe）。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `command` | string | 是 | Shell 命令 |
| `options` | object | 否 | 执行选项 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功执行 |
| `exitCode` | number | 退出码 |
| `stdout` | string | 标准输出 |
| `stderr` | string | 标准错误 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = execx.shell("echo Hello World");
if (result.success) {
    logx.info("输出: " + result.stdout);
}

// 多行命令
const result = execx.shell("dir /b && echo Done");
```

### powershell(command, options?)

执行 PowerShell 命令。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `command` | string | 是 | PowerShell 命令 |
| `options` | object | 否 | 执行选项 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功执行 |
| `exitCode` | number | 退出码 |
| `stdout` | string | 标准输出 |
| `stderr` | string | 标准错误 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = execx.powershell("Get-Process | Select-Object -First 5");
if (result.success) {
    logx.info(result.stdout);
}

// 执行 PowerShell 脚本
const result = execx.powershell("Get-ChildItem -Path C:\\ -Filter *.txt");
```

---

## 4. archive - 压缩解压

压缩解压模块，支持 ZIP、7z、TAR、TAR.GZ 等格式。

### 函数列表

| 函数 | 说明 |
|------|------|
| `extract(archivePath, destPath, options?)` | 自动检测类型并解压 |
| `extractZip(zipPath, destPath, options?)` | 解压 ZIP 文件 |
| `extract7z(archivePath, destPath, options?)` | 解压 7z 文件 |
| `extractTar(tarPath, destPath, options?)` | 解压 TAR 文件 |
| `extractTarGz(tarGzPath, destPath, options?)` | 解压 TAR.GZ 文件 |
| `list(archivePath)` | 列出压缩包内容 |
| `detectType(archivePath)` | 检测压缩包类型 |
| `isArchive(filePath)` | 检查文件是否为压缩包 |

### extract(archivePath, destPath, options?)

自动检测压缩包类型并解压。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `archivePath` | string | 是 | 压缩包路径 |
| `destPath` | string | 是 | 解压目标目录 |
| `options` | object | 否 | 解压选项 |

**options 选项：**

| 选项 | 类型 | 说明 |
|------|------|------|
| `password` | string | 密码（用于加密压缩包） |
| `overwrite` | boolean | 是否覆盖现有文件 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `extractedFiles` | array | 解压的文件列表 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = archive.extract("download.zip", "extracted/");
if (result.success) {
    logx.info("解压完成，文件数: " + result.extractedFiles.length);
}
```

### extractZip(zipPath, destPath, options?)

解压 ZIP 文件。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `zipPath` | string | 是 | ZIP 文件路径 |
| `destPath` | string | 是 | 解压目标目录 |
| `options` | object | 否 | 解压选项 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `extractedFiles` | array | 解压的文件列表 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = archive.extractZip("archive.zip", "output/");
```

### extract7z(archivePath, destPath, options?)

解压 7z 文件。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `archivePath` | string | 是 | 7z 文件路径 |
| `destPath` | string | 是 | 解压目标目录 |
| `options` | object | 否 | 解压选项 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `extractedFiles` | array | 解压的文件列表 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = archive.extract7z("archive.7z", "output/");
```

### extractTar(tarPath, destPath, options?)

解压 TAR 文件。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `tarPath` | string | 是 | TAR 文件路径 |
| `destPath` | string | 是 | 解压目标目录 |
| `options` | object | 否 | 解压选项 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `extractedFiles` | array | 解压的文件列表 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = archive.extractTar("archive.tar", "output/");
```

### extractTarGz(tarGzPath, destPath, options?)

解压 TAR.GZ 文件。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `tarGzPath` | string | 是 | TAR.GZ 文件路径 |
| `destPath` | string | 是 | 解压目标目录 |
| `options` | object | 否 | 解压选项 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `extractedFiles` | array | 解压的文件列表 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = archive.extractTarGz("archive.tar.gz", "output/");
```

### list(archivePath)

列出压缩包内容。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `archivePath` | string | 是 | 压缩包路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `files` | array | 文件列表（包含 name, size, isDir 等） |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = archive.list("archive.zip");
if (result.success) {
    for (const file of result.files) {
        logx.info(file.name + " - " + file.size + " 字节");
    }
}
```

### detectType(archivePath)

检测压缩包类型。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `archivePath` | string | 是 | 文件路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `type` | string | 压缩包类型 (zip/7z/tar/tar.gz/unknown) |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = archive.detectType("file.zip");
if (result.success) {
    logx.info("类型: " + result.type);
}
```

### isArchive(filePath)

检查文件是否为压缩包。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `filePath` | string | 是 | 文件路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `isArchive` | boolean | 是否为压缩包 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = archive.isArchive("file.zip");
if (result.success && result.isArchive) {
    logx.info("这是一个压缩包");
}
```

---

## 5. checksum - 校验和

校验和计算模块，支持 MD5、SHA256、SHA512 等算法。

### 函数列表

| 函数 | 说明 |
|------|------|
| `md5(filePath)` | 计算 MD5 校验和 |
| `sha256(filePath)` | 计算 SHA256 校验和 |
| `sha512(filePath)` | 计算 SHA512 校验和 |
| `verify(filePath, expectedHash, algorithm?)` | 验证文件校验和 |
| `string(input, algorithm)` | 计算字符串的校验和 |

### md5(filePath)

计算文件的 MD5 校验和。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `filePath` | string | 是 | 文件路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `hash` | string | MD5 哈希值（32位十六进制） |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = checksum.md5("file.txt");
if (result.success) {
    logx.info("MD5: " + result.hash);
}
```

### sha256(filePath)

计算文件的 SHA256 校验和。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `filePath` | string | 是 | 文件路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `hash` | string | SHA256 哈希值（64位十六进制） |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = checksum.sha256("installer.exe");
if (result.success) {
    logx.info("SHA256: " + result.hash);
}
```

### sha512(filePath)

计算文件的 SHA512 校验和。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `filePath` | string | 是 | 文件路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `hash` | string | SHA512 哈希值（128位十六进制） |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = checksum.sha512("large_file.iso");
if (result.success) {
    logx.info("SHA512: " + result.hash);
}
```

### verify(filePath, expectedHash, algorithm?)

验证文件的校验和。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `filePath` | string | 是 | 文件路径 |
| `expectedHash` | string | 是 | 期望的哈希值 |
| `algorithm` | string | 否 | 算法 (md5/sha256/sha512)，默认 sha256 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功计算 |
| `valid` | boolean | 校验是否通过 |
| `actualHash` | string | 实际计算的哈希值 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = checksum.verify(
    "installer.exe",
    "a1b2c3d4e5f6...",
    "sha256"
);
if (result.success) {
    if (result.valid) {
        logx.info("校验通过");
    } else {
        logx.error("校验失败，实际值: " + result.actualHash);
    }
}
```

### string(input, algorithm)

计算字符串的校验和。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `input` | string | 是 | 输入字符串 |
| `algorithm` | string | 是 | 算法 (md5/sha256/sha512) |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `hash` | string | 哈希值 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = checksum.string("Hello World", "sha256");
if (result.success) {
    logx.info("SHA256: " + result.hash);
}
```

---

## 6. chopsticksx - Chopsticks 核心

Chopsticks 系统核心模块，提供系统目录、环境变量、快捷方式等功能。

### 函数列表

| 函数 | 说明 |
|------|------|
| `getCookDir(appName, version?)` | 获取应用安装目录 |
| `getCurrentVersion()` | 获取 Chopsticks 当前版本 |
| `addToPath(dirPath, scope?)` | 添加目录到 PATH |
| `removeFromPath(dirPath, scope?)` | 从 PATH 移除目录 |
| `setEnv(key, value, scope?)` | 设置环境变量 |
| `getEnv(key)` | 获取环境变量 |
| `createShim(target, name?, options?)` | 创建命令快捷方式 |
| `removeShim(name)` | 删除命令快捷方式 |
| `persistData(appName, paths)` | 持久化应用数据 |
| `createShortcut(options)` | 创建快捷方式 |
| `getCacheDir()` | 获取缓存目录 |
| `getConfigDir()` | 获取配置目录 |
| `deleteEnv(key, scope?)` | 删除环境变量 |
| `getPath()` | 获取 PATH 环境变量 |
| `getShimDir()` | 获取 Shim 目录 |
| `getPersistDir()` | 获取持久化数据目录 |

### getCookDir(appName, version?)

获取应用的安装目录。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `appName` | string | 是 | 应用名称 |
| `version` | string | 否 | 版本号，默认当前版本 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `path` | string | 安装目录路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.getCookDir("git", "2.43.0");
if (result.success) {
    logx.info("安装目录: " + result.path);
}
```

### getCurrentVersion()

获取 Chopsticks 当前版本。

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `version` | string | 版本号 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.getCurrentVersion();
if (result.success) {
    logx.info("Chopsticks 版本: " + result.version);
}
```

### addToPath(dirPath, scope?)

添加目录到 PATH 环境变量。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `dirPath` | string | 是 | 要添加的目录 |
| `scope` | string | 否 | 作用域 (user/machine)，默认 user |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.addToPath("C:\\Program Files\\MyApp\\bin", "user");
if (result.success) {
    logx.info("已添加到 PATH");
}
```

### removeFromPath(dirPath, scope?)

从 PATH 环境变量移除目录。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `dirPath` | string | 是 | 要移除的目录 |
| `scope` | string | 否 | 作用域 (user/machine)，默认 user |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
chopsticksx.removeFromPath("C:\\OldApp\\bin", "user");
```

### setEnv(key, value, scope?)

设置环境变量。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `key` | string | 是 | 变量名 |
| `value` | string | 是 | 变量值 |
| `scope` | string | 否 | 作用域 (user/machine)，默认 user |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
chopsticksx.setEnv("JAVA_HOME", "C:\\Program Files\\Java\\jdk-17", "machine");
```

### getEnv(key)

获取环境变量值。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `key` | string | 是 | 变量名 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `value` | string | 变量值 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.getEnv("PATH");
if (result.success) {
    logx.info("PATH: " + result.value);
}
```

### createShim(target, name?, options?)

创建命令快捷方式（Shim）。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `target` | string | 是 | 目标可执行文件路径 |
| `name` | string | 否 | 快捷方式名称，默认使用目标文件名 |
| `options` | object | 否 | 选项 |

**options 选项：**

| 选项 | 类型 | 说明 |
|------|------|------|
| `args` | array | 默认参数 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `shimPath` | string | 创建的快捷方式路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.createShim(
    "C:\\Program Files\\Node\\node.exe",
    "node"
);
if (result.success) {
    logx.info("Shim 创建成功: " + result.shimPath);
}
```

### removeShim(name)

删除命令快捷方式。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 快捷方式名称 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
chopsticksx.removeShim("node");
```

### persistData(appName, paths)

持久化应用数据（在更新时保留）。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `appName` | string | 是 | 应用名称 |
| `paths` | array | 是 | 要持久化的路径列表 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
chopsticksx.persistData("vscode", ["data", "config/settings.json"]);
```

### createShortcut(options)

创建 Windows 快捷方式。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `options` | object | 是 | 快捷方式选项 |

**options 选项：**

| 选项 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `source` | string | 是 | 目标文件路径 |
| `name` | string | 是 | 快捷方式名称 |
| `description` | string | 否 | 描述 |
| `icon` | string | 否 | 图标路径 |
| `workingDir` | string | 否 | 工作目录 |
| `arguments` | string | 否 | 启动参数 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `shortcutPath` | string | 快捷方式路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.createShortcut({
    source: "C:\\Program Files\\App\\app.exe",
    name: "My Application",
    description: "My App Description",
    icon: "C:\\Program Files\\App\\app.ico",
    workingDir: "C:\\Program Files\\App",
    arguments: "--start"
});
```

### getCacheDir()

获取 Chopsticks 缓存目录。

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `path` | string | 缓存目录路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.getCacheDir();
if (result.success) {
    logx.info("缓存目录: " + result.path);
}
```

### getConfigDir()

获取 Chopsticks 配置目录。

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `path` | string | 配置目录路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.getConfigDir();
if (result.success) {
    logx.info("配置目录: " + result.path);
}
```

### deleteEnv(key, scope?)

删除环境变量。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `key` | string | 是 | 变量名 |
| `scope` | string | 否 | 作用域 (user/machine)，默认 user |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
chopsticksx.deleteEnv("OLD_VAR", "user");
```

### getPath()

获取 PATH 环境变量值。

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `paths` | array | PATH 中的目录列表 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.getPath();
if (result.success) {
    for (const p of result.paths) {
        logx.info("PATH 项: " + p);
    }
}
```

### getShimDir()

获取 Shim 目录路径。

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `path` | string | Shim 目录路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.getShimDir();
if (result.success) {
    logx.info("Shim 目录: " + result.path);
}
```

### getPersistDir()

获取持久化数据目录。

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `path` | string | 持久化目录路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = chopsticksx.getPersistDir();
if (result.success) {
    logx.info("持久化目录: " + result.path);
}
```

---

## 7. jsonx - JSON 处理

JSON 处理模块，提供序列化和解析功能。

### 函数列表

| 函数 | 说明 |
|------|------|
| `stringify(value, space?)` | 将对象序列化为 JSON 字符串 |
| `parse(text)` | 将 JSON 字符串解析为对象 |

### stringify(value, space?)

将 JavaScript 对象序列化为 JSON 字符串。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `value` | any | 是 | 要序列化的值 |
| `space` | number/string | 否 | 缩进空格数或字符串 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `json` | string | JSON 字符串 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const data = { name: "test", version: "1.0.0", items: [1, 2, 3] };

// 紧凑格式
const result1 = jsonx.stringify(data);
// 结果: {"name":"test","version":"1.0.0","items":[1,2,3]}

// 格式化输出
const result2 = jsonx.stringify(data, 2);
// 结果:
// {
//   "name": "test",
//   "version": "1.0.0",
//   "items": [1, 2, 3]
// }
```

### parse(text)

将 JSON 字符串解析为 JavaScript 对象。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `text` | string | 是 | JSON 字符串 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `data` | any | 解析后的对象 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const jsonStr = '{"name":"test","version":"1.0.0"}';
const result = jsonx.parse(jsonStr);
if (result.success) {
    logx.info("名称: " + result.data.name);
    logx.info("版本: " + result.data.version);
}

// 解析失败处理
const result2 = jsonx.parse("invalid json");
if (!result2.success) {
    logx.error("解析失败: " + result2.error);
}
```

---

## 8. logx - 日志

日志模块，提供分级日志输出功能。

### 函数列表

| 函数 | 说明 |
|------|------|
| `debug(message)` | 输出调试日志 |
| `info(message)` | 输出信息日志 |
| `warn(message)` | 输出警告日志 |
| `error(message)` | 输出错误日志 |

### debug(message)

输出调试日志（仅在调试模式显示）。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `message` | string | 是 | 日志消息 |

**返回值：** 无

**示例：**

```javascript
logx.debug("正在检查版本...");
logx.debug("变量值: " + JSON.stringify(data));
```

### info(message)

输出信息日志。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `message` | string | 是 | 日志消息 |

**返回值：** 无

**示例：**

```javascript
logx.info("开始安装应用...");
logx.info("下载完成，正在解压...");
```

### warn(message)

输出警告日志。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `message` | string | 是 | 日志消息 |

**返回值：** 无

**示例：**

```javascript
logx.warn("配置文件已存在，将被覆盖");
logx.warn("检测到旧版本，建议升级");
```

### error(message)

输出错误日志。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `message` | string | 是 | 日志消息 |

**返回值：** 无

**示例：**

```javascript
logx.error("下载失败: 网络连接错误");
logx.error("安装失败: " + error.message);
```

---

## 9. pathx - 路径操作

路径操作模块，提供跨平台的路径处理功能。

### 函数列表

| 函数 | 说明 |
|------|------|
| `join(...paths)` | 连接多个路径片段 |
| `abs(path)` | 转换为绝对路径 |
| `base(path, ext?)` | 获取文件名 |
| `dir(path)` | 获取目录名 |
| `ext(path)` | 获取扩展名 |
| `clean(path)` | 清理路径 |
| `isAbs(path)` | 检查是否为绝对路径 |
| `exists(path)` | 检查路径是否存在 |
| `isDir(path)` | 检查是否为目录 |

### join(...paths)

连接多个路径片段。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `...paths` | string... | 是 | 路径片段 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `path` | string | 连接后的路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = pathx.join("C:\\Users", "name", "Documents", "file.txt");
// 结果: C:\\Users\\name\\Documents\\file.txt

const result2 = pathx.join("/home", "user", "docs");
// 结果: /home/user/docs
```

### abs(path)

将相对路径转换为绝对路径。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `path` | string | 绝对路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = pathx.abs("./config.json");
// 结果: C:\\Current\\Directory\\config.json
```

### base(path, ext?)

获取路径的文件名部分。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 路径 |
| `ext` | string | 否 | 要移除的扩展名 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `name` | string | 文件名 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result1 = pathx.base("/path/to/file.txt");
// 结果: file.txt

const result2 = pathx.base("/path/to/file.txt", ".txt");
// 结果: file
```

### dir(path)

获取路径的目录部分。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `dir` | string | 目录路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = pathx.dir("/path/to/file.txt");
// 结果: /path/to
```

### ext(path)

获取路径的扩展名。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `ext` | string | 扩展名（包含点） |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = pathx.ext("/path/to/file.txt");
// 结果: .txt

const result2 = pathx.ext("/path/to/archive.tar.gz");
// 结果: .gz
```

### clean(path)

清理路径，移除冗余的 `.` 和 `..`。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `path` | string | 清理后的路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = pathx.clean("/path/./to/../file.txt");
// 结果: /path/file.txt
```

### isAbs(path)

检查路径是否为绝对路径。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `isAbs` | boolean | 是否为绝对路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result1 = pathx.isAbs("/usr/bin");
// result.isAbs: true

const result2 = pathx.isAbs("./relative/path");
// result.isAbs: false
```

### exists(path)

检查路径是否存在。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `exists` | boolean | 是否存在 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = pathx.exists("/path/to/check");
if (result.success && result.exists) {
    logx.info("路径存在");
}
```

### isDir(path)

检查路径是否为目录。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `isDir` | boolean | 是否为目录 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = pathx.isDir("/path/to/check");
if (result.success && result.isDir) {
    logx.info("这是一个目录");
}
```

---

## 10. registry - 注册表操作

Windows 注册表操作模块。

### 函数列表

| 函数 | 说明 |
|------|------|
| `setValue(keyPath, name, value, type?)` | 设置注册表值 |
| `getValue(keyPath, name)` | 获取注册表值 |
| `setDword(keyPath, name, value)` | 设置 DWORD 值 |
| `getDword(keyPath, name)` | 获取 DWORD 值 |
| `deleteValue(keyPath, name)` | 删除注册表值 |
| `createKey(keyPath)` | 创建注册表键 |
| `deleteKey(keyPath)` | 删除注册表键 |
| `keyExists(keyPath)` | 检查键是否存在 |
| `listKeys(keyPath)` | 列出子键 |
| `listValues(keyPath)` | 列出值 |

### setValue(keyPath, name, value, type?)

设置注册表字符串值。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径，如 `HKCU\\Software\\App` |
| `name` | string | 是 | 值名称 |
| `value` | string | 是 | 值内容 |
| `type` | string | 否 | 值类型，默认 `REG_SZ` |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = registry.setValue(
    "HKCU\\Software\\MyApp",
    "Version",
    "1.0.0"
);
```

### getValue(keyPath, name)

获取注册表值。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径 |
| `name` | string | 是 | 值名称 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `value` | string | 值内容 |
| `type` | string | 值类型 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = registry.getValue("HKCU\\Software\\MyApp", "Version");
if (result.success) {
    logx.info("版本: " + result.value);
}
```

### setDword(keyPath, name, value)

设置 DWORD 值。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径 |
| `name` | string | 是 | 值名称 |
| `value` | number | 是 | DWORD 值 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
registry.setDword("HKCU\\Software\\MyApp", "Count", 42);
```

### getDword(keyPath, name)

获取 DWORD 值。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径 |
| `name` | string | 是 | 值名称 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `value` | number | DWORD 值 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = registry.getDword("HKCU\\Software\\MyApp", "Count");
if (result.success) {
    logx.info("计数: " + result.value);
}
```

### deleteValue(keyPath, name)

删除注册表值。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径 |
| `name` | string | 是 | 值名称 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
registry.deleteValue("HKCU\\Software\\MyApp", "OldValue");
```

### createKey(keyPath)

创建注册表键。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
registry.createKey("HKCU\\Software\\MyApp\\Settings");
```

### deleteKey(keyPath)

删除注册表键。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
registry.deleteKey("HKCU\\Software\\MyApp\\OldSettings");
```

### keyExists(keyPath)

检查注册表键是否存在。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `exists` | boolean | 是否存在 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = registry.keyExists("HKCU\\Software\\MyApp");
if (result.success && result.exists) {
    logx.info("键存在");
}
```

### listKeys(keyPath)

列出指定键下的所有子键。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `keys` | array | 子键名称列表 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = registry.listKeys("HKCU\\Software");
if (result.success) {
    for (const key of result.keys) {
        logx.info("子键: " + key);
    }
}
```

### listValues(keyPath)

列出指定键下的所有值。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyPath` | string | 是 | 键路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `values` | array | 值信息列表（包含 name, type, value） |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = registry.listValues("HKCU\\Software\\MyApp");
if (result.success) {
    for (const val of result.values) {
        logx.info(val.name + " = " + val.value);
    }
}
```

---

## 11. semver - 版本控制

语义化版本控制模块，提供版本解析、比较等功能。

### 函数列表

| 函数 | 说明 |
|------|------|
| `parse(version)` | 解析版本字符串 |
| `compare(v1, v2)` | 比较两个版本 |
| `gt(v1, v2)` | 检查 v1 是否大于 v2 |
| `lt(v1, v2)` | 检查 v1 是否小于 v2 |
| `eq(v1, v2)` | 检查两个版本是否相等 |
| `gte(v1, v2)` | 检查 v1 是否大于等于 v2 |
| `lte(v1, v2)` | 检查 v1 是否小于等于 v2 |
| `satisfies(version, range)` | 检查版本是否满足范围 |

### parse(version)

解析版本字符串为结构化对象。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `version` | string | 是 | 版本字符串 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `major` | number | 主版本号 |
| `minor` | number | 次版本号 |
| `patch` | number | 修订号 |
| `prerelease` | string | 预发布标识 |
| `build` | string | 构建元数据 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = semver.parse("1.2.3-beta.1+build.123");
if (result.success) {
    logx.info("主版本: " + result.major);
    logx.info("次版本: " + result.minor);
    logx.info("修订号: " + result.patch);
}
```

### compare(v1, v2)

比较两个版本。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `v1` | string | 是 | 版本1 |
| `v2` | string | 是 | 版本2 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `result` | number | 比较结果：-1(v1<v2), 0(v1=v2), 1(v1>v2) |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = semver.compare("1.2.0", "1.2.3");
if (result.success) {
    if (result.result < 0) {
        logx.info("1.2.0 < 1.2.3");
    } else if (result.result > 0) {
        logx.info("1.2.0 > 1.2.3");
    } else {
        logx.info("版本相等");
    }
}
```

### gt(v1, v2)

检查 v1 是否大于 v2。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `v1` | string | 是 | 版本1 |
| `v2` | string | 是 | 版本2 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `result` | boolean | v1 是否大于 v2 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = semver.gt("2.0.0", "1.9.9");
if (result.success && result.result) {
    logx.info("2.0.0 > 1.9.9");
}
```

### lt(v1, v2)

检查 v1 是否小于 v2。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `v1` | string | 是 | 版本1 |
| `v2` | string | 是 | 版本2 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `result` | boolean | v1 是否小于 v2 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = semver.lt("1.0.0", "2.0.0");
if (result.success && result.result) {
    logx.info("1.0.0 < 2.0.0");
}
```

### eq(v1, v2)

检查两个版本是否相等。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `v1` | string | 是 | 版本1 |
| `v2` | string | 是 | 版本2 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `result` | boolean | 是否相等 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = semver.eq("1.2.3", "1.2.3");
if (result.success && result.result) {
    logx.info("版本相等");
}
```

### gte(v1, v2)

检查 v1 是否大于等于 v2。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `v1` | string | 是 | 版本1 |
| `v2` | string | 是 | 版本2 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `result` | boolean | v1 是否大于等于 v2 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = semver.gte("1.2.3", "1.0.0");
if (result.success && result.result) {
    logx.info("1.2.3 >= 1.0.0");
}
```

### lte(v1, v2)

检查 v1 是否小于等于 v2。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `v1` | string | 是 | 版本1 |
| `v2` | string | 是 | 版本2 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `result` | boolean | v1 是否小于等于 v2 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = semver.lte("1.0.0", "1.2.3");
if (result.success && result.result) {
    logx.info("1.0.0 <= 1.2.3");
}
```

### satisfies(version, range)

检查版本是否满足指定范围。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `version` | string | 是 | 版本 |
| `range` | string | 是 | 版本范围 |

**范围格式：**

| 格式 | 说明 | 示例 |
|------|------|------|
| `^x.y.z` | 兼容版本，允许次版本和修订号更新 | `^1.2.3` 匹配 `>=1.2.3 <2.0.0` |
| `~x.y.z` | 近似版本，允许修订号更新 | `~1.2.3` 匹配 `>=1.2.3 <1.3.0` |
| `>=x.y.z` | 大于等于 | `>=1.0.0` |
| `<=x.y.z` | 小于等于 | `<=2.0.0` |
| `x.y.z - a.b.c` | 范围 | `1.0.0 - 2.0.0` |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `result` | boolean | 是否满足范围 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result1 = semver.satisfies("1.2.3", "^1.0.0");
// result1.result: true

const result2 = semver.satisfies("2.0.0", "^1.0.0");
// result2.result: false

const result3 = semver.satisfies("1.2.5", ">=1.2.0 <1.3.0");
// result3.result: true
```

---

## 12. symlink - 符号链接

符号链接操作模块，支持 Windows 和类 Unix 系统。

### 函数列表

| 函数 | 说明 |
|------|------|
| `create(target, linkPath)` | 创建文件符号链接 |
| `createDir(target, linkPath)` | 创建目录符号链接 |
| `createHard(target, linkPath)` | 创建硬链接 |
| `createJunction(target, linkPath)` | 创建目录联接（Windows） |
| `is(path)` | 检查是否为符号链接 |
| `read(linkPath)` | 读取符号链接目标 |
| `remove(linkPath)` | 删除符号链接 |

### create(target, linkPath)

创建文件符号链接。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `target` | string | 是 | 目标文件路径 |
| `linkPath` | string | 是 | 链接路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = symlink.create(
    "C:\\Program Files\\App\\app.exe",
    "C:\\Links\\app.exe"
);
```

### createDir(target, linkPath)

创建目录符号链接。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `target` | string | 是 | 目标目录路径 |
| `linkPath` | string | 是 | 链接路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = symlink.createDir(
    "C:\\Program Files\\Node",
    "C:\\Links\\node"
);
```

### createHard(target, linkPath)

创建硬链接。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `target` | string | 是 | 目标文件路径 |
| `linkPath` | string | 是 | 链接路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = symlink.createHard("original.txt", "hardlink.txt");
```

### createJunction(target, linkPath)

创建 Windows 目录联接（Junction）。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `target` | string | 是 | 目标目录路径 |
| `linkPath` | string | 是 | 联接路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = symlink.createJunction(
    "D:\\Data\\Documents",
    "C:\\Users\\Name\\Documents"
);
```

### is(path)

检查路径是否为符号链接。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `isSymlink` | boolean | 是否为符号链接 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = symlink.is("C:\\Links\\app.exe");
if (result.success && result.isSymlink) {
    logx.info("这是一个符号链接");
}
```

### read(linkPath)

读取符号链接指向的目标。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `linkPath` | string | 是 | 符号链接路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `target` | string | 目标路径 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = symlink.read("C:\\Links\\app.exe");
if (result.success) {
    logx.info("目标: " + result.target);
}
```

### remove(linkPath)

删除符号链接。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `linkPath` | string | 是 | 符号链接路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
symlink.remove("C:\\Links\\app.exe");
```

---

## 13. installerx - 安装程序

安装程序执行模块，支持多种安装包格式。

### 函数列表

| 函数 | 说明 |
|------|------|
| `run(installerPath, args?, options?)` | 自动检测类型并运行安装程序 |
| `runNSIS(installerPath, args?, options?)` | 运行 NSIS 安装程序 |
| `runMSI(msiPath, args?, options?)` | 运行 MSI 安装程序 |
| `runInno(installerPath, args?, options?)` | 运行 Inno Setup 安装程序 |
| `detectType(installerPath)` | 检测安装程序类型 |

### run(installerPath, args?, options?)

自动检测安装程序类型并运行。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `installerPath` | string | 是 | 安装程序路径 |
| `args` | array | 否 | 安装参数 |
| `options` | object | 否 | 执行选项 |

**options 选项：**

| 选项 | 类型 | 说明 |
|------|------|------|
| `wait` | boolean | 是否等待完成 |
| `timeout` | number | 超时时间（毫秒） |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `exitCode` | number | 退出码 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = installerx.run("setup.exe", ["/S"], { wait: true });
if (result.success) {
    logx.info("安装完成，退出码: " + result.exitCode);
}
```

### runNSIS(installerPath, args?, options?)

运行 NSIS 安装程序。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `installerPath` | string | 是 | 安装程序路径 |
| `args` | array | 否 | 安装参数 |
| `options` | object | 否 | 执行选项 |

**常用 NSIS 参数：**

| 参数 | 说明 |
|------|------|
| `/S` | 静默安装 |
| `/D=path` | 指定安装目录 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `exitCode` | number | 退出码 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = installerx.runNSIS("installer.exe", [
    "/S",
    "/D=C:\\Program Files\\MyApp"
]);
```

### runMSI(msiPath, args?, options?)

运行 MSI 安装程序。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `msiPath` | string | 是 | MSI 文件路径 |
| `args` | array | 否 | 安装参数 |
| `options` | object | 否 | 执行选项 |

**常用 MSI 参数：**

| 参数 | 说明 |
|------|------|
| `/quiet` | 静默安装 |
| `/norestart` | 安装后不重启 |
| `/log file.log` | 记录日志 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `exitCode` | number | 退出码 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = installerx.runMSI("installer.msi", [
    "/quiet",
    "/norestart"
]);
```

### runInno(installerPath, args?, options?)

运行 Inno Setup 安装程序。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `installerPath` | string | 是 | 安装程序路径 |
| `args` | array | 否 | 安装参数 |
| `options` | object | 否 | 执行选项 |

**常用 Inno Setup 参数：**

| 参数 | 说明 |
|------|------|
| `/VERYSILENT` | 完全静默安装 |
| `/SILENT` | 静默安装（显示进度） |
| `/SUPPRESSMSGBOXES` | 抑制消息框 |
| `/NORESTART` | 不重启 |
| `/DIR="path"` | 指定安装目录 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `exitCode` | number | 退出码 |
| `error` | string | 错误信息（失败时） |

**示例：**

```javascript
const result = installerx.runInno("setup.exe", [
    "/VERYSILENT",
    "/SUPPRESSMSGBOXES",
    "/NORESTART",
    '/DIR="C:\\Program Files\\MyApp"'
]);
```

### detectType(installerPath)

检测安装程序类型。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `installerPath` | string | 是 | 安装程序路径 |

**返回值：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | boolean | 是否成功 |
| `type` | string | 安装程序类型 |
| `error` | string | 错误信息（失败时） |

**可能的类型：**

| 类型 | 说明 |
|------|------|
| `nsis` | NSIS 安装程序 |
| `msi` | Windows Installer (MSI) |
| `inno` | Inno Setup |
| `autoit` | AutoIt 安装程序 |
| `unknown` | 未知类型 |

**示例：**

```javascript
const result = installerx.detectType("setup.exe");
if (result.success) {
    logx.info("安装程序类型: " + result.type);
}
```

---

## 附录：返回值规范

所有 API 统一返回以下格式的结果对象：

### 成功响应

```javascript
{
    success: true,
    // 其他数据字段...
    error: null
}
```

### 失败响应

```javascript
{
    success: false,
    // 数据字段通常为 null
    error: "错误描述信息"
}
```

### 使用模式

```javascript
const result = someApi.someFunction();

if (!result.success) {
    logx.error("操作失败: " + result.error);
    return;
}

// 处理成功结果
logx.info("操作成功: " + result.someData);
```

---

_最后更新：2026-03-01_
_版本：v0.10.0-alpha_
