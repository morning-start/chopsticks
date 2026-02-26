# Chopsticks 用户指南

> 详细的使用说明和示例

---

## 1. 安装与配置

### 1.1 安装 Chopsticks

```powershell
# 克隆项目
git clone https://github.com/chopsticks-bows/main.git
cd main

# 编译
go build -o chopsticks.exe

# 验证
./chopsticks.exe --help
```

### 1.2 环境配置

Chopsticks 默认使用以下目录：

| 目录       | 环境变量          | 默认路径                            |
| ---------- | ----------------- | ----------------------------------- |
| 安装目录   | `CHOPSTICKS_HOME` | `%USERPROFILE%\.chopsticks`         |
| 应用目录   | -                 | `%USERPROFILE%\.chopsticks\apps`    |
| 缓存目录   | -                 | `%USERPROFILE%\.chopsticks\cache`   |
| 软件源目录 | -                 | `%USERPROFILE%\.chopsticks\sources` |

---

## 2. 软件源管理

### 2.1 添加软件源

```bash
# 添加远程软件源
chopsticks source add main https://github.com/chopsticks-bows/main

# 添加本地软件源
chopsticks source add local /path/to/local-source

# 指定分支
chopsticks source add extras https://github.com/chopsticks-bows/extras --branch develop
```

### 2.2 列出软件源

```bash
# 列出所有软件源
chopsticks source list
chopsticks source ls

# 使用简写
chopsticks s ls
```

### 2.3 更新软件源

```bash
# 更新所有软件源
chopsticks source update

# 更新指定软件源
chopsticks source update main
```

### 2.4 删除软件源

```bash
# 删除软件源
chopsticks source remove extras

# 删除并清理本地数据
chopsticks source remove extras --purge
```

---

## 3. 软件包管理

### 3.1 安装软件

```bash
# 安装最新版本
chopsticks install git

# 安装指定版本
chopsticks install nodejs@18.17.0
chopsticks install python@3.12.0

# 从指定软件源安装
chopsticks install extras/vscode

# 强制重新安装
chopsticks install git --force

# 使用别名
chopsticks serve git
chopsticks i git
```

### 3.2 卸载软件

```bash
# 卸载软件（保留配置数据）
chopsticks uninstall git
chopsticks clear git

# 彻底卸载（删除所有数据）
chopsticks uninstall git --purge
chopsticks rm git
```

### 3.3 更新软件

```bash
# 更新指定软件
chopsticks update git
chopsticks refresh git

# 更新所有软件
chopsticks update --all
chopsticks refresh --all

# 强制更新
chopsticks update git --force
```

### 3.4 查看软件

```bash
# 列出已安装软件
chopsticks list
chopsticks ls

# 仅显示已安装
chopsticks list --installed

# 搜索软件
chopsticks search vscode
chopsticks search editor

# 在指定软件源搜索
chopsticks search vscode --bucket extras
```

---

## 4. 命令别名

Chopsticks 支持多种命令格式，您可以自由选择：

| 主命令      | 别名                       | 说明       |
| ----------- | -------------------------- | ---------- |
| `install`   | `serve`, `i`               | 安装软件   |
| `uninstall` | `clear`, `remove`, `rm`    | 卸载软件   |
| `update`    | `refresh`, `upgrade`, `up` | 更新软件   |
| `search`    | `find`, `s`                | 搜索软件   |
| `list`      | `ls`                       | 列出软件   |
| `source`    | `bucket`, `s`              | 软件源管理 |
| `sync`      | -                          | 设备同步   |

---

## 5. Shell 自动补全

### 5.1 Bash

```bash
# 临时启用
source <(chopsticks completion bash)

# 永久启用
chopsticks completion bash >> ~/.bashrc
```

### 5.2 Zsh

```zsh
# 临时启用
source <(chopsticks completion zsh)

# 永久启用
chopsticks completion zsh >> ~/.zshrc
```

### 5.3 PowerShell

