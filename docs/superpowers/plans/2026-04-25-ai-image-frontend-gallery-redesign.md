# AI 图片前端艺廊化重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把现有 UniApp 小程序前端重构为统一的艺廊式体验，完成 `艺廊 / 创作 / 历史 / 我的 / 结果` 五个页面的视觉、导航和交互链路重组，同时保持现有后端接口和核心业务能力不变。

**Architecture:** 保持当前 `pages + components + services + stores` 结构不变，在页面和组件层引入一套共享的艺廊壳层、底部导航、状态组件和页面 view-model 纯函数。视觉层依赖静态场景元数据、现有 `scene_order / templates / history / packages / profile` 数据拼出页面，不新增后端字段；由于历史接口不返回 prompt，本轮结果页摘要明确退化为“场景名 + 模板键 + 生成时间 + 风格标签”。

**Tech Stack:** UniApp、Vue 3、TypeScript、Pinia、Vitest、Vite、WeChat Mini Program build target

---

## File Structure

### Shared foundation

- Create: `frontend/src/utils/navigation.ts`
  定义四个主导航 tab 的 key、label 和目标路径，供底部导航与页面壳复用。
- Create: `frontend/src/utils/__tests__/navigation.test.ts`
  覆盖导航配置与 tab 查找逻辑。
- Modify: `frontend/src/utils/scene.ts`
  从仅有 `getHero/getGallery` 扩展为场景元数据真源，提供场景展示信息、默认场景解析与艺廊分区逻辑。
- Modify: `frontend/src/utils/__tests__/scene.test.ts`
  覆盖场景元数据、hero/gallery 拆分与未知场景 fallback。
- Modify: `frontend/src/utils/generation.ts`
  增加状态标签、结果页状态推导、最近成功作品截取、历史时间格式化。
- Modify: `frontend/src/utils/__tests__/generation.test.ts`
  覆盖状态映射、结果分支与 recent 逻辑。
- Create: `frontend/src/components/layout/GalleryPageShell.vue`
  提供页面统一留白、内容区、底部安全区和 tab 导航容器。
- Create: `frontend/src/components/navigation/GalleryTabBar.vue`
  提供四项自定义底部导航。
- Create: `frontend/src/components/common/StatusBadge.vue`
  提供统一状态胶囊。
- Create: `frontend/src/components/common/EmptyStateCard.vue`
  提供统一空/错状态容器。
- Modify: `frontend/src/App.vue`
  保留 `onLaunch` 会话初始化，移除占位模板，加入全局 design tokens。
- Modify: `frontend/src/pages.json`
  调整页面标题文案和页面背景配置。

### Gallery + create flow

- Create: `frontend/src/pages/home/view-model.ts`
  基于 `scene_order / history / profile` 生成首页 hero、gallery、recent、权益卡数据。
- Create: `frontend/src/pages/home/view-model.test.ts`
  覆盖首页 hero 场景选择和 recent 作品截取。
- Create: `frontend/src/pages/scene/view-model.ts`
  负责创作页默认场景、模板校验、提交按钮文案和场景切换逻辑。
- Create: `frontend/src/pages/scene/view-model.test.ts`
  覆盖默认场景解析和必填校验。
- Modify: `frontend/src/pages/home/index.vue`
  重做艺廊首页。
- Modify: `frontend/src/pages/scene/index.vue`
  把当前“模板页”升级为完整创作页。
- Modify: `frontend/src/components/scene/SceneHeroCard.vue`
  改为策展 hero 卡。
- Modify: `frontend/src/components/scene/SceneGalleryCard.vue`
  改为作品型场景卡。
- Modify: `frontend/src/components/scene/TemplatePicker.vue`
  改为封面化模板选择区。
- Modify: `frontend/src/components/form/SceneFieldForm.vue`
  改为艺廊风格表单。

### History + profile

- Create: `frontend/src/pages/history/view-model.ts`
  负责历史筛选条、档案卡列表和状态说明。
- Create: `frontend/src/pages/history/view-model.test.ts`
  覆盖筛选、状态标签和空列表分支。
- Create: `frontend/src/pages/profile/view-model.ts`
  负责账户概览、最近作品、常用场景和套餐卡展示数据。
- Create: `frontend/src/pages/profile/view-model.test.ts`
  覆盖 quick scenes、recent 作品和套餐卡映射。
- Modify: `frontend/src/pages/history/index.vue`
  重做为作品档案页。
- Modify: `frontend/src/pages/profile/index.vue`
  重做为品牌化个人页。

### Result flow

- Create: `frontend/src/pages/result/view-model.ts`
  根据 `generation_id + historyItems` 推导结果页状态、摘要、标签和推荐作品。
- Create: `frontend/src/pages/result/view-model.test.ts`
  覆盖 `missing / pending / success / failed` 分支。
- Modify: `frontend/src/pages/result/index.vue`
  重做结果页状态壳、轮询表现和再次生成入口。
- Modify: `frontend/src/components/result/ResultPreviewCard.vue`
  改为“作品展示 + 摘要 + 双按钮”的主视觉容器。

---

### Task 1: 建立共享导航与场景元数据真源

**Files:**
- Create: `frontend/src/utils/navigation.ts`
- Create: `frontend/src/utils/__tests__/navigation.test.ts`
- Modify: `frontend/src/utils/scene.ts`
- Modify: `frontend/src/utils/__tests__/scene.test.ts`

- [ ] **Step 1: 先写 navigation 的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { PRIMARY_TABS, findPrimaryTab } from '../navigation'

describe('navigation', () => {
  it('defines the four primary tabs in gallery order', () => {
    expect(PRIMARY_TABS.map((tab) => tab.key)).toEqual([
      'gallery',
      'create',
      'history',
      'profile',
    ])
    expect(PRIMARY_TABS.map((tab) => tab.path)).toEqual([
      '/pages/home/index',
      '/pages/scene/index',
      '/pages/history/index',
      '/pages/profile/index',
    ])
  })

  it('returns the requested primary tab', () => {
    expect(findPrimaryTab('create')?.label).toBe('创作')
    expect(findPrimaryTab('profile')?.path).toBe('/pages/profile/index')
  })
})
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd frontend && npm test -- src/utils/__tests__/navigation.test.ts --run`
Expected: FAIL，提示 `Failed to resolve import "../navigation"` 或 `PRIMARY_TABS is not exported`

- [ ] **Step 3: 再写 scene 元数据的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import {
  buildScenePresentation,
  getDefaultSceneKey,
  getGalleryScenes,
  getHeroScene,
} from '../scene'

describe('scene presentation', () => {
  it('uses the first configured scene as hero and the rest as gallery', () => {
    const order = ['portrait', 'festival', 'invitation']

    expect(getHeroScene(order).key).toBe('portrait')
    expect(getGalleryScenes(order).map((scene) => scene.key)).toEqual([
      'festival',
      'invitation',
    ])
  })

  it('falls back to portrait when order is empty', () => {
    expect(getDefaultSceneKey([])).toBe('portrait')
  })

  it('builds presentation metadata for a known scene', () => {
    const scene = buildScenePresentation('poster')

    expect(scene.name).toBe('商业海报')
    expect(scene.tags).toEqual(['Editorial', 'Minimal'])
    expect(scene.eyebrow).toBe('Curated Collection')
  })
})
```

- [ ] **Step 4: 跑 scene 测试确认失败**

Run: `cd frontend && npm test -- src/utils/__tests__/scene.test.ts --run`
Expected: FAIL，提示 `buildScenePresentation is not a function` 或断言 `undefined`

- [ ] **Step 5: 写最小实现，建立导航和场景展示模型**

