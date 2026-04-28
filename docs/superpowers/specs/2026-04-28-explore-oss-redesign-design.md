# 发现页改造：从 OSS 运营内容展示

## 背景

当前发现页展示的是用户生成的 `generations` 作品列表。需求改为：从阿里云 OSS 的 `explore/` 目录读取运营精选图片，在发现页以全屏滑动方式展示。

## 目标

- 改造现有发现页，替代原有用户作品列表
- 图片全屏滑动浏览（类似抖音，一次一张）
- 保留点赞和"拍同款"交互
- 测试阶段提供一次性脚本将 OSS 图片同步到数据库

## 数据模型

### `explore_assets` 表

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `SERIAL PRIMARY KEY` | 自增 ID |
| `image_url` | `TEXT NOT NULL` | OSS 公网 HTTPS 直链 |
| `scene_key` | `TEXT NOT NULL DEFAULT ''` | 所属场景标识，用于"拍同款"跳转 |
| `prompt` | `TEXT NOT NULL DEFAULT ''` | 生成提示词，不暴露给用户，仅用于"拍同款"预填充 |
| `created_at` | `TIMESTAMP DEFAULT now()` | 入库时间 |

### `explore_likes` 表

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `SERIAL PRIMARY KEY` | |
| `user_id` | `INTEGER NOT NULL REFERENCES users(id)` | |
| `explore_asset_id` | `INTEGER NOT NULL REFERENCES explore_assets(id)` | |
| `created_at` | `TIMESTAMP DEFAULT now()` | |
| 唯一约束 | `(user_id, explore_asset_id)` | 防重复点赞 |

## 后端 API

### 改造 `ExploreFeedHandler`

- 查询源从 `generations` + `users` + `likes` 的多表 JOIN，改为单表查询 `explore_assets`
- 保留分页参数 `page`、`page_size`
- 返回结构保持与前端 `ExploreItem` 兼容

### 改造 `ExploreLikeHandler`

- 点赞目标字段从 `generation_id` 改为 `explore_asset_id`
- 操作表从 `likes` 改为 `explore_likes`
- 返回结构不变（`success`、`like_count`）

### 配置更新

`config.go` 的 `Config` 结构体补充 OSS 字段，供同步脚本使用：

- `OSSEndpoint`
- `OSSBucket`
- `OSSAccessKeyID`
- `OSSAccessKeySecret`

## 前端交互

### 页面布局

- 使用 `swiper` 组件，`vertical` 纵向滑动
- 每张图片占满全屏，`image` 组件 `mode="aspectFill"`
- 底部保留点赞、拍同款按钮，悬浮覆盖在图片上

### 拍同款

- 点击时从当前 `ExploreItem` 取出 `scene_key` 和 `prompt`
- 跳转到 `/pages/scene/index` 并预填充参数
- `prompt` 不展示在 UI 上

### 预加载

- `swiper` 开启前后各 1 张预加载，减少滑动白屏

## 数据同步脚本

临时脚本 `scripts/sync_explore_assets.go`：

1. 使用阿里云 OSS Go SDK 列出 `explore/` 前缀下所有 `.jpg`/`.jpeg` 对象
2. 构造公网 URL：`https://<bucket>.<endpoint>/explore/<filename>`
3. 批量插入 `explore_assets` 表
4. `scene_key` 和 `prompt` 初始为空字符串，后续由后台管理填充
5. 幂等：已存在的 `image_url` 跳过

## 后续扩展

- 后台管理接口：增删改查 `explore_assets` 记录
- 支持为图片配置排序权重
- 支持批量上传图片到 OSS 并自动入库
