# AGENTS 操作手册

本文档说明项目中可用的 Agent 技能及其使用方法。

---

## git-doc-isolation

**描述**: Git 文档隔离管理技能，使用孤立分支策略确保 docs/ 目录绝不进入 main 分支历史，支持本地开发文档的安全管理和临时使用

### 使用场景

- 需要在本地维护开发文档但不想将其提交到主干分支
- 需要基于 docs 分支中的任务文档执行开发任务
- 管理临时任务文档，确保文档与代码分离

### 核心原则

1. **绝对隔离原则**
   - main 分支历史干净: docs/ 目录从未出现在 main 分支的任何提交中
   - docs 分支完全独立: 使用 orphan 分支创建，与 main 无共同历史
   - temp 分支独立: 用于存放临时文档，与 main 无共同历史
   - 无法静默合并: 由于无共同祖先，误操作 merge 会立即报错

2. **三阶段分离原则**
   - 第一阶段（准备）: 只负责创建 temp 分支，存放临时文档
   - 第二阶段（执行）: 只负责创建 task 分支，展示文档并执行任务
   - 第三阶段（完成）: 只负责清理、更新文档、合并代码

### 使用方法

调用技能时，使用以下提示词触发相应阶段：

| 阶段 | 触发提示词 |
|------|-----------|
| 准备阶段 | "开始准备阶段" / "创建 temp 分支" / "初始化临时文档" |
| 执行阶段 | "开始执行阶段" / "创建 task 分支" / "展示文档并执行任务" |
| 完成阶段 | "开始完成阶段" / "完成任务" / "清理并合并" |

### 分支结构

```
main  →  [A] — [B] — [C] — [D]     (代码历史，无 docs/)
         ↑
         └─  task-* 分支从此分出，执行后合并

docs  →  [X]                      (孤立分支，存放正式文档)

temp  →  [Y]                      (孤立分支，存放临时任务文档)
       ↳ task-20250227/
         ↳ task-20250228/
```

### 完整示例

```bash
# ===== 第一阶段：准备 =====
# 提示词: "开始准备阶段"
git checkout main
git checkout --orphan temp
git rm -rf .
mkdir task-20250227 && echo "# Task" > task-20250227/instruction.md
git add task-20250227/ && git commit -m "chore(temp): add task docs"
git checkout main

# ===== 第二阶段：执行 =====
# 提示词: "开始执行阶段"
git checkout main
git checkout -b task-20250227-feature
git show temp:task-20250227/instruction.md > instruction.md
echo "*.md" > .git/info/exclude
# ... 根据 instruction.md 执行任务 ...

# ===== 第三阶段：完成 =====
# 提示词: "开始完成阶段"
rm instruction.md && rm .git/info/exclude
git add . && git commit -m "feat: 完成任务"
git checkout temp && echo "✅ 已完成" >> task-20250227/README.md
git add . && git commit -m "docs: 更新状态"
git checkout main && git merge task-20250227-feature
git branch -d task-20250227-feature
```

### 注意事项

1. **三阶段必须分开执行**：每个阶段都有明确的触发提示词，不要混合执行
2. **temp 分支是孤立的**：与 main 无共同历史，无法直接合并
3. **task 分支基于 main**：确保代码修改基于最新的 main 分支
4. **临时文档不进入历史**：通过 .git/info/exclude 确保临时文档不被提交
5. **temp 分支文档保留**：任务完成后，temp 分支的文档保留作为记录
