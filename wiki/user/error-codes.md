# 错误代码参考

> Chopsticks 完整错误代码列表与解决方案

---

## 错误代码格式

Chopsticks 错误代码格式: `CHP-XXXX`

- `CHP`: Chopsticks 前缀
- `XXXX`: 4位数字错误码

---

## 1xxx - 系统错误

### CHP-1001: 权限不足

**错误消息**: `Permission denied: 需要管理员权限`

**原因**: 
- 尝试写入系统目录
- 没有足够的文件权限

**解决方案**:
```bash
# 方案1: 以管理员身份运行
# 右键 PowerShell -> 以管理员身份运行

# 方案2: 修改安装目录到用户目录
chopsticks config set install_dir "%USERPROFILE%\tools"
```

**预防措施**:
- 使用默认用户目录安装
- 避免修改系统目录

---

### CHP-1002: 磁盘空间不足

**错误消息**: `Insufficient disk space`

**原因**:
- 磁盘空间不足以下载或安装软件

**解决方案**:
```bash
# 清理缓存
chopsticks cache clean

# 查看磁盘空间
chopsticks doctor
```

---

## 2xxx - 网络错误

### CHP-2001: 网络连接失败

**错误消息**: `Network connection failed`

**原因**:
- 无法连接到软件源
- 代理配置错误

**解决方案**:
```bash
# 检查网络连接
ping github.com

# 配置代理
chopsticks config set proxy.http http://proxy.company.com:8080
chopsticks config set proxy.https https://proxy.company.com:8080
```

---

### CHP-2002: 下载超时

**错误消息**: `Download timeout`

**原因**:
- 网络速度慢
- 下载源响应慢

**解决方案**:
```bash
# 增加超时时间
chopsticks config set network.timeout 300

# 使用镜像源
chopsticks bucket add main-mirror https://mirror.example.com/main
```

---

## 3xxx - 软件源错误

### CHP-3001: 软件源不存在

**错误消息**: `Bucket not found: {bucket_name}`

**原因**:
- 软件源名称错误
- 软件源未添加

**解决方案**:
```bash
# 查看已添加的软件源
chopsticks bucket list

# 添加软件源
chopsticks bucket add main https://github.com/chopsticks-bucket/main
```

---

### CHP-3002: 软件源更新失败

**错误消息**: `Failed to update bucket: {bucket_name}`

**原因**:
- 网络问题
- 软件源仓库被删除

**解决方案**:
```bash
# 重新添加软件源
chopsticks bucket remove main
chopsticks bucket add main https://github.com/chopsticks-bucket/main
```

---

## 4xxx - 安装错误

### CHP-4001: 软件未找到

**错误消息**: `Package not found: {package_name}`

**原因**:
- 软件名称错误
- 软件源中不存在该软件

**解决方案**:
```bash
# 搜索软件
chopsticks search {keyword}

# 查看所有可用软件
chopsticks list --all
```

---

### CHP-4002: 安装脚本执行失败

**错误消息**: `Installation script failed`

**原因**:
- 脚本语法错误
- 依赖缺失

**解决方案**:
```bash
# 查看详细日志
chopsticks install {package} --verbose

# 手动执行安装步骤查看错误
```

---

### CHP-4003: Hash 校验失败

**错误消息**: `Hash verification failed`

**原因**:
- 下载文件损坏
- 文件被篡改

**解决方案**:
```bash
# 清理缓存重新下载
chopsticks cache clean --package {package}
chopsticks install {package}
```

---

## 5xxx - 配置错误

### CHP-5001: 配置文件格式错误

**错误消息**: `Invalid config file format`

**原因**:
- YAML 格式错误
- 配置项类型错误

**解决方案**:
```bash
# 重置配置
chopsticks config reset

# 或手动编辑配置文件
notepad %USERPROFILE%\.chopsticks\config.yaml
```

---

## 9xxx - 其他错误

### CHP-9001: 未知错误

**错误消息**: `Unknown error occurred`

**解决方案**:
```bash
# 查看详细日志
chopsticks doctor

# 重置 Chopsticks
chopsticks reset
```

---

## 快速排查指南

### 遇到错误时的排查步骤

1. **查看错误代码**: 记录完整的错误代码和消息
2. **查看日志**: 使用 `--verbose` 参数获取详细日志
3. **运行诊断**: 执行 `chopsticks doctor` 检查系统状态
4. **搜索文档**: 在本文档中查找错误代码
5. **清理重试**: 清理缓存后重试操作

### 常用诊断命令

```bash
# 系统诊断
chopsticks doctor

# 查看详细日志
chopsticks {command} --verbose

# 查看配置
chopsticks config list

# 清理缓存
chopsticks cache clean
```

---

_最后更新：2026-03-01_
_版本：v0.10.0-alpha_
