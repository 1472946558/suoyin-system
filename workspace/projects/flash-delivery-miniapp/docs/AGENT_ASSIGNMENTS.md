# Agent 任务分发

## 总控结论

任务类型：新项目启动 + 原生微信小程序前端开发。

今日目标：完成原创同城即时配送小程序 MVP，不做“闪送”1:1 商标/素材复刻。

## 任务表

| Task ID | 任务 | 负责人 | 输入 | 输出 | 完成标准 |
| --- | --- | --- | --- | --- | --- |
| FD-T001 | 项目接入与商用边界 | AI PM | 老板指令 | 项目卡、范围、风险 | 明确可做/不可做 |
| FD-T002 | 产品与业务链设计 | AI Business Analyst | 同城急送业务 | 用户路径、字段、订单状态 | 主链路完整 |
| FD-T003 | 小程序架构与页面拆分 | AI Architect / Tech Lead | PRD、技术栈 | 页面、模块、数据流 | 文件范围清楚 |
| FD-T004 | 原生小程序前端实现 | WeApp / Frontend Agent | 页面清单、UI方向 | WXML/WXSS/JS | 5 个页面可运行 |
| FD-T005 | Mock 数据与本地订单流 | Backend Agent | 订单字段 | mock API、本地存储 | 可生成和查看订单 |
| FD-T006 | QA 验收 | QA Agent | 小程序工程 | 验收表、问题清单 | 主路径通过 |
| FD-T007 | 商用上线清单 | Compliance / DevOps | 商用目标 | 上线风险和后续接口 | 知道上线缺口 |
| FD-T008 | 视觉审稿与图标系统 | UI/UX Agent | 截图、页面结构、图标授权 | 视觉问题清单、修复建议、评分 | CTA 不越界、图标统一、页面高级感达标 |

## 修改范围

允许修改：

- `workspace/projects/flash-delivery-miniapp/**`
- `workspace/dashboard/tasks.json`
- OpenClaw / 飞书任务状态记录

禁止修改：

- 既有 `sccair-mini-frontend` 项目
- 既有 `fullstack-portfolio-website` 项目
- 任何含真实密钥的配置

## 验收标准

1. 微信开发者工具可导入项目。
2. tabBar 正常。
3. 首页可跳转下单。
4. 下单页可生成模拟订单。
5. 订单页可看到新订单。
6. 跟踪页可展示状态进度。
7. 我的页面有常用入口。
8. 不出现“闪送”商标和素材。
9. 所有核心入口必须有统一图标，不允许临时手绘低质图标。
10. 首页双 CTA 必须等宽、等高、同一视觉语法，窄屏下不能越出卡片。
