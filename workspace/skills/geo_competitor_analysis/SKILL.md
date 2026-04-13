# Skill: GEO Competitor Analysis

## Purpose
GEO/AI 搜索领域竞品分析框架。

## 分析维度

### 产品层
- 核心功能是什么
- 目标用户是谁
- 定价策略
- 技术栈
- 差异化点

### 市场层
- 市场规模
- 增长趋势
- 用户获取渠道
- 竞品数量和强度

### 技术层
- AI 模型使用（自研/第三方）
- 数据来源
- API 开放程度
- 集成能力

## 竞品追踪表

```yaml
competitor:
  name: [公司名]
  product: [产品名]
  url: [官网]
  founded: YYYY
  funding: [融资阶段]
  features: [核心功能列表]
  pricing: [定价]
  strength: [优势]
  weakness: [劣势]
  threat_level: HIGH / MEDIUM / LOW
  last_checked: YYYY-MM-DD
```

## 输出格式

```markdown
# 竞品分析报告

## 概览
[1段话总结]

## 竞品对比表
| 维度 | 我方 | 竞品A | 竞品B |
|------|------|-------|-------|

## 机会点
1. [未被满足的需求]
2. [技术空白]
3. [定价空间]

## 威胁评估
1. [直接威胁]
2. [间接威胁]

## 建议
- 短期：[立即行动]
- 中期：[3个月规划]
```

## 使用 Agent
geo-product-agent, research-agent
