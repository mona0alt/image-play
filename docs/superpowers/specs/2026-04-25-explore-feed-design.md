# 发现页面（Explore Feed）设计文档

## 背景

将底部导航的"历史"页面移入"我的"页面，原历史导航栏位置替换为"发现"页面。发现页面采用全屏沉浸式 Feed 设计，参考 `docs/page_design/explore.html` 的抖音式上下滑动体验。

## 信息架构变更

### 底部导航调整

**变更前：**
- `gallery` → 艺廊 (`/pages/home/index`)
- `create` → 创作 (`/pages/scene/index`)
- `history` → 历史 (`/pages/history/index`)
- `profile` → 我的 (`/pages/profile/index`)

**变更后：**
- `gallery` → 艺廊 (`/pages/home/index`)
- `create` → 创作 (`/pages/scene/index`)
- `explore` → 发现 (`/pages/explore/index`)
- `profile` → 我的 (`/pages/profile/index`)

### "我的"页面调整

- 在"余额"区块之后、"最近作品"区块之前新增 **"历史档案"入口卡片**
- 点击后 `uni.navigateTo({ url: '/pages/history/index' })` 进入历史子页面
- 历史子页面保留现有筛选器和列表布局，但不再显示底部导航

## 发现页面设计

### 布局规范

- **容器**：全屏高度（100vh），隐藏滚动条。使用 `swiper` 组件纵向模式实现全屏切换（比 `scroll-snap` 在 uni-app 中兼容性更好）
- **作品项**：每项占满一屏（100vh × 100vw），图片 `object-fit: cover` 铺满
- **导航栏**：`pages.json` 中配置 `"navigationStyle": "custom"` 隐藏系统导航栏，由 ExploreCard 自行处理安全区
- **右侧按钮**：固定在右下角偏上（bottom: 140px, right: 16px），垂直排列，间距 20px
- **信息卡片**：固定在左下角（bottom: 100px, left: 16px, right: 80px），max-width 约 280px
- **安全区**：顶部预留状态栏高度（约 44px），底部预留 TabBar + safe-area-inset-bottom

### 交互行为

- **上下滑动**：切换作品，`scroll-snap` 吸附到下一屏
- **点赞**：点击心形按钮 → 调用 `POST /explore/like` → 切换填充状态并显示动画
- **同款**：点击 ✦ 按钮 → 携带该作品的 `scene_key` 和 `prompt` 参数 → `uni.reLaunch` 到创作页（`/pages/scene/index`）
- **预加载**：当前页前后各预加载 1 张图片

### 样式细节（适配 uni-app）

- 玻璃效果：使用 `backdrop-filter: blur(20px)`，Android 不支持时降级为半透明背景 `rgba(0,0,0,0.4)`
- 按钮阴影：使用 `box-shadow: 0 4px 20px rgba(0,0,0,0.3)`
- 文字阴影：所有覆盖在图片上的文字使用 `text-shadow: 0 1px 3px rgba(0,0,0,0.5)`
- 图标：使用 Material Symbols Outlined（与 explore.html 一致）

## 数据流与 API 设计

### 新增端点

| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/explore/feed` | 获取精选推荐作品列表（分页） |
| POST | `/explore/like` | 点赞/取消点赞作品（参数：`generation_id`, `action: 'like' | 'unlike'`） |
| GET | `/explore/like-status` | 批量获取点赞状态（可选优化） |

### Feed 数据模型

```json
{
  "items": [
    {
      "id": 123,
      "user": {
        "id": "user_456",
        "nickname": "林溪_AI",
        "avatar_url": "https://..."
      },
      "image_url": "https://...",
      "thumbnail_url": "https://...",
      "prompt": "超高清、极简主义、渐变...",
      "scene_key": "portrait",
      "like_count": 128,
      "is_liked": false,
      "description": "流动的梦境系列...",
      "created_at": "2026-04-20T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 100,
    "has_more": true
  }
}
```

### 同款跳转参数

```
/pages/scene/index?scene_key={scene_key}&prompt={encodeURIComponent(prompt)}&reference_id={id}
```

## 组件拆分

### 新增文件

```
frontend/src/
├── pages/
│   └── explore/
│       ├── index.vue          # 发现页面主组件
│       ├── view-model.ts      # 数据转换与状态管理
│       └── view-model.test.ts # 单元测试
├── components/
│   ├── explore/
│   │   ├── ExploreFeed.vue    # 全屏 Feed 容器
│   │   ├── ExploreCard.vue    # 单个作品项
│   │   └── LikeButton.vue     # 点赞按钮
│   └── profile/
│       └── HistoryEntryCard.vue  # 历史入口卡片
└── utils/
    └── image-preloader.ts     # 图片预加载工具
```

### 修改文件

```
frontend/src/
├── utils/navigation.ts        # PRIMARY_TABS：history → explore
├── pages.json                 # 注册 explore 页面
├── pages/profile/             # 新增历史入口
├── components/navigation/
│   └── GalleryTabBar.vue      # 调整 activeKey
└── services/api.ts            # 新增 explore API
```

## 错误处理

### API 异常

- **Feed 首次加载失败**：显示空状态卡片，提供重试按钮
- **分页加载失败**：静默失败，用户上滑时再次触发
- **点赞失败**：回滚本地状态，Toast 提示
- **网络恢复**：自动恢复 pending 请求

### 图片加载异常

- **主图加载失败**：显示占位渐变背景
- **头像加载失败**：显示默认头像
- **预加载失败**：静默处理，滑动到该项时重新尝试

### 兼容性降级

- **scroll-snap 不支持**：降级为普通 scroll-view
- **backdrop-filter 不支持**：降级为纯半透明背景
- **基础库版本过低**：启动时检测，提示更新微信版本

### 同款跳转异常

- **scene_key 已下架**：Scene 页面显示提示并清空参数
- **prompt 过长**：截断至 500 字符
- **未登录用户**：复用现有登录检查逻辑

## 性能策略

- **图片预加载**：当前展示项加载完成后，预加载前后各 1 项
- **分页触发**：滑动到第 8 条时自动请求下一页
- **内存管理**：列表超过 30 条时，移除视口外较远的项（保留前后各 5 项）
- **快速滑动防抖**：scroll 事件节流 200ms
- **图片尺寸**：先加载 `thumbnail_url` 占位，再加载 `image_url` 大图
