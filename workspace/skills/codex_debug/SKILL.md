# Skill: Codex Debug

## Purpose
使用 Codex / AI 辅助调试的标准化流程。

## 调试流程

1. **复现** — 确认 Bug 可稳定复现
2. **定位** — 找到出错的具体文件和行
3. **分析** — 理解 root cause
4. **修复** — 最小改动
5. **验证** — 确认修复且无副作用

## Prompt 模板

```markdown
## Bug Description
[现象描述]

## Reproduction Steps
1. [步骤1]
2. [步骤2]

## Error Output
```
[错误日志/堆栈]
```

## Suspected File
[文件路径]:[行号]

## Expected Behavior
[期望行为]

## Constraints
- 只修改 [指定文件]
- 不重构周边代码
- 保持向后兼容
```

## 调试原则

- 不猜，用日志和数据说话
- 每次只改一个变量
- 修复后必须跑验证
- 记录 root cause 到 memory

## 使用 Agent
golf-ios-agent, delivery-agent
