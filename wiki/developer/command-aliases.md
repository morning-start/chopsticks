# 命令别名规范

> Chopsticks CLI 命令别名定义与使用指南

---

## 标准别名列表

| 主命令      | 别名            | 说明           | 示例                     |
| ----------- | --------------- | -------------- | ------------------------ |
| `install`   | `i`             | 安装软件       | `chopsticks i git`       |
| `uninstall` | `rm`, `remove`  | 卸载软件       | `chopsticks rm git`      |
| `update`    | `up`, `upgrade` | 更新软件       | `chopsticks up git`      |
| `search`    | `s`, `find`     | 搜索软件       | `chopsticks s git`       |
| `list`      | `ls`            | 列出已安装软件 | `chopsticks ls`          |
| `bucket`    | -               | 软件源管理     | `chopsticks bucket list` |

---

## 使用示例

### 安装软件

```bash
# 使用完整命令
chopsticks install git

# 使用别名
chopsticks i git

# 安装多个软件
chopsticks i git nodejs python
```

### 卸载软件

```bash
# 使用完整命令
chopsticks uninstall git

# 使用别名
chopsticks rm git
# 或
chopsticks remove git
```

### 更新软件

```bash
# 使用完整命令
chopsticks update git

# 使用别名
chopsticks up git
# 或
chopsticks upgrade git

# 更新所有软件
chopsticks up --all
```

### 搜索软件

```bash
# 使用完整命令
chopsticks search git

# 使用别名
chopsticks s git
# 或
chopsticks find git
```

### 列出软件

```bash
# 使用完整命令
chopsticks list

# 使用别名
chopsticks ls

# 列出所有可用软件
chopsticks ls --all
```

---

## 4. 版本历史

| 版本          | 变更                                         |
| ------------- | -------------------------------------------- |
| v0.10.0-alpha | 标准化别名，更新版本号                       |
| v0.6.0-alpha  | 标准化别名，移除 `serve`, `clear`, `refresh` |
| v0.5.0-alpha  | 初始别名支持                                 |

---

_文档版本: v1.0_
_最后更新: 2026-03-01_
