# 登录引导体验优化设计文档

## 1. 概述

修复当前老用户每次登录都被强制跳转到昵称设置页的问题,优化为:

- **老用户回流**: 登录成功后静默直达首页,不再打扰
- **新用户首次**: 进入 onboarding 页(改造现有 `nickname-setup`),使用微信原生 `<input type="nickname">` 让用户**一键填入微信昵称**,可跳过
- **个人中心**: 新增"修改昵称"入口,任何时候都可以改

只处理昵称,不引入头像功能(头像沿用 dicebear 占位)。

## 2. 当前问题回顾

| 问题 | 现状 | 影响 |
|---|---|---|
| 强制昵称设置 | `pages/login/index.vue:23` 登录后无条件 `uni.reLaunch` 到 `nickname-setup` | 老用户每次清缓存重新登录都要再走一遍 |
| 后端信号被丢弃 | `auth_handler.go:44` 用 `_` 忽略 `userSvc.GetOrCreateByWxCode` 第二个返回值 `isNew` | 前端拿不到"是否新用户"信号 |
| 个人中心无编辑入口 | `pages/profile/index.vue` 显示昵称但没有修改按钮 | 用户错过登录时的设置就再也改不了 |
| 输入体验一般 | `nickname-setup` 用普通 `<input type="text">` | 用户要手打,无法一键拿到微信昵称 |

## 3. 后端改动

### 3.1 `auth_handler.go` — `LoginResponse` 增加 `is_new`

```go
type LoginResponse struct {
    AccessToken string `json:"access_token"`
    User        User   `json:"user"`
    IsNew       bool   `json:"is_new"`
}
```

`LoginHandler` 中:

```go
account, isNew, err := userSvc.GetOrCreateByWxCode(c.Request.Context(), req.Code, wxClient)
// ...
c.JSON(http.StatusOK, LoginResponse{
    AccessToken: accessToken,
    User:        User{ /* ... */ },
    IsNew:       isNew,
})
```

不需要新增数据库字段或表结构变更。

### 3.2 `auth_handler_test.go` — 测试

补充用例验证 `is_new` 字段:

- 首次调用(数据库无 openid): 期望 `is_new=true`
- 重复调用(数据库已有): 期望 `is_new=false`

其余测试断言保持不动。

## 4. 前端改动

### 4.1 `pages/login/index.vue` — 按 `is_new` 分流

```ts
const res = await login(loginRes.code)
uni.setStorageSync('access_token', res.access_token)

if (res.is_new) {
  uni.reLaunch({ url: '/pages/nickname-setup/index' })
} else {
  uni.reLaunch({ url: '/pages/home/index' })
}
```

### 4.2 `pages/nickname-setup/index.vue` — 改用微信原生昵称组件

```vue
<input
  v-model="nickname"
  type="nickname"
  placeholder="请输入昵称"
  maxlength="20"
/>
```

**变化点**: `type="text"` → `type="nickname"`。其他模板/样式/逻辑不动。

**效果**: 输入框聚焦时,微信键盘上方会出现"使用微信昵称"按钮,用户点击即填入真实微信昵称(2.21.2+ 基础库内置能力),也可手动输入或修改。

**额外保留行为**:
- 跳过按钮: 直接 `uni.reLaunch` 到首页,保留后端默认的"创作者XXX"随机昵称
- 确认按钮: 与现有逻辑一致,调 `PUT /api/me`

### 4.3 `pages/profile/index.vue` — 新增"修改昵称"入口

页面标题 `model.accountTitle` 由 `GalleryPageShell` 的 `title` slot 渲染,不直接绑事件。**新增**一个独立的昵称行,放在 hero 卡片**上方**:

```vue
<view class="profile-page__nickname-row" @click="editNickname">
  <text class="profile-page__nickname">{{ userStore.profile?.nickname || '用户' }}</text>
  <text class="profile-page__edit-icon">&#x270E;</text>
</view>
<view class="profile-page__hero">
  <!-- 原有余额/免费额度内容保持不变 -->
</view>
```

