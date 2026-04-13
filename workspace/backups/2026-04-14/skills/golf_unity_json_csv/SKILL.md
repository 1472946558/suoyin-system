# Skill: Golf Unity JSON/CSV Protocol

## Purpose
Unity 与 Native iOS 之间的 JSON/CSV 数据交换协议。

## JSON 协议

### 请求格式（Native → Unity）
```json
{
  "type": "command_type",
  "timestamp": 1713000000,
  "data": { }
}
```

### 响应格式（Unity → Native）
```json
{
  "type": "response_type",
  "status": "ok | error",
  "data": { },
  "error": null
}
```

## CSV 协议

### 格式规范
- 分隔符：逗号
- 换行符：`\n`
- 编码：UTF-8
- 第一行为表头
- 空值用空字符串表示

### 标准字段
```
id,timestamp,type,value,unit,status
```

## 对齐规则

- 字段名大小写必须与 Android 一致
- 数值精度：浮点数保留 2 位
- 时间戳：Unix timestamp（秒级）
- 空数据：JSON 用 null，CSV 用空字符串

## 常见问题

| 问题 | 原因 | 解决 |
|------|------|------|
| Unity 收不到数据 | Bridge 未初始化 | 检查 UnityBridge 设置 |
| CSV 解析失败 | 换行符不一致 | 统一用 \n |
| JSON 字段缺失 | 版本不同步 | 对齐 Android JSON schema |

## 使用 Agent
golf-ios-agent, golf-qa-agent
