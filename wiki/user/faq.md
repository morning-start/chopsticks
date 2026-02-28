# 常见问题解答

> Chopsticks 常见问题与解决方案

---

## 1. 基础问题

### 什么是 Chopsticks

Chopsticks（筷子）是一个受 Scoop 启发的 Windows 包管理器，采用 Go 语言开发，使用 JavaScript 作为包定义脚本语言。它可以帮助您快速安装、管理和更新 Windows 软件。

### Chopsticks 与其他包管理器的区别

Chopsticks 具有以下特点：

- **绿色环保**：无需管理员权限，软件安装在用户目录
- **脚本驱动**：使用 JavaScript 定义安装逻辑，灵活可控
- **开放架构**：支持自定义软件源（Bucket）
- **跨设备同步**：支持设备间软件配置同步

### 支持哪些操作系统

目前 Chopsticks 主要支持 Windows 10 及以上版本。Linux 和 macOS 支持正在规划中。

---

## 2. 安装问题

### 如何安装 Chopsticks

```powershell
git clone https://github.com/chopsticks-bucket/main.git
cd main
go build -o chopsticks.exe
```

### 安装失败怎么办

1. 检查 Go 环境：确保已安装 Go 1.21 或更高版本
2. 检查网络：确保可以访问 GitHub
3. 使用调试模式：

```bash
chopsticks install <package> --debug
```

### 如何验证安装成功

```bash
chopsticks --version
chopsticks --help
```

---

## 3. 软件源（Bucket）问题

### 什么是软件源

软件源（Bucket）是软件的集合，类似于 apt 的 PPA 或 yum 的仓库。每个软件源包含多个软件包（App）的定义。

### 如何添加软件源

```bash
chopsticks bucket add main https://github.com/chopsticks-bucket/main
```

### 软件源加载失败怎么办

1. 检查网络连接
2. 确认仓库地址正确
3. 尝试指定分支：

```bash
chopsticks bucket add main https://github.com/chopsticks-bucket/main --branch main
```

### 如何创建自己的软件源

请参考《Bucket 创建指南》文档。

### 可以同时使用多个软件源吗

可以。您可以添加多个软件源：

```bash
chopsticks bucket add main https://github.com/chopsticks-bucket/main
chopsticks bucket add extras https://github.com/chopsticks-bucket/extras
```

---

## 4. 软件包（App）问题

### 如何安装软件

```bash
chopsticks install git
chopsticks install vscode
chopsticks install nodejs@18.17.0
```

### 如何指定软件版本

使用 @ 符号指定版本：

```bash
chopsticks install nodejs@18.17.0
chopsticks install python@3.12.0
```

### 如何更新软件

```bash
chopsticks update git
chopsticks update --all
```

### 如何卸载软件

```bash
chopsticks uninstall git
chopsticks uninstall git --purge  # 完全删除，包括配置
```

### 软件安装失败怎么办

1. 使用详细模式查看错误：

```bash
chopsticks install <package> --verbose
```

2. 常见解决方案：
   - 检查网络连接
   - 确认软件源可用
   - 尝试强制安装：`chopsticks install <package> --force`

### 如何搜索软件

```bash
chopsticks search vscode
chopsticks search git --bucket main  # 在指定软件源搜索
```

---

## 5. 使用问题

### 软件安装在哪里

默认安装在用户目录下：

```
%USERPROFILE%\.chopsticks\apps\
```

### 如何查看已安装的软件

```bash
chopsticks list
chopsticks ls
```

### 如何查看软件信息

```bash
chopsticks info git
```

### 如何设置环境变量

在 App 脚本中使用 chopsticks 模块：

```javascript
await chopsticks.setEnv("MY_VAR", "value");
await chopsticks.addToPath("path/to/bin");
```

### PATH 环境变量不生效怎么办

1. 重新打开终端
2. 检查 App 脚本是否正确配置了 env_path()
3. 手动刷新环境变量：

```bash
# PowerShell
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","User") + ";" + [System.Environment]::GetEnvironmentVariable("Path","Machine")

# CMD
refreshenv
```

---

## 6. 故障排除

### 命令执行超时

网络环境不佳时，可以尝试：

1. 使用镜像源（如果有）
2. 增加超时时间（在配置文件中设置）
3. 检查防火墙和代理设置

### 数据库错误

如果遇到数据库错误，可以尝试：

```bash
# 备份现有数据
cp ~/.chopsticks/data.db ~/.chopsticks/data.db.bak

# 删除数据库（会丢失已安装软件记录）
rm ~/.chopsticks/data.db
```

### 权限问题

Chopsticks 默认不需要管理员权限。如果遇到权限错误：

1. 确保用户目录可写
2. 检查杀毒软件是否拦截
3. 尝试以管理员身份运行（不推荐）

### 缓存问题

清理下载缓存：

```bash
# 清理所有缓存
chopsticks cache clean

# 清理指定软件缓存
chopsticks cache clean git
```

### 日志文件位置

日志文件位于：

```
%USERPROFILE%\.chopsticks\logs\
```

---

## 7. 高级问题

### 如何创建绿色软件的 App 脚本

请参考《App 最佳实践》文档中的"绿色软件模式"章节。

### 如何处理安装程序

对于需要运行安装程序的软件，使用 installer 模块：

```javascript
async onPostExtract(ctx) {
    await installer.run("setup.exe", ["/S", "/D=" + ctx.cookDir]);
}
```

### 如何处理依赖关系

在 App 脚本中声明依赖：

```javascript
depends() {
    return ["nodejs", "python"];
}
```

### 如何处理多架构

在 getDownloadInfo 中根据 arch 参数返回不同的下载链接：

```javascript
async getDownloadInfo(version, arch) {
    const archMap = {
        amd64: "x64",
        x86: "x86"
    };
    // ...
}
```

### 如何迁移已安装的软件

使用同步功能：

```bash
# 在新设备上
chopsticks sync install
```

详细说明请参考《用户指南》中的设备同步章节。

---

## 8. 错误代码

查看完整的错误代码列表和解决方案，请访问 [错误代码参考](error-codes.md)。

### 常见错误代码快速参考

| 错误代码 | 说明 | 解决方案 |
|---------|------|----------|
| CHP-1001 | 权限不足 | 以管理员身份运行或修改安装目录 |
| CHP-2001 | 网络连接失败 | 检查网络连接或配置代理 |
| CHP-3001 | 软件源不存在 | 检查软件源名称或添加正确的软件源 |
| CHP-4001 | 软件未找到 | 检查软件名称或搜索正确的软件包 |
| CHP-4003 | Hash 校验失败 | 清理缓存后重新安装 |
| CHP-5001 | 配置文件格式错误 | 重置配置或手动修复配置文件 |

---

## 9. 相关链接

- [Bucket 创建指南](bucket-guide.md)
- [App 最佳实践](app-best-practices.md)
- [用户指南](../wiki/USAGE.md)
- [API 参考](../wiki/API.md)
- [术语表](../wiki/GLOSSARY.md)

---

## 10. 获取帮助

如果以上方法无法解决问题：

1. 查看日志文件获取详细错误信息
2. 使用 --debug 模式重新执行命令
3. 在 GitHub 仓库提交 Issue
4. 加入社区讨论

---

_最后更新：2026-02-28_
