# 缓存管理指南

> Chopsticks 缓存机制详解与管理方法

---

## 1. 缓存概述

Chopsticks 使用多级缓存机制来提升性能和减少网络请求：

- **下载缓存**: 存储下载的软件包
- **元数据缓存**: 存储软件源索引和软件信息
- **临时文件**: 安装过程中的临时数据

---

## 2. 缓存目录结构

```
%USERPROFILE%\.chopsticks\
├── cache/
│   ├── downloads/          # 下载的软件包
│   │   ├── git-2.43.0.zip
│   │   └── node-20.0.0.msi
│   ├── metadata/           # 元数据缓存
│   │   ├── main-index.json
│   │   └── extras-index.json
│   └── temp/               # 临时文件
│       └── install-*.tmp
```

---

## 3. 缓存类型

### 3.1 下载缓存

**位置**: `cache/downloads/`

**用途**: 
- 存储下载的软件安装包
- 避免重复下载相同版本
- 支持离线安装

**管理**:
```bash
# 查看下载缓存大小
chopsticks cache size downloads

# 清理特定软件的缓存
chopsticks cache clean --package git

# 清理所有下载缓存
chopsticks cache clean --downloads
```

### 3.2 元数据缓存

**位置**: `cache/metadata/`

**用途**:
- 存储软件源索引
- 缓存软件元数据
- 加速搜索操作

**管理**:
```bash
# 更新元数据缓存
chopsticks bucket update

# 清理元数据缓存
chopsticks cache clean --metadata
```

### 3.3 临时文件

**位置**: `cache/temp/`

**用途**:
- 安装过程中的解压文件
- 下载中的临时文件
- 自动清理

**管理**:
```bash
# 清理临时文件
chopsticks cache clean --temp
```

---

## 4. 缓存管理命令

### 4.1 查看缓存

```bash
# 查看缓存总大小
chopsticks cache size

# 查看详细缓存信息
chopsticks cache size --verbose

# 查看特定类型缓存
chopsticks cache size --downloads
chopsticks cache size --metadata
```

### 4.2 清理缓存

```bash
# 清理所有缓存
chopsticks cache clean

# 清理下载缓存（保留最近7天）
chopsticks cache clean --downloads --keep-days 7

# 清理特定软件的下载缓存
chopsticks cache clean --package git --package nodejs

# 清理元数据缓存
chopsticks cache clean --metadata

# 清理临时文件
chopsticks cache clean --temp

# 强制清理（不提示确认）
chopsticks cache clean --force
```

### 4.3 配置缓存

```bash
# 设置下载缓存大小限制（MB）
chopsticks config set cache.download_limit 1024

# 设置元数据缓存过期时间（小时）
chopsticks config set cache.metadata_ttl 24

# 启用自动清理
chopsticks config set cache.auto_clean true

# 设置自动清理阈值（MB）
chopsticks config set cache.auto_clean_threshold 512
```

---

## 5. 自动清理策略

Chopsticks 支持自动缓存清理：

### 5.1 基于大小的清理

当缓存总大小超过阈值时，自动清理最旧的文件：

```yaml
# config.yaml
cache:
  auto_clean: true
  auto_clean_threshold: 512  # MB
  priority: "downloads"  # 优先清理下载缓存
```

### 5.2 基于时间的清理

定期清理过期缓存：

```yaml
# config.yaml
cache:
  download_max_age: 30  # 天
  metadata_max_age: 7   # 天
  temp_max_age: 1       # 天
```

### 5.3 安装后清理

安装完成后自动清理临时文件：

```bash
# 启用安装后自动清理
chopsticks config set cache.clean_after_install true
```

---

## 6. 故障排除

### 问题: 缓存占用空间过大

**解决方案**:
```bash
# 查看缓存详情
chopsticks cache size --verbose

# 清理所有缓存
chopsticks cache clean

# 调整缓存限制
chopsticks config set cache.download_limit 512
```

### 问题: 缓存损坏导致安装失败

**解决方案**:
```bash
# 清理特定软件的缓存
chopsticks cache clean --package <软件名>

# 重新安装
chopsticks install <软件名>
```

### 问题: 元数据过期

**解决方案**:
```bash
# 更新所有软件源
chopsticks bucket update

# 或清理元数据缓存
chopsticks cache clean --metadata
```

---

## 7. 最佳实践

1. **定期清理**: 建议每周运行一次 `chopsticks cache clean`
2. **设置限制**: 根据磁盘空间设置合理的缓存限制
3. **保留常用**: 对于经常安装的软件，保留其缓存可加速安装
4. **监控大小**: 定期使用 `chopsticks cache size` 监控缓存占用

---

_文档版本: v1.0_  
_最后更新: 2026-02-28_