```powershell
# 临时启用
chopsticks completion powershell | Invoke-Expression

# 永久启用
chopsticks completion powershell >> $PROFILE
```

### 5.4 Fish

```fish
# 永久启用
chopsticks completion fish > ~/.config/fish/completions/chopsticks.fish
```

---

## 6. 高级用法

### 6.1 版本指定

```bash
# 使用 @ 指定版本
chopsticks install nodejs@18.17.0
chopsticks install python@3.11.5

# 不指定版本安装最新稳定版
chopsticks install git
```

### 6.2 架构指定

```bash
# 指定架构 (需要软件包支持)
chopsticks install app --arch amd64
chopsticks install app --arch x86
```

### 6.3 详细输出

```bash
# 详细模式
chopsticks install git --verbose

# 调试模式
chopsticks install git --debug
```

---

## 7. 故障排除

### 7.1 常见问题

**安装失败**

```bash
# 使用 --verbose 查看详细错误
chopsticks install git --verbose

# 使用 --debug 查看调试信息
chopsticks install git --debug
```

**网络问题**

```bash
# 检查网络连接
chopsticks source update --verbose
```

**权限问题**

```bash
# 确保有写入权限
chopsticks list --verbose
```

### 7.2 清理缓存

```bash
# 清理下载缓存
# (待实现)
chopsticks cache clean
```

---

## 8. 配置文件

### 8.1 配置文件位置

- Windows: `%USERPROFILE%\.chopsticks\config.yaml`
- Linux/macOS: `~/.chopsticks/config.yaml`

### 8.2 配置示例

```yaml
# 目录配置
apps_path: "C:\\Users\\Username\\.chopsticks\\apps"
cache_path: "C:\\Users\\Username\\.chopsticks\\cache"
sources_path: "C:\\Users\\Username\\.chopsticks\\sources"

# 行为配置
auto_update: true
verify_checksum: true

# 网络配置
timeout: 30
retry: 3
```

---

## 9. 设备同步

### 9.1 功能概述

设备同步功能允许用户快速在新设备上恢复所有已安装的软件。当您需要更换电脑或重新安装系统时，只需复制整个 `.chopsticks` 目录到新设备，然后运行同步命令即可。

### 9.2 目录结构

```
%USERPROFILE%\.chopsticks\
├── sources/           # 软件源
├── apps/              # 已安装的软件
├── cache/             # 下载缓存
├── data.db            # 全局数据库（包含已安装软件和软件源配置）
```

### 9.3 使用场景

**场景一：换电脑**

1. 在旧电脑上，将 `%USERPROFILE%\.chopsticks` 目录复制到 U 盘或云盘
2. 在新电脑上，将目录复制到相同位置 `%USERPROFILE%\.chopsticks`
3. 运行同步命令恢复软件

**场景二：重装系统**

1. 重装系统前备份 `.chopsticks` 目录
2. 重装系统后，将备份目录恢复到 `%USERPROFILE%\.chopsticks`
3. 运行同步命令恢复软件

### 9.4 命令用法

```bash
# 查看将同步的软件列表（不实际安装）
chopsticks sync list

# 同步安装所有已记录的软件
chopsticks sync install

# 同步安装（跳过已安装的）
chopsticks sync install --skip-installed
```

### 9.5 工作原理

`sync install` 命令会执行以下操作：

1. 读取 `data.db` 中的 `buckets` 表，获取所有已添加的软件源
2. 读取 `data.db` 中的 `installed` 表，获取所有已安装的软件
3. 重新克隆软件源仓库（如果 sources 目录为空）
4. 遍历已安装列表，重新安装每个软件

### 9.6 注意事项

- **数据库完整**：确保复制的 `data.db` 数据库文件完整无损
- **网络连接**：同步过程需要重新下载软件，请确保网络畅通
- **版本兼容**：部分软件可能在新设备上需要不同版本，请注意检查

---

_最后更新：2026-02-26_
_版本：v0.1.0-alpha_
