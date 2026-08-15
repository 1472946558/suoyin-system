# 13 - 统一素材上传规格（管理后台）

> 版本：v1.0（2026-08-15）
> 状态：接口草案，未实现。当前后端**没有**通用图片上传接口（仅有回收照片 OSS 两段式直传 `/api/v1/uploads/recycle-photos/prepare|complete`，可复用其 OSS 配置）。

## 1. 目标

为后台所有"顾客端内容管理"模块提供统一的图片上传/删除/预览能力，图片字段一律存**可访问 URL**（不存裸本地路径）。

## 2. 接口设计

### 2.1 上传图片

```
POST /api/admin/uploads/images
Content-Type: multipart/form-data
Auth: 管理员 token（session: 前缀）
Permission: customer_content.manage 或对应业务 ability
```

**Form 字段：**

| 字段 | 必填 | 说明 |
|------|------|------|
| file | 是 | 二进制文件 |
| scene | 是 | 上传场景枚举：`banner` / `style_main` / `style_detail` / `store` / `service_intro` |
| refId | 否 | 关联对象 id（如款式 id、门店 id），仅用于审计 |

**Response 200：**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "img-20260815-0001",
    "url": "https://jinjiangguan.com/assets/2026/08/img-20260815-0001.webp",
    "path": "/assets/2026/08/img-20260815-0001.webp",
    "scene": "banner",
    "size": 284512,
    "mimeType": "image/webp",
    "createdAt": "2026-08-15T12:00:00Z"
  }
}
```

> 路径前缀说明：见 §4 存储策略。`url` 为可直接访问的绝对 URL（顾客端免拼接）；`path` 保留相对路径便于迁移。

**Error Codes：**

| code | HTTP | 说明 | 前端提示文案 |
|------|------|------|--------------|
| 40010 | 400 | 未选择文件 | 请选择要上传的图片 |
| 40011 | 400 | 格式不支持（魔数校验失败） | 仅支持 jpg / jpeg / png / webp 格式 |
| 40012 | 400 | 超过大小限制 | 图片不能超过 5MB，请压缩后重试 |
| 40301 | 403 | 无上传权限 | 当前角色无上传权限 |
| 50002 | 500 | 存储写入失败 | 上传失败，请稍后重试 |

### 2.2 删除图片

```
DELETE /api/admin/uploads/images/:id
Auth: 管理员 token
Permission: customer_content.manage
```

逻辑删除（软删除）：标记 deleted，文件保留 30 天后由清理任务物理删除。若图片仍被业务引用（banner/款式/门店在用），返回 40013，前端提示"该图片正在被使用，请先替换"。

### 2.3 素材列表（可选，管理素材库）

```
GET /api/admin/uploads/images?scene=banner&page=1&pageSize=20
```

供后台"从素材库选择"入口使用。

## 3. 限制规则

| 规则 | 值 | 说明 |
|------|-----|------|
| 支持格式 | jpg / jpeg / png / webp | 后端按**文件头魔数**校验，不信任扩展名 |
| 单文件大小 | ≤ 5MB | banner 场景建议 ≤ 2MB、750×300px |
| 单次上传 | 1 张/请求 | 多图由前端循环调用 |
| 文件名 | 服务端重命名 | `{scene}-{yyyyMMdd}-{seq}.{ext}`，禁止保留用户原始文件名 |
| 并发引用 | 同一图片可被多处引用 | 删除时校验引用计数 |

前端上传失败统一 toast 提示上表的"前端提示文案"，并保留已填表单内容不丢失。

## 4. 存储策略（复用现有能力）

优先级：

1. **OSS（推荐）**：复用现有 `settings.Storage`（回收照片直传已用的 OSS 配置）。管理后台直传模式改为"后端中转"：multipart 收文件 → 后端以 ServiceAccount 写入 OSS → 返回 publicUrl。objectKey 规则：`customer-mgmt/{scene}/{yyyyMM}/{filename}`。
2. **本地磁盘（兜底）**：未配置 OSS 时存服务器本地 `/srv/miniapp/backend/assets/customer-mgmt/{scene}/{yyyyMM}/`，由后端静态路由（或 Nginx `location /assets/`）对外提供访问，返回 `https://jinjiangguan.com/assets/customer-mgmt/...`。

无论哪种存储，**入库字段统一存绝对 URL**；若存相对路径（`/assets/...`），顾客端沿用 customer-services.js 的 `absUrl()` 补全域名（现有约定）。

## 5. 上传场景 ↔ 顾客端页面对应表

| scene | 上传入口（后台） | 展示位置（顾客端） | 建议尺寸 |
|-------|-----------------|-------------------|----------|
| banner | 顾客端配置→首页 Banner | 首页轮播图 | 750×300，≤2MB |
| style_main | 款式/工费管理→款式编辑（主图） | 款式列表卡片、详情首图 | 800×800 正方形 |
| style_detail | 款式/工费管理→款式编辑（详情多图） | 款式详情轮播 | 800×800 |
| store | 门店管理→门店编辑 | 门店列表卡片、门店详情头图 | 750×500 |
| service_intro | 顾客端配置→服务说明配图 | 我的页服务说明 | 750×任意 |

## 6. 数据模型

新增 `UploadAsset`（存 app_configs，key：`upload_assets`；量小无需独立表）：

```go
type UploadAsset struct {
    ID        string    `json:"id"`        // img-yyyyMMdd-seq
    OrgID     string    `json:"orgId"`
    Scene     string    `json:"scene"`     // banner / style_main / style_detail / store / service_intro
    URL       string    `json:"url"`       // 可访问 URL（绝对）
    Path      string    `json:"path"`      // 相对路径（迁移备用）
    Storage   string    `json:"storage"`   // oss / local
    ObjectKey string    `json:"objectKey"` // oss key 或本地相对路径
    MimeType  string    `json:"mimeType"`
    Size      int64     `json:"size"`
    RefID     string    `json:"refId"`     // 引用对象 id（审计用）
    Deleted   bool      `json:"deleted"`
    CreatedBy string    `json:"createdBy"`
    CreatedAt time.Time `json:"createdAt"`
}
```

## 7. 安全要求

1. 仅管理后台 token + ability 校验通过可上传（顾客 token 一律 403）。
2. 魔数校验防止伪装成图片的可执行文件；不做图片解析渲染，仅存储。
3. 响应不回显服务器内部路径，仅返回 url/path。
4. 上传写操作日志（谁、何时、哪个场景、哪个文件）。
5. URL 访问走 HTTPS；本地存储目录禁止列目录（Nginx `autoindex off`）。

## 8. 前端（管理后台 Vue）统一上传组件要求

- 统一封装 `ImageUploader` 组件：点击选择 / 预览 / 删除 / 重新上传；
- 上传前本地校验格式与大小（失败即提示，不发请求）；
- 上传中显示 loading，成功后回填 URL 到表单字段；
- 支持从素材库（2.3）选择历史图片复用；
- 所有图片表单字段值为 URL 字符串（或 URL 数组），不存 File 对象。
