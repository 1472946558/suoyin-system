# Skill: Golf BLE Protocol

## Purpose
高尔夫硬件 BLE 通信协议规范。

## 协议基础

- 标准 BLE GATT
- 设备作为 Peripheral，手机作为 Central
- Service UUID 和 Characteristic UUID 需对齐硬件文档

## 数据包格式

### 发送（手机 → 设备）
| 字段 | 偏移 | 长度 | 说明 |
|------|------|------|------|
| Header | 0 | 1 | 包头标识 |
| Cmd | 1 | 1 | 命令类型 |
| Len | 2 | 1 | 数据长度 |
| Data | 3 | N | 数据内容 |
| Checksum | 3+N | 1 | 校验和 |

### 接收（设备 → 手机）
同上格式，Cmd 对应响应码。

## 关键命令

| Cmd | 方向 | 说明 |
|-----|------|------|
| 0x01 | → | 连接握手 |
| 0x02 | ← | 握手响应 |
| 0x10 | → | 开始测距 |
| 0x11 | ← | 测距数据 |
| 0x20 | → | 请求设备信息 |
| 0x21 | ← | 设备信息响应 |

## 对齐规则

- Android 和 iOS 必须使用完全相同的数据包格式
- 字节序：大端（Big-Endian）
- 字符串编码：UTF-8
- 超时时间：5000ms

## 使用 Agent
golf-ios-agent, golf-qa-agent
