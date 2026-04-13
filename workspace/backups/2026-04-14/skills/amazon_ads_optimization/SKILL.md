# Skill: Amazon Ads Optimization

## Purpose
Amazon.de 广告优化决策框架。

## Campaign 结构

```
Campaign: [产品名] - [自动/手动]
  Ad Group 1: 精准词（高转化词）
  Ad Group 2: 广泛词（拓流量）
  Ad Group 3: ASIN 定向（抢竞品流量）
  Ad Group 4: 类目定向（品类覆盖）
```

## 优化节奏

### 每日
- 检查 ACOS 是否在阈值内
- 否定无效搜索词
- 调整日预算

### 每周
- 分析搜索词报告
- 新增高转化词到精准组
- 否定低转化词
- 调整出价策略

### 每月
- 整体 ACOS 回顾
- Campaign 结构优化
- 竞品广告策略分析

## 出价调整规则

| 场景 | 动作 | 幅度 |
|------|------|------|
| ACOS < 目标且转化升 | 加价 | +10-20% |
| ACOS < 目标且转化稳 | 维持 | 0% |
| ACOS > 目标但转化升 | 微降 | -5-10% |
| ACOS > 目标且转化降 | 降价 | -15-25% |
| ACOS > 50% 持续 7 天 | 暂停 | - |

## 关键指标

- ACOS（广告成本销售比）
- TACOS（总广告成本销售比）
- 转化率（CVR）
- 点击率（CTR）
- CPC（单次点击成本）

## 使用 Agent
amazon-de-agent