```ts
// frontend/src/utils/navigation.ts
export type PrimaryTabKey = 'gallery' | 'create' | 'history' | 'profile'

export interface PrimaryTab {
  key: PrimaryTabKey
  label: string
  path: string
}

export const PRIMARY_TABS: PrimaryTab[] = [
  { key: 'gallery', label: '艺廊', path: '/pages/home/index' },
  { key: 'create', label: '创作', path: '/pages/scene/index' },
  { key: 'history', label: '历史', path: '/pages/history/index' },
  { key: 'profile', label: '我的', path: '/pages/profile/index' },
]

export function findPrimaryTab(key: PrimaryTabKey) {
  return PRIMARY_TABS.find((tab) => tab.key === key)
}
```

```ts
// frontend/src/utils/scene.ts
export interface ScenePresentation {
  key: string
  name: string
  description: string
  eyebrow: string
  icon: string
  tags: string[]
  accent: 'portrait' | 'festival' | 'invitation' | 'tshirt' | 'poster'
}

const SCENE_META: Record<string, Omit<ScenePresentation, 'key'>> = {
  portrait: {
    name: '人像写真',
    description: '打造更适合展示与社交使用的质感人物作品。',
    eyebrow: 'Premium Service',
    icon: '✦',
    tags: ['Portrait', 'Minimal'],
    accent: 'portrait',
  },
  festival: {
    name: '节日海报',
    description: '把节庆氛围浓缩成一张干净、克制的视觉作品。',
    eyebrow: 'Seasonal Edit',
    icon: '✺',
    tags: ['Greeting', 'Warm Light'],
    accent: 'festival',
  },
  invitation: {
    name: '邀请函',
    description: '用更轻盈的排版和留白组织你的邀请信息。',
    eyebrow: 'Paper Studio',
    icon: '✧',
    tags: ['Stationery', 'Elegant'],
    accent: 'invitation',
  },
  tshirt: {
    name: 'T恤图案',
    description: '将主题文案转成适合服饰呈现的图案风格。',
    eyebrow: 'Graphic Lab',
    icon: '✷',
    tags: ['Print', 'Streetwear'],
    accent: 'tshirt',
  },
  poster: {
    name: '商业海报',
    description: '适合活动、餐饮和商业宣传的编辑式画面。',
    eyebrow: 'Curated Collection',
    icon: '✹',
    tags: ['Editorial', 'Minimal'],
    accent: 'poster',
  },
}

export function buildScenePresentation(key: string): ScenePresentation {
  const meta = SCENE_META[key] ?? SCENE_META.portrait
  return { key, ...meta }
}

export function getDefaultSceneKey(order: string[]): string {
  return order[0] ?? 'portrait'
}

export function getHeroScene(order: string[]): ScenePresentation {
  return buildScenePresentation(getDefaultSceneKey(order))
}

export function getGalleryScenes(order: string[]): ScenePresentation[] {
  return order.slice(1).map((key) => buildScenePresentation(key))
}
```

- [ ] **Step 6: 跑测试确认通过**

Run: `cd frontend && npm test -- src/utils/__tests__/navigation.test.ts src/utils/__tests__/scene.test.ts --run`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add frontend/src/utils/navigation.ts frontend/src/utils/__tests__/navigation.test.ts frontend/src/utils/scene.ts frontend/src/utils/__tests__/scene.test.ts
git commit -m "feat: add gallery navigation and scene metadata"
```

### Task 2: 建立共享状态、页面壳和全局设计 token

**Files:**
- Modify: `frontend/src/utils/generation.ts`
- Modify: `frontend/src/utils/__tests__/generation.test.ts`
- Create: `frontend/src/components/layout/GalleryPageShell.vue`
- Create: `frontend/src/components/navigation/GalleryTabBar.vue`
- Create: `frontend/src/components/common/StatusBadge.vue`
- Create: `frontend/src/components/common/EmptyStateCard.vue`
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/pages.json`

- [ ] **Step 1: 先写 generation 展示辅助逻辑的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import {
  formatHistoryDate,
  getResultViewState,
  getStatusMeta,
  takeRecentSuccessItems,
} from '../generation'

describe('generation display helpers', () => {
  it('maps running-like statuses to pending tone', () => {
    expect(getStatusMeta('queued')).toEqual({ label: '排队中', tone: 'pending' })
    expect(getStatusMeta('result_auditing')).toEqual({ label: '审核中', tone: 'pending' })
  })

  it('maps result records into result page states', () => {
    expect(getResultViewState(undefined)).toBe('missing')
    expect(getResultViewState({ status: 'running', resultUrl: '' })).toBe('pending')
    expect(getResultViewState({ status: 'success', resultUrl: 'https://x/y.png' })).toBe('success')
    expect(getResultViewState({ status: 'failed', resultUrl: '' })).toBe('failed')
  })

  it('returns recent successful items only', () => {
    const items = [
      { id: 1, status: 'failed', resultUrl: '' },
      { id: 2, status: 'success', resultUrl: 'https://x/2.png' },
      { id: 3, status: 'success', resultUrl: 'https://x/3.png' },
    ]

    expect(takeRecentSuccessItems(items, 2).map((item) => item.id)).toEqual([2, 3])
  })

  it('formats unix-second timestamps into yyyy.mm.dd', () => {
    expect(formatHistoryDate('1714003200')).toBe('2024.04.25')
  })
})
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd frontend && npm test -- src/utils/__tests__/generation.test.ts --run`
Expected: FAIL，提示 `getStatusMeta` 或 `formatHistoryDate` 未定义

- [ ] **Step 3: 写最小实现，让状态、结果分支和 recent 截取可复用**

```ts
// frontend/src/utils/generation.ts
export function isGenerationPending(status: string): boolean {
  return status === 'queued' || status === 'running' || status === 'result_auditing'
}

export function getStatusMeta(status: string): { label: string; tone: 'neutral' | 'pending' | 'success' | 'danger' } {
  switch (status) {
    case 'queued':
      return { label: '排队中', tone: 'pending' }
    case 'running':
      return { label: '生成中', tone: 'pending' }
    case 'result_auditing':
      return { label: '审核中', tone: 'pending' }
    case 'success':
      return { label: '已完成', tone: 'success' }
    case 'failed':
      return { label: '失败', tone: 'danger' }
    default:
      return { label: status, tone: 'neutral' }
  }
}

export function getResultViewState(item?: { status: string; resultUrl: string }): 'missing' | 'pending' | 'success' | 'failed' | 'empty' {
  if (!item) return 'missing'
  if (isGenerationPending(item.status)) return 'pending'
  if (item.status === 'success' && item.resultUrl) return 'success'
  if (item.status === 'failed') return 'failed'
  return 'empty'
}

export function takeRecentSuccessItems<T extends { status: string; resultUrl: string }>(items: T[], limit = 3): T[] {
  return items.filter((item) => item.status === 'success' && !!item.resultUrl).slice(0, limit)
}

export function formatHistoryDate(ts: string): string {
  const date = new Date(Number(ts) * 1000)
  return `${date.getFullYear()}.${String(date.getMonth() + 1).padStart(2, '0')}.${String(date.getDate()).padStart(2, '0')}`
}

export function filterHistoryItems<T extends { status: string }>(items: T[], status: string): T[] {
  if (status === 'all') return items
  return items.filter((item) => item.status === status)
}

