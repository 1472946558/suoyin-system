# Skill: Cursor Prompt Builder

## Purpose
生成高质量的 Cursor AI 编程 Prompt。

## Prompt 结构

```markdown
## Context
[项目背景、当前状态、技术栈]

## Task
[具体要做什么]

## Constraints
- [约束1]
- [约束2]

## Files to Modify
- [文件1]: [改动说明]
- [文件2]: [改动说明]

## Acceptance Criteria
- [ ] [验收标准1]
- [ ] [验收标准2]

## DO NOT
- [不要做的事1]
- [不要做的事2]
```

## 编写规则

1. **具体** — 不说"优化性能"，说"将列表渲染时间从 2s 降到 500ms"
2. **约束明确** — 列出所有不允许做的事
3. **文件精确** — 给出要改的文件路径
4. **验收可执行** — 每个 AC 都可以 yes/no 判断
5. **上下文完整** — 包含相关的代码片段或接口定义

## Prompt 质量检查

| 检查项 | 标准 |
|--------|------|
| 任务描述 | 可用一句话概括 |
| 约束条件 | 至少 3 条 |
| 文件范围 | 明确列出 |
| 验收标准 | 可执行 |
| 不做的事 | 明确列出 |

## 使用 Agent
prompt-engineer-agent
