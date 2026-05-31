# 疾蜂同城急送小程序 部署快照

- 项目 ID：flash-delivery-miniapp
- 项目目录：/Users/xiaoliao/Desktop/openclaw/workspace/projects/flash-delivery-miniapp
- 项目类型：微信小程序前端交付包
- 推荐部署方式：微信开发者工具 / 公众平台发布
- 当前 readiness：60/100
- 当前结论：可 Demo，不可直接商用上线

## 1. 版本信息

- 发布分支：待补齐
- 发布版本 / Tag：前端 MVP 交付基线
- Commit SHA：待补齐
- 上一稳定版本：无正式生产版

## 2. 服务器与网络

- 环境：shared prod server baseline / local demo
- 服务器公网 IP：`8.155.20.204`
- 服务器 SSH：`ssh t` -> `root@8.155.20.204:22`
- 域名：未提供
- HTTPS 证书：依赖客户正式 HTTPS API 域名
- 暴露端口：`80` / `443` 已由现有 Nginx 使用；项目预留反代端口 `127.0.0.1:18092`
- 正式 API 域名：未提供
- 测试 API 域名：未提供

## 3. 运行方式

- 主发布路径：微信开发者工具 / 公众平台发布
- 当前导入 AppID：`touristappid`（仅供 Demo）
- 进程守护方式：当前项目后端未部署；systemd 待补
- 反向代理：服务器已有 Nginx；已新增未启用示例 `/etc/nginx/sites-available/flash-delivery-miniapp.example`
- 日志路径：预留 `/opt/flash-delivery-miniapp/logs`
- 监控 / 告警：当前无正式生产监控

## 4. 配置与依赖

- `.env` 来源：当前无，前端配置位于 `utils/config.js`
- 当前模式：`mock`
- 当前 API Base URL：`https://api.your-domain.com`（占位值）
- 数据库：已在共享机创建 `flash_delivery` 独立库；正式表结构未导入
- Redis / 缓存：共享机已有 Redis，但本项目未接入
- 服务器语言环境：`Node v18.20.8`、`npm 10.8.2`、`Go 1.22.2`
- 服务器项目目录：`/opt/flash-delivery-miniapp`
- 第三方密钥：微信小程序 AppID 未提供；腾讯地图 Key 未提供；微信支付商户信息未提供
- 小程序发布：待客户提供正式 AppID 后进入正式发布流程
- 协议地址：隐私政策和用户协议 URL 未提供
- 客服入口：未提供

## 5. 回滚方案

- 镜像 / 构建物回滚点：当前以微信开发者工具导入包和本地代码基线为准
- 数据库回滚策略：当前无真实数据库
- 回滚执行人：待补齐
- 回滚完成验证：重新导入上一个可演示版本并执行 `npm run validate:delivery`

## 6. 实际发布记录

- 发布时间：尚未进入正式发布
- 操作人：待补齐
- smoke test 结论：`npm run validate` 通过；真机和正式环境未验证
- 剩余风险：缺正式 AppID、正式域名、HTTPS API、真实后端、地图、支付、客服

## 7. 已识别资料来源

- `README.md`
- `docs/CUSTOMER_INTEGRATION_GUIDE.md`
- `08_go_live/CUSTOMER_LAUNCH_INPUTS.md`

## 8. 当前识别备注

- 当前版本仅适合演示和客户确认前端范围。
- 没有客户正式 AppID 和合法域名前，不进入正式上传、提审和上线动作。