export function findHistoryItemById<T extends { id: number }>(items: T[], id: number): T | undefined {
  return items.find((item) => item.id === id)
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `cd frontend && npm test -- src/utils/__tests__/generation.test.ts --run`
Expected: PASS

- [ ] **Step 5: 写共享页面壳、底部导航和状态组件**

```vue
<!-- frontend/src/components/navigation/GalleryTabBar.vue -->
<script setup lang="ts">
import { PRIMARY_TABS, type PrimaryTabKey } from '../../utils/navigation'

const props = defineProps<{ activeKey: PrimaryTabKey }>()

function go(path: string) {
  uni.reLaunch({ url: path })
}
</script>

<template>
  <view class="tab-bar">
    <view
      v-for="tab in PRIMARY_TABS"
      :key="tab.key"
      class="tab-bar__item"
      :class="{ 'tab-bar__item--active': tab.key === props.activeKey }"
      @click="go(tab.path)"
    >
      <text class="tab-bar__label">{{ tab.label }}</text>
    </view>
  </view>
</template>

<style scoped>
.tab-bar {
  display: flex;
  gap: 12rpx;
  padding: 24rpx 24rpx calc(24rpx + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.92);
  border-top: 1rpx solid var(--gallery-border);
}

.tab-bar__item {
  flex: 1;
  padding: 16rpx 0;
  border-radius: 999rpx;
  text-align: center;
}

.tab-bar__item--active {
  background: var(--gallery-accent);
}

.tab-bar__label {
  font-size: 22rpx;
  letter-spacing: 0.12em;
  color: var(--gallery-muted);
}

.tab-bar__item--active .tab-bar__label {
  color: #ffffff;
}
</style>
```

```vue
<!-- frontend/src/components/layout/GalleryPageShell.vue -->
<script setup lang="ts">
import type { PrimaryTabKey } from '../../utils/navigation'
import GalleryTabBar from '../navigation/GalleryTabBar.vue'

defineProps<{
  title?: string
  subtitle?: string
  activeTab?: PrimaryTabKey
}>()
</script>

<template>
  <view class="page-shell">
    <view class="page-shell__inner">
      <view v-if="title" class="page-shell__header">
        <text v-if="subtitle" class="page-shell__subtitle">{{ subtitle }}</text>
        <text class="page-shell__title">{{ title }}</text>
      </view>
      <slot />
    </view>
    <GalleryTabBar v-if="activeTab" :active-key="activeTab" />
  </view>
</template>

<style scoped>
.page-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--gallery-bg);
}

.page-shell__inner {
  flex: 1;
  padding: 32rpx 24rpx 32rpx;
}

.page-shell__header {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-bottom: 32rpx;
}

.page-shell__subtitle {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.page-shell__title {
  font-size: 48rpx;
  line-height: 1.2;
  font-weight: 600;
  color: var(--gallery-text);
}
</style>
```

```vue
<!-- frontend/src/components/common/StatusBadge.vue -->
<script setup lang="ts">
defineProps<{ label: string; tone: 'neutral' | 'pending' | 'success' | 'danger' }>()
</script>

<template>
  <view class="status-badge" :class="`status-badge--${tone}`">
    <text class="status-badge__text">{{ label }}</text>
  </view>
</template>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  background: #ece7e5;
}

.status-badge__text {
  font-size: 20rpx;
  letter-spacing: 0.08em;
  color: var(--gallery-muted);
}

.status-badge--pending {
  background: #ede9e0;
}

.status-badge--success {
  background: #e2ece7;
}

.status-badge--danger {
  background: #f4e2e2;
}
</style>
```

```vue
<!-- frontend/src/components/common/EmptyStateCard.vue -->
<script setup lang="ts">
defineProps<{ title: string; description: string; actionLabel?: string }>()
const emit = defineEmits<{ (e: 'action'): void }>()
</script>

<template>
  <view class="empty-state">
    <text class="empty-state__title">{{ title }}</text>
    <text class="empty-state__description">{{ description }}</text>
    <button v-if="actionLabel" class="empty-state__action" @click="emit('action')">{{ actionLabel }}</button>
  </view>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 40rpx 32rpx;
  border-radius: 32rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
  text-align: center;
}

.empty-state__title {
  font-size: 34rpx;
  font-weight: 600;
}

.empty-state__description {
  font-size: 24rpx;
  line-height: 1.6;
  color: var(--gallery-muted);
}

.empty-state__action {
  margin-top: 12rpx;
  background: var(--gallery-accent);
  color: #ffffff;
  border-radius: 999rpx;
}
</style>
```

- [ ] **Step 6: 注入全局 token，并更新页面配置**

```vue
<!-- frontend/src/App.vue -->
<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'
import { ensureMockSession } from './services/session'

onLaunch(() => {
  void ensureMockSession()
})
</script>

<style>
page {
  --gallery-bg: #fdf8f8;
  --gallery-surface: #ffffff;
  --gallery-surface-soft: #f4efed;
  --gallery-border: rgba(28, 27, 27, 0.08);
  --gallery-text: #1c1b1b;
  --gallery-muted: #6d6865;
  --gallery-accent: #111111;
  background: var(--gallery-bg);
  color: var(--gallery-text);
}

view,
text,
button,
image,
scroll-view,
input,
textarea {
  box-sizing: border-box;
}

button {
  border: none;
}

button::after {
  border: none;
}
</style>
```

```json
{
  "pages": [
    {
      "path": "pages/home/index",
      "style": {
        "navigationBarTitleText": "艺廊",
        "navigationBarBackgroundColor": "#fdf8f8",
        "backgroundColor": "#fdf8f8"
      }
    },
    {
      "path": "pages/scene/index",
      "style": {
        "navigationBarTitleText": "创作",
        "navigationBarBackgroundColor": "#fdf8f8",
        "backgroundColor": "#fdf8f8"
      }
    },
    {
      "path": "pages/result/index",
      "style": {
        "navigationBarTitleText": "生成结果",
        "navigationBarBackgroundColor": "#fdf8f8",
        "backgroundColor": "#fdf8f8"
      }
    },
    {
      "path": "pages/history/index",
      "style": {
        "navigationBarTitleText": "历史",
        "navigationBarBackgroundColor": "#fdf8f8",
        "backgroundColor": "#fdf8f8"
      }
    },
    {
      "path": "pages/profile/index",
      "style": {
        "navigationBarTitleText": "我的",
        "navigationBarBackgroundColor": "#fdf8f8",
        "backgroundColor": "#fdf8f8"
      }
    }
  ]
}
```

- [ ] **Step 7: 跑工具测试和类型检查**

Run: `cd frontend && npm test -- src/utils/__tests__/generation.test.ts --run && npm run type-check`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add frontend/src/utils/generation.ts frontend/src/utils/__tests__/generation.test.ts frontend/src/components/layout/GalleryPageShell.vue frontend/src/components/navigation/GalleryTabBar.vue frontend/src/components/common/StatusBadge.vue frontend/src/components/common/EmptyStateCard.vue frontend/src/App.vue frontend/src/pages.json
git commit -m "feat: add gallery shell and status primitives"
```

### Task 3: 重做艺廊页与创作页

**Files:**
- Create: `frontend/src/pages/home/view-model.ts`
- Create: `frontend/src/pages/home/view-model.test.ts`
- Create: `frontend/src/pages/scene/view-model.ts`
- Create: `frontend/src/pages/scene/view-model.test.ts`
- Modify: `frontend/src/pages/home/index.vue`
- Modify: `frontend/src/pages/scene/index.vue`
- Modify: `frontend/src/components/scene/SceneHeroCard.vue`
- Modify: `frontend/src/components/scene/SceneGalleryCard.vue`
- Modify: `frontend/src/components/scene/TemplatePicker.vue`
- Modify: `frontend/src/components/form/SceneFieldForm.vue`

- [ ] **Step 1: 先写首页 view-model 的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { buildHomeViewModel } from './view-model'

describe('home view model', () => {
  it('builds hero, gallery and recent work sections', () => {
    const vm = buildHomeViewModel({
      sceneOrder: ['portrait', 'festival', 'invitation'],
      historyItems: [
        { id: 1, sceneKey: 'festival', templateKey: 'spring', status: 'success', resultUrl: 'https://x/1.png', createdAt: '1714003200' },
        { id: 2, sceneKey: 'portrait', templateKey: 'office-pro', status: 'failed', resultUrl: '', createdAt: '1714089600' },
      ],
      profile: { balance: 5, free_quota: 2 },
    })

    expect(vm.heroScene.key).toBe('portrait')
    expect(vm.galleryScenes.map((scene) => scene.key)).toEqual(['festival', 'invitation'])
    expect(vm.recentWorks.map((item) => item.id)).toEqual([1])
    expect(vm.creditTitle).toBe('剩余额度')
    expect(vm.creditValue).toBe('2')
  })
})
```

- [ ] **Step 2: 跑首页 view-model 测试确认失败**

Run: `cd frontend && npm test -- src/pages/home/view-model.test.ts --run`
Expected: FAIL，提示 `buildHomeViewModel` 未定义

- [ ] **Step 3: 再写创作页 view-model 的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { buildScenePageModel, getSceneSubmitError, resolveSceneKey } from './view-model'

describe('scene view model', () => {
  it('uses requested scene when it exists in scene order', () => {
    expect(resolveSceneKey(['portrait', 'festival'], 'festival')).toBe('festival')
    expect(resolveSceneKey(['portrait', 'festival'], 'unknown')).toBe('portrait')
  })

  it('returns the first required field error before submit', () => {
    expect(getSceneSubmitError({
      key: 'office-pro',
      name: '通勤职业',
      sceneKey: 'portrait',
      formSchema: [
        { name: 'subject_name', label: '拍摄对象', type: 'text', required: true },
      ],
    }, {})).toBe('请填写: 拍摄对象')
  })

  it('builds a submit label from the loading state', () => {
    expect(buildScenePageModel({
      sceneOrder: ['portrait'],
      currentSceneKey: 'portrait',
      templates: [],
      selectedTemplateKey: '',
      isSubmitting: true,
    }).submitLabel).toBe('提交中...')
  })
})
```

- [ ] **Step 4: 跑创作页 view-model 测试确认失败**

Run: `cd frontend && npm test -- src/pages/scene/view-model.test.ts --run`
Expected: FAIL，提示 `resolveSceneKey` 或 `getSceneSubmitError` 未定义

- [ ] **Step 5: 写最小 view-model 实现**

```ts
// frontend/src/pages/home/view-model.ts
import type { UserProfile } from '../../store/user'
import { takeRecentSuccessItems } from '../../utils/generation'
import { getGalleryScenes, getHeroScene } from '../../utils/scene'

interface HistoryLike {
  id: number
  sceneKey: string
  templateKey: string
  status: string
  resultUrl: string
  createdAt: string
}

export function buildHomeViewModel(input: {
  sceneOrder: string[]
  historyItems: HistoryLike[]
  profile: Pick<UserProfile, 'balance' | 'free_quota'> | null
}) {
  return {
    heroScene: getHeroScene(input.sceneOrder),
    galleryScenes: getGalleryScenes(input.sceneOrder),
    recentWorks: takeRecentSuccessItems(input.historyItems, 3),
    creditTitle: '剩余额度',
    creditValue: String(input.profile?.free_quota ?? 0),
    balanceValue: String(input.profile?.balance ?? 0),
  }
}
```

```ts
// frontend/src/pages/scene/view-model.ts
import type { Template } from '../../types/scene'
import { buildScenePresentation, getDefaultSceneKey } from '../../utils/scene'

export function resolveSceneKey(sceneOrder: string[], requestedKey = ''): string {
  return sceneOrder.includes(requestedKey) ? requestedKey : getDefaultSceneKey(sceneOrder)
}

export function getSceneSubmitError(template: Template | null, formValues: Record<string, string>): string | null {
  if (!template) return '请选择模板'
  const missingField = template.formSchema.find((field) => field.required && !formValues[field.name])
  return missingField ? `请填写: ${missingField.label}` : null
}

export function buildScenePageModel(input: {
  sceneOrder: string[]
  currentSceneKey: string
  templates: Template[]
  selectedTemplateKey: string
  isSubmitting: boolean
}) {
  return {
    sceneChoices: input.sceneOrder.map((key) => buildScenePresentation(key)),
    currentScene: buildScenePresentation(input.currentSceneKey),
    submitLabel: input.isSubmitting ? '提交中...' : '开始生成作品',
    hasTemplates: input.templates.length > 0,
    selectedTemplateKey: input.selectedTemplateKey,
  }
}
```

- [ ] **Step 6: 跑 view-model 测试确认通过**

Run: `cd frontend && npm test -- src/pages/home/view-model.test.ts src/pages/scene/view-model.test.ts --run`
Expected: PASS

- [ ] **Step 7: 重做艺廊页与创作页 UI**

```vue
<!-- frontend/src/pages/home/index.vue -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import SceneHeroCard from '../../components/scene/SceneHeroCard.vue'
import SceneGalleryCard from '../../components/scene/SceneGalleryCard.vue'
import { getClientConfig, getHistory, getMe, mapHistoryItem } from '../../services/api'
import { useConfigStore } from '../../store/config'
import { useUserStore } from '../../store/user'
import { buildHomeViewModel } from './view-model'

const configStore = useConfigStore()
const userStore = useUserStore()
const historyItems = ref<ReturnType<typeof mapHistoryItem>[]>([])
const loading = ref(true)
const error = ref('')

const model = computed(() => buildHomeViewModel({
  sceneOrder: configStore.clientConfig?.scene_order ?? [],
  historyItems: historyItems.value,
  profile: userStore.profile,
}))

async function loadPage() {
  loading.value = true
  error.value = ''
  try {
    const [configRes, historyRes, meRes] = await Promise.all([
      configStore.clientConfig ? Promise.resolve(configStore.clientConfig) : getClientConfig(),
      getHistory(),
      userStore.profile ? Promise.resolve(userStore.profile) : getMe(),
    ])
    if (!configStore.clientConfig) configStore.setClientConfig(configRes)
    if (!userStore.profile) userStore.setProfile(meRes)
    historyItems.value = (historyRes.items || []).map(mapHistoryItem)
  } catch (err) {
    error.value = '艺廊加载失败，请重试'
  } finally {
    loading.value = false
  }
}

function openCreate(sceneKey: string) {
  uni.reLaunch({ url: `/pages/scene/index?scene_key=${sceneKey}` })
}

function openResult(id: number) {
  uni.navigateTo({ url: `/pages/result/index?generation_id=${id}` })
}

onMounted(loadPage)
</script>

<template>
  <GalleryPageShell active-tab="gallery">
    <EmptyStateCard
      v-if="error"
      title="艺廊暂时不可用"
      :description="error"
      action-label="重新加载"
      @action="loadPage"
    />
    <view v-else-if="!loading" class="home-page">
      <SceneHeroCard :scene="model.heroScene" @tap="openCreate" />

      <view class="home-page__section">
        <text class="home-page__eyebrow">Curated Collection</text>
        <view class="home-page__gallery">
          <SceneGalleryCard
            v-for="scene in model.galleryScenes"
            :key="scene.key"
            :scene="scene"
            @tap="openCreate"
          />
        </view>
      </view>

      <view class="home-page__section">
        <text class="home-page__eyebrow">Recent Works</text>
        <scroll-view scroll-x class="home-page__recent-list">
          <view
            v-for="item in model.recentWorks"
            :key="item.id"
            class="home-page__recent-item"
            @click="openResult(item.id)"
          >
            <image class="home-page__recent-image" :src="item.resultUrl" mode="aspectFill" />
          </view>
        </scroll-view>
      </view>

      <view class="home-page__credit-card">
        <text class="home-page__eyebrow">{{ model.creditTitle }}</text>
        <text class="home-page__credit-value">{{ model.creditValue }}</text>
        <text class="home-page__credit-meta">余额 {{ model.balanceValue }}</text>
      </view>
    </view>
  </GalleryPageShell>
</template>
```

```vue
<!-- frontend/src/pages/scene/index.vue -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import TemplatePicker from '../../components/scene/TemplatePicker.vue'
import SceneFieldForm from '../../components/form/SceneFieldForm.vue'
import { getClientConfig } from '../../services/api'
import { useConfigStore } from '../../store/config'
import { useGenerationStore } from '../../store/generation'
import { buildScenePageModel, getSceneSubmitError, resolveSceneKey } from './view-model'

const configStore = useConfigStore()
const generationStore = useGenerationStore()
const requestedSceneKey = ref('')
const pageLoading = ref(true)
const pageError = ref('')

const currentSceneKey = computed(() =>
  resolveSceneKey(configStore.clientConfig?.scene_order ?? [], requestedSceneKey.value),
)

const model = computed(() => buildScenePageModel({
  sceneOrder: configStore.clientConfig?.scene_order ?? [],
  currentSceneKey: currentSceneKey.value,
  templates: generationStore.templates,
  selectedTemplateKey: generationStore.selectedTemplate?.key ?? '',
  isSubmitting: generationStore.isSubmitting,
}))

onLoad((query: any) => {
  requestedSceneKey.value = query.scene_key || ''
})

async function loadScene() {
  pageLoading.value = true
  pageError.value = ''
  try {
    if (!configStore.clientConfig) {
      configStore.setClientConfig(await getClientConfig())
    }
    await generationStore.loadTemplates(currentSceneKey.value)
  } catch (err) {
    pageError.value = '创作页加载失败，请重试'
  } finally {
    pageLoading.value = false
  }
}

async function switchScene(sceneKey: string) {
  requestedSceneKey.value = sceneKey
  await generationStore.loadTemplates(sceneKey)
}

async function handleSubmit() {
  const error = getSceneSubmitError(generationStore.selectedTemplate, generationStore.formValues)
  if (error) {
    uni.showToast({ title: error, icon: 'none' })
    return
  }
  const generationId = await generationStore.submitGeneration()
  uni.navigateTo({ url: `/pages/result/index?generation_id=${generationId}` })
}

onMounted(loadScene)
</script>

<template>
  <GalleryPageShell active-tab="create" :title="model.currentScene.name" :subtitle="model.currentScene.eyebrow">
    <EmptyStateCard
      v-if="pageError"
      title="创作页暂时不可用"
      :description="pageError"
      action-label="重新加载"
      @action="loadScene"
    />
    <view v-else-if="!pageLoading" class="scene-page">
      <scroll-view scroll-x class="scene-page__scene-strip">
        <view
          v-for="scene in model.sceneChoices"
          :key="scene.key"
          class="scene-page__scene-pill"
          :class="{ 'scene-page__scene-pill--active': scene.key === currentSceneKey }"
          @click="switchScene(scene.key)"
        >
          <text>{{ scene.name }}</text>
        </view>
      </scroll-view>

      <TemplatePicker
        :scene="model.currentScene"
        :templates="generationStore.templates"
        :selected-key="generationStore.selectedTemplate?.key ?? ''"
        @select="generationStore.setTemplate"
      />

      <SceneFieldForm
        v-if="generationStore.selectedTemplate"
        :schema="generationStore.selectedTemplate.formSchema"
        :model-value="generationStore.formValues"
        @update:model-value="generationStore.setFormValues"
      />

      <button class="scene-page__submit" :disabled="generationStore.isSubmitting" @click="handleSubmit">
        {{ model.submitLabel }}
      </button>
    </view>
  </GalleryPageShell>
</template>
```

- [ ] **Step 8: 补上 scene 组件的最小结构升级**

```vue
<!-- frontend/src/components/scene/SceneHeroCard.vue -->
<script setup lang="ts">
import type { ScenePresentation } from '../../utils/scene'

defineProps<{ scene: ScenePresentation }>()
const emit = defineEmits<{ (e: 'tap', key: string): void }>()
</script>

<template>
  <view class="hero-card" :class="`hero-card--${scene.accent}`" @click="emit('tap', scene.key)">
    <text class="hero-card__eyebrow">{{ scene.eyebrow }}</text>
    <text class="hero-card__title">{{ scene.name }}</text>
    <text class="hero-card__description">{{ scene.description }}</text>
    <view class="hero-card__tags">
      <text v-for="tag in scene.tags" :key="tag" class="hero-card__tag">{{ tag }}</text>
    </view>
    <button class="hero-card__cta">进入创作</button>
  </view>
</template>

<style scoped>
.hero-card {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 40rpx 32rpx;
  border-radius: 36rpx;
  color: #ffffff;
  background: linear-gradient(155deg, #151515 0%, #7a6a61 100%);
}

.hero-card--festival {
  background: linear-gradient(155deg, #5d3426 0%, #c38f62 100%);
}

.hero-card--invitation {
  background: linear-gradient(155deg, #b9a89d 0%, #efe8e2 100%);
  color: #1c1b1b;
}

.hero-card__eyebrow {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.hero-card__title {
  font-size: 56rpx;
  font-weight: 700;
}

.hero-card__description {
  font-size: 26rpx;
  line-height: 1.6;
}

.hero-card__tags {
  display: flex;
  gap: 12rpx;
  flex-wrap: wrap;
}

.hero-card__tag,
.hero-card__cta {
  border-radius: 999rpx;
}

.hero-card__cta {
  margin-top: 8rpx;
  background: rgba(255, 255, 255, 0.16);
  color: inherit;
}
</style>
```

```vue
<!-- frontend/src/components/scene/SceneGalleryCard.vue -->
<script setup lang="ts">
import type { ScenePresentation } from '../../utils/scene'

defineProps<{ scene: ScenePresentation }>()
const emit = defineEmits<{ (e: 'tap', key: string): void }>()
</script>

<template>
  <view class="gallery-card" :class="`gallery-card--${scene.accent}`" @click="emit('tap', scene.key)">
    <text class="gallery-card__eyebrow">{{ scene.eyebrow }}</text>
    <text class="gallery-card__name">{{ scene.name }}</text>
    <text class="gallery-card__description">{{ scene.description }}</text>
  </view>
</template>

<style scoped>
.gallery-card {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  min-height: 260rpx;
  padding: 28rpx 24rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.gallery-card__eyebrow {
  font-size: 20rpx;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.gallery-card__name {
  font-size: 32rpx;
  font-weight: 600;
}

.gallery-card__description {
  font-size: 24rpx;
  line-height: 1.6;
  color: var(--gallery-muted);
}
</style>
```

```vue
<!-- frontend/src/components/scene/TemplatePicker.vue -->
<script setup lang="ts">
import type { Template } from '../../types/scene'
import type { ScenePresentation } from '../../utils/scene'

defineProps<{
  scene: ScenePresentation
  templates: Template[]
  selectedKey?: string
}>()
const emit = defineEmits<{ (e: 'select', template: Template): void }>()
</script>

<template>
  <view class="template-picker">
    <text class="template-picker__eyebrow">{{ scene.eyebrow }}</text>
    <text class="template-picker__title">选择模板</text>
    <view class="template-picker__list">
      <view
        v-for="template in templates"
        :key="template.key"
        class="template-card"
        :class="{ 'template-card--active': selectedKey === template.key }"
        @click="emit('select', template)"
      >
        <image v-if="template.sampleImageUrl" class="template-card__image" :src="template.sampleImageUrl" mode="aspectFill" />
        <view v-else class="template-card__image template-card__image--fallback">
          <text class="template-card__icon">{{ scene.icon }}</text>
        </view>
        <text class="template-card__name">{{ template.name }}</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.template-picker {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.template-picker__eyebrow {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.template-picker__title {
  font-size: 40rpx;
  font-weight: 600;
}

.template-picker__list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.template-card {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  padding: 16rpx;
  border-radius: 24rpx;
  background: var(--gallery-surface);
  border: 1rpx solid transparent;
}

.template-card--active {
  border-color: var(--gallery-text);
}

.template-card__image {
  width: 100%;
  height: 220rpx;
  border-radius: 18rpx;
  background: var(--gallery-surface-soft);
}

.template-card__image--fallback {
  display: flex;
  align-items: center;
  justify-content: center;
}

.template-card__name {
  font-size: 26rpx;
  color: var(--gallery-text);
}
</style>
```

```vue
<!-- frontend/src/components/form/SceneFieldForm.vue -->
<script setup lang="ts">
import { ref, watch } from 'vue'
import type { FormField } from '../../types/scene'

interface Props {
  schema: FormField[]
  modelValue: Record<string, string>
}

const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'update:modelValue', values: Record<string, string>): void }>()
const localValues = ref<Record<string, string>>({ ...props.modelValue })

watch(
  () => props.schema,
  () => {
    const next: Record<string, string> = {}
    for (const field of props.schema) next[field.name] = localValues.value[field.name] ?? ''
    localValues.value = next
    emit('update:modelValue', next)
  },
  { immediate: true },
)

watch(
  () => props.modelValue,
  (value) => {
    localValues.value = { ...value }
  },
  { deep: true },
)

function updateField(name: string, value: string) {
  localValues.value[name] = value
  emit('update:modelValue', { ...localValues.value })
}
</script>

<template>
  <view class="scene-field-form">
    <view v-for="field in schema" :key="field.name" class="scene-field-form__item">
      <text class="scene-field-form__label">{{ field.label }}{{ field.required ? ' *' : '' }}</text>
      <input
        v-if="field.type === 'text' || field.type === 'date'"
        :type="field.type === 'date' ? 'date' : 'text'"
        class="scene-field-form__input"
        :value="localValues[field.name] ?? ''"
        :placeholder="field.label"
        @input="updateField(field.name, ($event as any).detail.value)"
      />
      <textarea
        v-else-if="field.type === 'textarea'"
        class="scene-field-form__textarea"
        :value="localValues[field.name] ?? ''"
        :placeholder="field.label"
        @input="updateField(field.name, ($event as any).detail.value)"
      />
      <picker
        v-else-if="field.type === 'select' && field.options"
        mode="selector"
        :range="field.options"
        :value="Math.max(0, field.options.indexOf(localValues[field.name] ?? ''))"
        @change="updateField(field.name, field.options![($event as any).detail.value] as string)"
      >
        <view class="scene-field-form__input">{{ localValues[field.name] || '请选择' }}</view>
      </picker>
    </view>
  </view>
</template>

<style scoped>
.scene-field-form {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.scene-field-form__item {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}

.scene-field-form__label {
  font-size: 24rpx;
  color: var(--gallery-muted);
}

.scene-field-form__input,
.scene-field-form__textarea {
  width: 100%;
  padding: 24rpx;
  border-radius: 20rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
  font-size: 28rpx;
  color: var(--gallery-text);
}

.scene-field-form__textarea {
  min-height: 180rpx;
}
</style>
```

- [ ] **Step 9: 跑页面相关测试和类型检查**

Run: `cd frontend && npm test -- src/pages/home/view-model.test.ts src/pages/scene/view-model.test.ts src/store/__tests__/generation.test.ts --run && npm run type-check`
Expected: PASS

- [ ] **Step 10: Commit**

```bash
git add frontend/src/pages/home/view-model.ts frontend/src/pages/home/view-model.test.ts frontend/src/pages/scene/view-model.ts frontend/src/pages/scene/view-model.test.ts frontend/src/pages/home/index.vue frontend/src/pages/scene/index.vue frontend/src/components/scene/SceneHeroCard.vue frontend/src/components/scene/SceneGalleryCard.vue frontend/src/components/scene/TemplatePicker.vue frontend/src/components/form/SceneFieldForm.vue
git commit -m "feat: redesign gallery and create pages"
```

### Task 4: 重做历史页与我的页

**Files:**
- Create: `frontend/src/pages/history/view-model.ts`
- Create: `frontend/src/pages/history/view-model.test.ts`
- Create: `frontend/src/pages/profile/view-model.ts`
- Create: `frontend/src/pages/profile/view-model.test.ts`
- Modify: `frontend/src/pages/history/index.vue`
- Modify: `frontend/src/pages/profile/index.vue`

- [ ] **Step 1: 先写历史页 view-model 的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { buildHistoryViewModel } from './view-model'

describe('history view model', () => {
  it('filters archive cards and maps status labels', () => {
    const vm = buildHistoryViewModel({
      items: [
        { id: 1, sceneKey: 'portrait', templateKey: 'office-pro', status: 'success', resultUrl: 'https://x/1.png', createdAt: '1714003200' },
        { id: 2, sceneKey: 'festival', templateKey: 'spring', status: 'running', resultUrl: '', createdAt: '1714089600' },
      ],
      filter: 'running',
    })

    expect(vm.cards).toHaveLength(1)
    expect(vm.cards[0]?.status.label).toBe('生成中')
    expect(vm.cards[0]?.title).toBe('节日海报')
  })
})
```

- [ ] **Step 2: 跑历史页测试确认失败**

Run: `cd frontend && npm test -- src/pages/history/view-model.test.ts --run`
Expected: FAIL，提示 `buildHistoryViewModel` 未定义

- [ ] **Step 3: 再写我的页 view-model 的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { buildProfileViewModel } from './view-model'

describe('profile view model', () => {
  it('builds quick scenes, recent works and package cards', () => {
    const vm = buildProfileViewModel({
      profile: { balance: 8, free_quota: 3 },
      packages: [{ code: 'pack10', title: '10次包', price: '8.00', count: 10 }],
      historyItems: [
        { id: 1, sceneKey: 'poster', templateKey: 'concert', status: 'success', resultUrl: 'https://x/1.png', createdAt: '1714003200' },
      ],
      sceneOrder: ['portrait', 'poster'],
    })

    expect(vm.accountTitle).toBe('我的作品室')
    expect(vm.quickScenes.map((scene) => scene.key)).toEqual(['portrait', 'poster'])
    expect(vm.recentWorks).toHaveLength(1)
    expect(vm.packages[0]?.actionLabel).toBe('购买')
  })
})
```

- [ ] **Step 4: 跑我的页测试确认失败**

Run: `cd frontend && npm test -- src/pages/profile/view-model.test.ts --run`
Expected: FAIL，提示 `buildProfileViewModel` 未定义

- [ ] **Step 5: 写最小 view-model 实现**

```ts
// frontend/src/pages/history/view-model.ts
import { filterHistoryItems, formatHistoryDate, getStatusMeta } from '../../utils/generation'
import { buildScenePresentation } from '../../utils/scene'

interface HistoryLike {
  id: number
  sceneKey: string
  templateKey: string
  status: string
  resultUrl: string
  createdAt: string
}

export function buildHistoryViewModel(input: { items: HistoryLike[]; filter: string }) {
  return {
    filters: ['all', 'queued', 'running', 'result_auditing', 'success', 'failed'],
    cards: filterHistoryItems(input.items, input.filter).map((item) => ({
      id: item.id,
      title: buildScenePresentation(item.sceneKey).name,
      subtitle: `模板 · ${item.templateKey}`,
      date: formatHistoryDate(item.createdAt),
      imageUrl: item.resultUrl,
      status: getStatusMeta(item.status),
    })),
  }
}
```

```ts
// frontend/src/pages/profile/view-model.ts
import { takeRecentSuccessItems } from '../../utils/generation'
import { buildScenePresentation } from '../../utils/scene'

export function buildProfileViewModel(input: {
  profile: { balance: number; free_quota: number } | null
  packages: { code: string; title: string; price: string; count: number }[]
  historyItems: { id: number; sceneKey: string; templateKey: string; status: string; resultUrl: string; createdAt: string }[]
  sceneOrder: string[]
}) {
  return {
    accountTitle: '我的作品室',
    balance: String(input.profile?.balance ?? 0),
    freeQuota: String(input.profile?.free_quota ?? 0),
    recentWorks: takeRecentSuccessItems(input.historyItems, 4),
    quickScenes: input.sceneOrder.slice(0, 3).map((key) => buildScenePresentation(key)),
    packages: input.packages.map((item) => ({
      ...item,
      actionLabel: '购买',
    })),
  }
}
```

- [ ] **Step 6: 跑 view-model 测试确认通过**

Run: `cd frontend && npm test -- src/pages/history/view-model.test.ts src/pages/profile/view-model.test.ts --run`
Expected: PASS

- [ ] **Step 7: 重做历史页与我的页**

```vue
<!-- frontend/src/pages/history/index.vue -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import { getHistory, mapHistoryItem } from '../../services/api'
import { buildHistoryViewModel } from './view-model'

const items = ref<ReturnType<typeof mapHistoryItem>[]>([])
const filter = ref('all')
const loading = ref(false)
const error = ref('')

const model = computed(() => buildHistoryViewModel({ items: items.value, filter: filter.value }))

async function loadHistory() {
  loading.value = true
  error.value = ''
  try {
    const res = await getHistory()
    items.value = (res.items || []).map(mapHistoryItem)
  } catch (err) {
    error.value = '历史记录加载失败，请重试'
  } finally {
    loading.value = false
  }
}

function openResult(id: number) {
  uni.navigateTo({ url: `/pages/result/index?generation_id=${id}` })
}

onMounted(loadHistory)
</script>

<template>
  <GalleryPageShell active-tab="history" title="历史档案" subtitle="Archive">
    <EmptyStateCard
      v-if="error"
      title="历史加载失败"
      :description="error"
      action-label="重新加载"
      @action="loadHistory"
    />
    <view v-else class="history-page">
      <scroll-view scroll-x class="history-page__filters">
        <view
          v-for="status in model.filters"
          :key="status"
          class="history-page__filter"
          :class="{ 'history-page__filter--active': filter === status }"
          @click="filter = status"
        >
          <text>{{ status }}</text>
        </view>
      </scroll-view>

      <EmptyStateCard
        v-if="!loading && model.cards.length === 0"
        title="还没有作品记录"
        description="从创作页开始生成第一张作品。"
        action-label="去创作"
        @action="uni.reLaunch({ url: '/pages/scene/index' })"
      />

      <view v-else class="history-page__list">
        <view
          v-for="card in model.cards"
          :key="card.id"
          class="history-page__card"
          @click="openResult(card.id)"
        >
          <image
            v-if="card.imageUrl"
            class="history-page__image"
            :src="card.imageUrl"
            mode="aspectFill"
          />
          <view v-else class="history-page__image history-page__image--placeholder"></view>
          <view class="history-page__info">
            <text class="history-page__title">{{ card.title }}</text>
            <text class="history-page__subtitle">{{ card.subtitle }}</text>
            <StatusBadge :label="card.status.label" :tone="card.status.tone" />
            <text class="history-page__date">{{ card.date }}</text>
          </view>
        </view>
      </view>
    </view>
  </GalleryPageShell>
</template>
```

```vue
<!-- frontend/src/pages/profile/index.vue -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import { createOrder, getClientConfig, getHistory, getMe, getPackages, mapHistoryItem } from '../../services/api'
import { useConfigStore } from '../../store/config'
import { useUserStore } from '../../store/user'
import { buildProfileViewModel } from './view-model'

const configStore = useConfigStore()
const userStore = useUserStore()
const packagesList = ref<{ code: string; title: string; price: string; count: number }[]>([])
const historyItems = ref<ReturnType<typeof mapHistoryItem>[]>([])
const loading = ref(false)

const model = computed(() => buildProfileViewModel({
  profile: userStore.profile,
  packages: packagesList.value,
  historyItems: historyItems.value,
  sceneOrder: configStore.clientConfig?.scene_order ?? ['portrait', 'festival', 'invitation'],
}))

async function loadProfilePage() {
  const [profile, packagesRes, historyRes] = await Promise.all([
    userStore.profile ? Promise.resolve(userStore.profile) : getMe(),
    getPackages(),
    getHistory(),
  ])
  if (!userStore.profile) userStore.setProfile(profile)
  if (!configStore.clientConfig) configStore.setClientConfig(await getClientConfig())
  packagesList.value = packagesRes.packages || []
  historyItems.value = (historyRes.items || []).map(mapHistoryItem)
}

async function handleBuy(packageCode: string) {
  loading.value = true
  try {
    const order = await createOrder(packageCode)
    uni.showModal({ title: '模拟支付', content: `订单号: ${order.order_no}\n金额: ${order.amount}`, showCancel: false })
  } finally {
    loading.value = false
  }
}

onMounted(loadProfilePage)
</script>

<template>
  <GalleryPageShell active-tab="profile" title="我的作品室" subtitle="Profile">
    <view class="profile-page">
      <view class="profile-page__hero">
        <text class="profile-page__balance">{{ model.balance }}</text>
        <text class="profile-page__meta">余额</text>
        <text class="profile-page__quota">免费额度 {{ model.freeQuota }}</text>
      </view>

      <view class="profile-page__section">
        <text class="profile-page__eyebrow">Recent Works</text>
        <scroll-view scroll-x class="profile-page__recent-list">
          <view
            v-for="item in model.recentWorks"
            :key="item.id"
            class="profile-page__recent-item"
            @click="uni.navigateTo({ url: `/pages/result/index?generation_id=${item.id}` })"
          >
            <image class="profile-page__recent-image" :src="item.resultUrl" mode="aspectFill" />
          </view>
        </scroll-view>
      </view>

      <view class="profile-page__section">
        <text class="profile-page__eyebrow">Quick Scenes</text>
        <view class="profile-page__quick-scenes">
          <view
            v-for="scene in model.quickScenes"
            :key="scene.key"
            class="profile-page__quick-scene"
            @click="uni.reLaunch({ url: `/pages/scene/index?scene_key=${scene.key}` })"
          >
            <text>{{ scene.name }}</text>
          </view>
        </view>
      </view>

      <view class="profile-page__section">
        <text class="profile-page__eyebrow">Packages</text>
        <view
          v-for="pkg in model.packages"
          :key="pkg.code"
          class="profile-page__package"
        >
          <view>
            <text class="profile-page__package-title">{{ pkg.title }}</text>
            <text class="profile-page__package-meta">¥{{ pkg.price }} / {{ pkg.count }} 次</text>
          </view>
          <button :disabled="loading" @click="handleBuy(pkg.code)">{{ pkg.actionLabel }}</button>
        </view>
      </view>
    </view>
  </GalleryPageShell>
</template>
```

- [ ] **Step 8: 跑测试和类型检查**

Run: `cd frontend && npm test -- src/pages/history/view-model.test.ts src/pages/profile/view-model.test.ts --run && npm run type-check`
Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add frontend/src/pages/history/view-model.ts frontend/src/pages/history/view-model.test.ts frontend/src/pages/profile/view-model.ts frontend/src/pages/profile/view-model.test.ts frontend/src/pages/history/index.vue frontend/src/pages/profile/index.vue
git commit -m "feat: redesign history and profile pages"
```

### Task 5: 重做结果页并完成整体验证

**Files:**
- Create: `frontend/src/pages/result/view-model.ts`
- Create: `frontend/src/pages/result/view-model.test.ts`
- Modify: `frontend/src/pages/result/index.vue`
- Modify: `frontend/src/components/result/ResultPreviewCard.vue`

- [ ] **Step 1: 先写结果页 view-model 的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { buildResultViewModel } from './view-model'

describe('result view model', () => {
  const items = [
    { id: 1, sceneKey: 'portrait', templateKey: 'office-pro', status: 'running', resultUrl: '', createdAt: '1714003200' },
    { id: 2, sceneKey: 'poster', templateKey: 'concert', status: 'success', resultUrl: 'https://x/2.png', createdAt: '1714089600' },
    { id: 3, sceneKey: 'festival', templateKey: 'spring', status: 'failed', resultUrl: '', createdAt: '1714176000' },
  ]

  it('builds pending state for in-progress generations', () => {
    expect(buildResultViewModel(items, 1).state).toBe('pending')
  })

  it('builds success summary and recommendations', () => {
    const vm = buildResultViewModel(items, 2)
    expect(vm.state).toBe('success')
    expect(vm.title).toBe('商业海报')
    expect(vm.summary).toContain('concert')
    expect(vm.recommendations.map((item) => item.id)).toEqual([2])
  })

  it('returns missing when generation is absent', () => {
    expect(buildResultViewModel(items, 999).state).toBe('missing')
  })
})
```

- [ ] **Step 2: 跑结果页测试确认失败**

Run: `cd frontend && npm test -- src/pages/result/view-model.test.ts --run`
Expected: FAIL，提示 `buildResultViewModel` 未定义

- [ ] **Step 3: 写最小 view-model 实现**

```ts
// frontend/src/pages/result/view-model.ts
import { findHistoryItemById, formatHistoryDate, getResultViewState, takeRecentSuccessItems } from '../../utils/generation'
import { buildScenePresentation } from '../../utils/scene'

interface HistoryLike {
  id: number
  sceneKey: string
  templateKey: string
  status: string
  resultUrl: string
  createdAt: string
}

export function buildResultViewModel(items: HistoryLike[], generationId: number) {
  const currentItem = findHistoryItemById(items, generationId)
  const state = getResultViewState(currentItem)

  if (!currentItem) {
    return { state, title: '未找到生成记录', summary: '', chips: [], recommendations: [] as HistoryLike[] }
  }

  const scene = buildScenePresentation(currentItem.sceneKey)

  return {
    state,
    title: scene.name,
    summary: `${currentItem.templateKey} · ${formatHistoryDate(currentItem.createdAt)}`,
    chips: scene.tags,
    currentItem,
    recommendations: takeRecentSuccessItems(items, 4),
  }
}
```

- [ ] **Step 4: 跑 view-model 测试确认通过**

Run: `cd frontend && npm test -- src/pages/result/view-model.test.ts --run`
Expected: PASS

- [ ] **Step 5: 重做结果页和主视觉卡**

```vue
<!-- frontend/src/components/result/ResultPreviewCard.vue -->
<script setup lang="ts">
defineProps<{
  imageUrl: string
  title: string
  summary: string
  chips: string[]
}>()

const emit = defineEmits<{
  (e: 'save'): void
  (e: 'share'): void
}>()
</script>

<template>
  <view class="result-card">
    <image class="result-card__image" :src="imageUrl" mode="aspectFit" @click="uni.previewImage({ urls: [imageUrl], current: imageUrl })" />
    <view class="result-card__body">
      <text class="result-card__label">Creative Summary</text>
      <text class="result-card__title">{{ title }}</text>
      <text class="result-card__summary">{{ summary }}</text>
      <view class="result-card__chips">
        <text v-for="chip in chips" :key="chip" class="result-card__chip">{{ chip }}</text>
      </view>
      <button class="result-card__primary" @click="emit('save')">保存到相册</button>
      <button class="result-card__secondary" @click="emit('share')">微信分享</button>
    </view>
  </view>
</template>
```

```vue
<!-- frontend/src/pages/result/index.vue -->
<script setup lang="ts">
import { computed, onMounted, ref, watchEffect } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import ResultPreviewCard from '../../components/result/ResultPreviewCard.vue'
import { getHistory, mapHistoryItem } from '../../services/api'
import { mapTrackingEvent, track } from '../../services/tracking'
import { buildResultViewModel } from './view-model'

const generationId = ref(0)
const historyItems = ref<ReturnType<typeof mapHistoryItem>[]>([])

const model = computed(() => buildResultViewModel(historyItems.value, generationId.value))

onLoad((query: any) => {
  generationId.value = Number(query.generation_id || 0)
})

async function refreshHistory() {
  const res = await getHistory()
  historyItems.value = (res.items || []).map(mapHistoryItem)
}

watchEffect((cleanup) => {
  if (model.value.state === 'pending') {
    const timer = setTimeout(() => {
      void refreshHistory()
    }, 1500)
    cleanup(() => clearTimeout(timer))
  }
})

async function onSave() {
  await track(mapTrackingEvent('save'), { generation_id: generationId.value })
  uni.showToast({ title: '已保存', icon: 'success' })
}

async function onShare() {
  await track(mapTrackingEvent('share'), { generation_id: generationId.value })
  uni.showShareMenu({ withShareTicket: true })
}

function onGenerateAnother() {
  const sceneKey = model.value.currentItem?.sceneKey ?? 'portrait'
  uni.reLaunch({ url: `/pages/scene/index?scene_key=${sceneKey}` })
}

onMounted(refreshHistory)
</script>

<template>
  <GalleryPageShell :title="model.title" subtitle="Generated Artwork">
    <EmptyStateCard
      v-if="model.state === 'missing'"
      title="未找到生成记录"
      description="请回到创作页重新生成。"
      action-label="去创作"
      @action="onGenerateAnother"
    />

    <view v-else-if="model.state === 'pending'" class="result-page__state">
      <text class="result-page__pending-title">作品正在生成中</text>
      <text class="result-page__pending-desc">系统会自动刷新当前状态，请稍候。</text>
    </view>

    <EmptyStateCard
      v-else-if="model.state === 'failed'"
      title="本次生成失败"
      description="可以保留当前场景设置，再试一次。"
      action-label="再来一张"
      @action="onGenerateAnother"
    />

    <view v-else-if="model.state === 'success'" class="result-page">
      <ResultPreviewCard
        :image-url="model.currentItem!.resultUrl"
        :title="model.title"
        :summary="model.summary"
        :chips="model.chips"
        @save="onSave"
        @share="onShare"
      />

      <view class="result-page__section">
        <text class="result-page__eyebrow">Try Again</text>
        <button class="result-page__again" @click="onGenerateAnother">再来一张</button>
      </view>

      <view class="result-page__section">
        <text class="result-page__eyebrow">Recent Works</text>
        <scroll-view scroll-x class="result-page__recent-list">
          <view
            v-for="item in model.recommendations"
            :key="item.id"
            class="result-page__recent-item"
            @click="uni.navigateTo({ url: `/pages/result/index?generation_id=${item.id}` })"
          >
            <image class="result-page__recent-image" :src="item.resultUrl" mode="aspectFill" />
          </view>
        </scroll-view>
      </view>
    </view>
  </GalleryPageShell>
</template>
```

- [ ] **Step 6: 跑完整前端验证**

Run: `cd frontend && npm test -- --run && npm run type-check && npm run build -- --platform mp-weixin`
Expected: PASS，最后输出 `Build complete` 或成功产物写入 `dist/build/mp-weixin`

- [ ] **Step 7: 做一次微信小程序人工走查**

Run: `make dev-frontend`
Expected: 生成 `frontend/dist/build/mp-weixin`，然后在微信开发者工具中确认以下路径：
- `艺廊 -> 创作 -> 结果`
- `艺廊 -> 最近作品 -> 结果`
- `历史 -> 结果`
- `我的 -> 快捷场景 -> 创作`
- `结果 -> 再来一张`

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/result/view-model.ts frontend/src/pages/result/view-model.test.ts frontend/src/pages/result/index.vue frontend/src/components/result/ResultPreviewCard.vue
git commit -m "feat: redesign result experience"
```

---

## Self-Review Checklist

- Spec coverage:
  - `信息架构` 由 Task 2 的壳层和 Task 3/4 的页面重组落地。
  - `艺廊 / 创作 / 历史 / 我的 / 结果` 五页均有对应任务。
  - `结果页摘要不依赖 prompt` 的边界在 Task 5 明确处理。
  - `加载/空/错状态` 通过 `EmptyStateCard`、`StatusBadge` 和页面状态分支覆盖。
  - `自动化验证 + 小程序人工走查` 在 Task 5 明确要求。
- Placeholder scan:
  - 未使用 `TODO / TBD / similar to task N / add tests later`。
  - 每个任务都给了具体文件、代码和命令。
- Type consistency:
  - `PrimaryTabKey`、`ScenePresentation`、`getStatusMeta`、`build*ViewModel` 命名在任务间保持一致。
  - `scene_order / sceneKey / templateKey / resultUrl / createdAt` 与现有 DTO/映射字段一致。
