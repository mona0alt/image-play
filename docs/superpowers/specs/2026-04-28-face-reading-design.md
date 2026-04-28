# 面相分析功能设计文档

## 背景
将 DMXAPI 图片识别（面相分析）能力集成到 image-play 小程序中，用户在首页点击入口后，可上传人物照片，由后端代理调用 DMXAPI 进行面相解析并返回结果。

## 目标
- 在小程序内提供一键面相分析体验
- API Key 不暴露在前端
- 分析结果仅临时展示，支持复制
- 风格与现有小程序页面保持一致

---

## 架构

```
+-------------+     +----------------+     +-----------+
|  微信小程序  | --> |  Go Backend    | --> |  DMXAPI   |
|  (uni-app)  |     |  /api/face-reading |  |  Moonshot |
+-------------+     +----------------+     +-----------+
```

### 新增后端接口
- `POST /api/face-reading`（JWT 保护）
  - 请求：`{ "image_base64": "data:image/jpeg;base64,..." }`
  - 响应：`{ "result": "分析文本..." }`
  - 后端用标准 `net/http` 向 DMXAPI 发起 Chat Completion 请求，Prompt 固定为面相分析模板

### 新增前端页面
- `pages/face-reading/index`：独立页面
  - 图片选择区（拍照/相册），选中后预览
  - 提交按钮 → Loading → 结果展示
  - 结果区提供「复制结果」按钮

### 入口
- `pages/home/index` 艺廊页面中，在现有内容区域新增「面相分析」功能卡片
- 点击后 `uni.navigateTo` 跳转到新页面

---

## 数据流

1. 用户点击首页「面相分析」卡片 → 进入新页面
2. 选择图片：`uni.chooseImage({ sizeType: ['compressed'] })`
3. 前端用 `uni.getFileSystemManager().readFile` 读取为 base64
4. 调用后端 `POST /api/face-reading`
5. 后端从配置读取 Key，调 DMXAPI（`https://api.moonshot.cn/v1`）
6. 拿到结果 → 前端渲染，支持一键复制

---

## 配置

`backend/config.go` / `config.yaml` 新增字段：

| 字段 | 说明 | 默认值 |
|---|---|---|
| `dmx_api_key` | DMXAPI 令牌 | 空字符串 |
| `dmx_api_base_url` | DMXAPI 基础地址 | `https://api.moonshot.cn/v1` |
| `dmx_model` | 模型名称 | `kimi-k2.6` |

支持环境变量覆盖：`DMX_API_KEY`, `DMX_API_BASE_URL`, `DMX_MODEL`。

---

## 错误处理

| 场景 | 前端行为 | 后端行为 |
|---|---|---|
| 未选择图片就点击分析 | Toast：「请先选择一张照片」 | — |
| 图片过大（base64 > 5MB） | Toast：「图片过大，请选择较小的图片」 | 返回 `413 Payload Too Large` |
| DMXAPI 超时/失败 | Toast：「分析服务暂时繁忙，请稍后重试」 | 返回 `502 Bad Gateway`，记录日志 |
| 网络异常 | Toast：「网络异常，请检查连接」 | — |

---

## UI 风格

沿用项目现有设计体系：
- 背景色：`--gallery-bg`（`#fdf8f8`）
- 卡片/表面色：`--gallery-surface`（白色）
- 圆角：28rpx 大卡片、999rpx 按钮
- 字体：主标题 40rpx 加粗，正文 24rpx， eyebrow 20rpx uppercase muted
- 布局：垂直 flex column，gap 24rpx
- 使用 `GalleryPageShell` 包裹页面，保持导航和底部 Tab 一致

### 中文文案

- 页面标题：「面相分析」
- 说明文案：「上传一张正面清晰照片，AI 将基于传统面相学进行解析」
- 按钮：「选择照片」、「重新选择」、「开始分析」、「复制结果」
- 加载文案：「正在分析面相，请稍候…」
- 错误提示：「分析服务繁忙，请稍后重试」

---

## 安全

- 接口走现有 JWT 中间件，必须登录后才能调用
- API Key 仅存于后端配置/环境变量，不暴露给前端
- 后续 Higress 接入后可在此接口做统一限流和审计