样式: 浅色文字 + 铅笔图标(`✎` U+270E),点击区域整行,与现有黑白灰风格一致。

`editNickname` 方法用 `uni.showModal` inline 修改:

```ts
async function editNickname() {
  const current = userStore.profile?.nickname || ''
  const res = await uni.showModal({
    title: '修改昵称',
    editable: true,
    placeholderText: '2-20 个字符',
    content: current,
  })
  if (!res.confirm) return
  const trimmed = (res.content || '').trim()
  if (trimmed.length < 2 || trimmed.length > 20) {
    uni.showToast({ title: '昵称长度需在 2-20 个字符之间', icon: 'none' })
    return
  }
  if (trimmed === current) return
  try {
    const updated = await updateMe(trimmed)
    userStore.setProfile(updated)
    uni.showToast({ title: '已更新', icon: 'success' })
  } catch {
    uni.showToast({ title: '更新失败,请重试', icon: 'none' })
  }
}
```

`uni.showModal` 在微信小程序原生支持 `editable: true`,无需新建编辑页。

### 4.4 `services/api.ts` — `login` 类型扩展

```ts
export function login(code: string) {
  return request<{
    access_token: string
    user: { id: number; nickname: string; balance: number; free_quota: number }
    is_new: boolean
  }>({ /* ... */ })
}
```

## 5. 流程图

```
用户点击"微信一键登录"
   │
   ▼
uni.login → POST /api/auth/login
   │
   ▼
后端返回 { access_token, user, is_new }
   │
   ├── is_new=true ─→ /pages/nickname-setup (微信昵称组件)
   │                     │
   │                     ├─ 完成 → PUT /api/me → /pages/home
   │                     └─ 跳过 → /pages/home (默认创作者XXX)
   │
   └── is_new=false ─→ /pages/home (静默)

用户在个人中心点击昵称
   │
   ▼
uni.showModal({ editable: true, content: 当前昵称 })
   │
   ├─ 取消 → 关闭
   └─ 确认 → 校验长度 → PUT /api/me → 更新 store + toast
```

## 6. 测试计划

### 6.1 后端单元测试

- `auth_handler_test.go`:
  - 首次登录返回 `is_new=true`
  - 二次登录返回 `is_new=false`
  - `is_new` 字段 JSON tag 正确

### 6.2 前端手动验证

| 用例 | 预期 |
|---|---|
| 全新 openid 用户首次登录 | 跳转 `nickname-setup` |
| `nickname-setup` 输入框聚焦 | 微信键盘上方出现"使用微信昵称"提示 |
| 点击"使用微信昵称" | 输入框自动填入微信用户名 |
| 点击"完成" | 跳到首页,profile 显示填入的昵称 |
| 点击"跳过" | 跳到首页,profile 显示"创作者XXX" |
| 老用户清缓存后重新登录 | 直接跳首页,不再经过 `nickname-setup` |
| profile 页点击昵称区域 | 弹出可编辑 modal,默认填当前昵称 |
| modal 中输入新昵称提交 | toast"已更新",profile 立即刷新 |
| modal 中长度 < 2 | toast"长度需在 2-20 个字符之间" |
| modal 中输入与原值一样 | 不调接口直接关闭 |

## 7. 兼容性

- `<input type="nickname">`: 微信基础库 2.21.2+(2022 年 5 月发布,目前覆盖度 99%+)
- `uni.showModal({ editable: true })`: 微信小程序原生支持,uni-app 透传
- 不影响现有 token、JWT、生成流程等其他模块

## 8. 不在本次范围

- 头像功能(`<button open-type="chooseAvatar">`):用户明确不要,Explore 流的自己头像继续走 dicebear 兜底
- 多端适配:本次只针对微信小程序;H5/App 端 `type="nickname"` 会降级为普通 input
- 自定义协议页面:登录页协议链接仍是占位
