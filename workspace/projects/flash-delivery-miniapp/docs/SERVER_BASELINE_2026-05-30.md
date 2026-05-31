# 疾蜂同城急送小程序 服务器基线 2026-05-30

## 当前结论

服务器基础环境已经补齐到“可继续部署后端”的状态，但还不是“项目已经上线”。

今天完成的是：

- 验证 SSH 可直连服务器 `8.155.20.204`
- 确认 `nginx`、`mysql`、`node` 已存在且运行中
- 新安装 `Go 1.22.2`
- 为本项目创建独立系统用户和目录
- 为本项目创建独立 MySQL 库和本地账号
- 写入一个未启用的 Nginx 项目示例配置

今天没有做的是：

- 没有启用正式域名
- 没有启用 HTTPS 证书
- 没有部署业务后端
- 没有创建正式 systemd 服务

原因很直接：当前项目仓库还是微信小程序前端模板，不包含现成可上线的 Node/Go 后端。

## 服务器真实状态

### 已验证

- SSH 别名：`t`
- 公网 IP：`8.155.20.204`
- 用户：`root`
- 系统：`Ubuntu 24.04.3 LTS`

### 当前环境版本

- `nginx 1.24.0`
- `mysql 8.0.45`
- `node v18.20.8`
- `npm 10.8.2`
- `go 1.22.2`

### 当前共享机特征

- `nginx` 已有多个站点在跑
- `mysql` 已在运行
- `80/443/3306/8080` 已是共享机敏感端口
- `nginx -t` 通过，但现有站点有重复 `server_name` 警告

所以今天我没有去动现有线上站点，只做了安全增量动作。

## 今天实际创建的项目基线

### 系统用户

- `flash-delivery`

### 目录

- `/opt/flash-delivery-miniapp`
- `/opt/flash-delivery-miniapp/app`
- `/opt/flash-delivery-miniapp/shared`
- `/opt/flash-delivery-miniapp/logs`
- `/etc/flash-delivery-miniapp`

### Nginx 示例配置

- `/etc/nginx/sites-available/flash-delivery-miniapp.example`

说明：

- 这里只是示例配置
- 还没有 `enable`
- 还没有绑定正式域名
- 预留反代端口是 `127.0.0.1:18092`

### MySQL

- 已创建独立数据库：`flash_delivery`
- 已创建独立本地账号：`flash_delivery_app`

敏感密码未写入公开项目文档，只记录在私有台账：

- `/Users/xiaoliao/Desktop/openclaw/workspace/config/private/服务器部署资料台账.local.md`

## 现在能说什么

现在可以说：

- 服务器基础运行环境已经准备好
- 这个项目已经有独立目录、独立数据库、独立本地账号
- 后面可以继续部署 Node 或 Go 后端

现在不能说：

- 项目已经正式上线
- 小程序已经接通正式后端
- 域名和 HTTPS 已经启用

## 下一步最短路径

要把这台服务器继续推进到“项目可访问”，还差 4 件事：

1. 选后端实现：`Node` 还是 `Go`
2. 把后端代码部署到 `/opt/flash-delivery-miniapp/app`
3. 创建 `systemd` 服务并监听 `127.0.0.1:18092`
4. 提供正式域名后启用 Nginx 站点和 HTTPS

## 验证证据

今天已验证通过：

- `ssh t`
- `go version`
- `node -v`
- `npm -v`
- `mysql --version`
- `mysqladmin ping`
- `nginx -t`
