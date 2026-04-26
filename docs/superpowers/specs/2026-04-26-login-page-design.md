# 小程序登录页设计文档

## 1. 概述

将现有的静默自动登录改为**强制登录页流程**。启动后若未登录，用户必须完成微信授权登录；登录成功后引导用户设置昵称（可跳过）。

设计参考 `docs/page_design/login.html` 的黑白灰高端艺廊风格。

## 2. 新增页面

### 2.1 `pages/login/index.vue` — 登录页

**布局（从上到下）：**

1. **品牌区**
   - 几何 logo：两个旋转细线方框 + 中心图标，纯 CSS 绘制
   - 主标题："精品场景馆"，`font-h3`（48rpx），字重 600
   - 副标题："开启您的艺术创作之旅"，小字大写跟踪字距

2. **操作区**
   - **微信一键登录按钮**：黑色填充（`#000000`），白色文字，圆角 16rpx，高度 112rpx，左侧带微信图标
   - **装饰图组**：按钮下方 3 张纵向比例图片（灰度画廊预览风格），增加艺廊氛围

3. **底部区**
   - 协议勾选行：自定义 checkbox（方形细边框），文字"我已阅读并同意《用户协议》和《隐私政策》"
   - 协议链接为占位，点击提示"协议内容即将上线"

**交互：**
- 未勾选协议时点击登录：触发 `uni.vibrateShort()` + 提示"请先同意用户协议"
- 勾选后点击：调用 `uni.login` 获取 code → POST `/api/auth/login` → 保存 token → 跳转昵称设置页
- 按钮 loading 状态：禁用重复点击

### 2.2 `pages/nickname-setup/index.vue` — 昵称设置页

**布局：**
- 顶部返回箭头（可回登录页）
- 主标题："如何称呼您？"，`font-h3`（48rpx），字重 500
- 副标题："您可以随时在个人中心修改"
- 输入框：底部细线样式（1rpx solid `#c4c7c7`），placeholder "请输入昵称"，focus 时线条变黑
- 字数提示：右下角 `0/20`
- **确认按钮**：黑色填充，白色文字，圆角 16rpx，高度 96rpx
- **跳过**：透明背景，黑色文字，更小字号

**交互：**
- 点击"确认"：校验长度 2-20 → `PUT /api/me` → 成功则 `uni.reLaunch` 到首页
- 点击"跳过"：直接 `uni.reLaunch` 到首页，保留后端默认随机昵称
- 长度不符时提示："昵称长度需在 2-20 个字符之间"

## 3. 登录流程

```
App.vue onLaunch
  │
  ▼
已有 access_token？─────是─────▶ 正常进入首页
  │否
  ▼
uni.reLaunch 至 pages/login/index
  │
  ▼
用户勾选协议 + 点击「微信一键登录」
  │
  ▼
uni.login 获取 code ──▶ POST /api/auth/login
  │
  ▼
保存 access_token ────▶ uni.reLaunch 至 pages/nickname-setup/index
  │
  ▼
用户输入昵称（或跳过）──▶ PUT /api/me
  │
  ▼
uni.reLaunch 至首页
```

## 4. 后端新增接口

### `PUT /api/me`

更新当前登录用户的信息。

- **请求体：** `{ "nickname": "string" }`
- **校验：** nickname 必填，长度 2-20
- **响应：** `{ id, nickname, balance, free_quota }`
- **权限：** 需要 Bearer Token

## 5. 现有文件修改

### 5.1 `App.vue`

`onLaunch` 不再自动调用 `ensureSession()`：
- 检查 `uni.getStorageSync('access_token')`
- 无 token：`uni.reLaunch({ url: '/pages/login/index' })`
- 有 token：正常进入首页

### 5.2 `pages.json`

在 `pages` 数组中注册两个新页面（`pages[0]` 保持为首页，避免已登录用户启动时闪屏）：
- `pages/login/index`
- `pages/nickname-setup/index`

两个页面配置：
- `navigationBarTitleText`：登录页留空或"SCENE GALLERY"，昵称页"设置昵称"
- `navigationBarBackgroundColor`：`#fdf8f8`
- `backgroundColor`：`#fdf8f8`

### 5.3 `session.ts`

保持 `ensureSession()` 逻辑不变，供其他页面/API 调用时使用，但不再由 `App.vue` 自动触发。

### 5.4 `api.ts` / `request()`

- 接口返回 401 时：清除本地 token + `uni.reLaunch({ url: '/pages/login/index' })`
- 所有页面的 401 处理统一收敛到这里，无需各页面自行处理

### 5.5 导航拦截（App.vue）

在 `onLaunch` 中通过 `uni.addInterceptor` 拦截 `navigateTo`、`redirectTo`、`switchTab`、`reLaunch`：
- 若目标页不是登录页且本地无 token，则拦截并跳转到登录页

## 6. 错误处理

| 场景 | 处理方式 |
|---|---|
| `uni.login` 失败 | `uni.showToast({ title: '微信登录失败，请重试', icon: 'none' })` |
| 后端登录接口失败 | 同上 |
| 网络异常 | `uni.showToast({ title: '网络异常，请检查网络', icon: 'none' })` |
| 未勾选协议点击登录 | `uni.vibrateShort()` + 提示"请先同意用户协议" |
| 昵称更新接口失败 | 提示失败但允许进入首页（非阻塞）|
| Token 过期（401） | 清除 token，强制回到登录页 |

## 7. 样式规范

- 页面背景：`#fdf8f8`
- 页面安全边距：左右 64rpx
- 文字主色：`#1c1b1b`
- 文字次色：`#6d6865`
- 按钮主色：`#000000`（黑底白字）
- 按钮次色：透明底 + 黑字
- 边框/分割线：`#c4c7c7`
- 激活态：`transform: scale(0.98)`
- 字体：沿用现有 `--gallery-*` CSS 变量体系

## 8. 边界情况

- 用户关闭小程序再打开：若 token 未过期，直接进入首页；若已过期，API 401 触发回登录页
- 用户点击返回：登录页是入口页，无返回可点；昵称设置页可返回登录页
- 登录过程中杀进程：下次启动重新走完整流程
