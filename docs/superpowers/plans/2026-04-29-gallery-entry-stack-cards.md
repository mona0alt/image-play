# Gallery Entry Stack Cards Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把艺廊首页入口区从 `主卡 + 面相分析工具卡 + 两列场景卡` 重构为统一的叠册卡组，并保持现有数据边界与跳转链路不变。

**Architecture:** 保持当前 `pages + components + utils` 结构不变，在首页下新增两个纯逻辑文件：一个负责把 `scene_order` 和 `face-reading` 入口组装成统一的展示卡序列，另一个负责叠册卡组的可见窗口与滑动状态机。页面层只负责拉取数据和路由跳转，新建 `GalleryEntryStack.vue` 消费这些纯函数并承载银行卡材质感的 UI。

**Tech Stack:** UniApp、Vue 3 `<script setup>`、TypeScript、Vitest、vue-tsc、WeChat Mini Program touch events

---

## File Structure

### Home entry data model

- Create: `frontend/src/pages/home/entry-cards.ts`
  负责把 `scene_order` 和 `face-reading` 合成为首页入口卡序列，并定义卡片展示元数据、accent 和跳转路径。
- Create: `frontend/src/pages/home/entry-cards.test.ts`
  覆盖 `face-reading` 插入顺序、默认场景回退和路由目标。
- Modify: `frontend/src/pages/home/view-model.ts`
  从输出 `heroScene + galleryScenes` 改为输出 `entryCards + recentWorks + credit`。
- Modify: `frontend/src/pages/home/view-model.test.ts`
  调整为断言新的 `entryCards` 结构，而不是旧的 hero/gallery 模型。

### Stack interaction state

- Create: `frontend/src/pages/home/entry-stack.ts`
  负责叠册卡组的纯逻辑：激活索引裁剪、可见卡窗口和纵向滑动后的索引变化。
- Create: `frontend/src/pages/home/entry-stack.test.ts`
  覆盖 `active + peek cards` 计算、上下滑阈值和首尾边界。

### UI integration

- Create: `frontend/src/components/home/GalleryEntryStack.vue`
  渲染激活卡与露边卡，应用银行卡材质感样式，消费 `entry-stack.ts` 的逻辑，并向页面发出 `open` 事件。
- Modify: `frontend/src/pages/home/index.vue`
  移除 `SceneHeroCard`、独立工具卡和 `SceneGalleryCard` 列表，改为单一 `GalleryEntryStack` 入口区，并按卡片 `kind` 处理场景页/面相分析页跳转。

---

### Task 1: 建立首页入口卡展示模型

**Files:**
- Create: `frontend/src/pages/home/entry-cards.ts`
- Create: `frontend/src/pages/home/entry-cards.test.ts`
- Modify: `frontend/src/pages/home/view-model.ts`
- Modify: `frontend/src/pages/home/view-model.test.ts`

- [ ] **Step 1: 先写 `entry-cards` 的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { buildHomeEntryCards, HOME_DEFAULT_SCENE_ORDER } from './entry-cards'

describe('home entry cards', () => {
  it('inserts face reading right after the lead scene', () => {
    const cards = buildHomeEntryCards(['portrait', 'festival', 'invitation'])

    expect(cards.map((card) => card.key)).toEqual([
      'portrait',
      'face-reading',
      'festival',
      'invitation',
    ])
    expect(cards[1]).toMatchObject({
      kind: 'tool',
      title: '面相分析',
      accent: 'analysis',
      path: '/pages/face-reading/index',
    })
  })

  it('falls back to the static gallery order when scene_order is empty', () => {
    expect(HOME_DEFAULT_SCENE_ORDER).toEqual([
      'portrait',
      'festival',
      'invitation',
      'tshirt',
      'poster',
    ])
    expect(buildHomeEntryCards([]).map((card) => card.key)).toEqual([
      'portrait',
      'face-reading',
      'festival',
      'invitation',
      'tshirt',
      'poster',
    ])
  })
})
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd frontend && npm test -- src/pages/home/entry-cards.test.ts --run`
Expected: FAIL，提示 `Failed to resolve import "./entry-cards"` 或 `buildHomeEntryCards is not exported`

- [ ] **Step 3: 写最小实现，生成统一入口卡序列**

```ts
// frontend/src/pages/home/entry-cards.ts
import { buildScenePresentation, type ScenePresentation } from '../../utils/scene'

export const HOME_DEFAULT_SCENE_ORDER = [
  'portrait',
  'festival',
  'invitation',
  'tshirt',
  'poster',
] as const

export type HomeEntryAccent = ScenePresentation['accent'] | 'analysis'

export interface HomeEntryCard {
  kind: 'scene' | 'tool'
  key: string
  title: string
  description: string
  eyebrow: string
  tags: string[]
  accent: HomeEntryAccent
  path: string
  sceneKey?: string
}

const FACE_READING_ENTRY: HomeEntryCard = {
  kind: 'tool',
  key: 'face-reading',
  title: '面相分析',
  description: '上传照片，AI 解析面部特征与数字命格。',
  eyebrow: 'AI Analysis',
  tags: ['Insight', 'Portrait'],
  accent: 'analysis',
  path: '/pages/face-reading/index',
}

function mapSceneEntry(scene: ScenePresentation): HomeEntryCard {
  return {
    kind: 'scene',
    key: scene.key,
    title: scene.name,
    description: scene.description,
    eyebrow: scene.eyebrow,
    tags: scene.tags,
    accent: scene.accent,
    path: `/pages/scene/index?scene_key=${scene.key}`,
    sceneKey: scene.key,
  }
}

export function buildHomeEntryCards(sceneOrder: string[]): HomeEntryCard[] {
  const normalizedOrder = sceneOrder.length
    ? sceneOrder
    : [...HOME_DEFAULT_SCENE_ORDER]

  const sceneCards = normalizedOrder.map((sceneKey) =>
    mapSceneEntry(buildScenePresentation(sceneKey)),
  )

  const leadCard =
    sceneCards[0] ?? mapSceneEntry(buildScenePresentation('portrait'))

  return [leadCard, FACE_READING_ENTRY, ...sceneCards.slice(1)]
}
```

- [ ] **Step 4: 更新首页 view-model 测试，先让它失败**

```ts
import { describe, expect, it } from 'vitest'
import { buildHomeViewModel } from './view-model'

describe('home view model', () => {
  it('builds entry cards, recent works and credit summary', () => {
    const vm = buildHomeViewModel({
      sceneOrder: ['portrait', 'festival', 'invitation'],
      historyItems: [
        {
          id: 1,
          sceneKey: 'festival',
          templateKey: 'spring',
          status: 'success',
          resultUrl: 'https://x/1.png',
          createdAt: '1714003200',
        },
        {
          id: 2,
          sceneKey: 'portrait',
          templateKey: 'office-pro',
          status: 'failed',
          resultUrl: '',
          createdAt: '1714089600',
        },
      ],
      profile: { balance: 5, free_quota: 2 },
    })

    expect(vm.entryCards.map((card) => card.key)).toEqual([
      'portrait',
      'face-reading',
      'festival',
      'invitation',
    ])
    expect(vm.entryCards[1]?.path).toBe('/pages/face-reading/index')
    expect(vm.recentWorks.map((item) => item.id)).toEqual([1])
    expect(vm.creditTitle).toBe('剩余额度')
    expect(vm.creditValue).toBe('2')
  })
})
```

- [ ] **Step 5: 跑首页 view-model 测试确认失败**

Run: `cd frontend && npm test -- src/pages/home/view-model.test.ts --run`
Expected: FAIL，提示 `entryCards` 为 `undefined` 或断言仍然命中旧的 `heroScene/galleryScenes`

- [ ] **Step 6: 修改首页 view-model，改为输出统一入口卡**

```ts
// frontend/src/pages/home/view-model.ts
import type { UserProfile } from '../../store/user'
import { takeRecentSuccessItems } from '../../utils/generation'
import { buildHomeEntryCards } from './entry-cards'

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
    entryCards: buildHomeEntryCards(input.sceneOrder),
    recentWorks: takeRecentSuccessItems(input.historyItems, 3),
    creditTitle: '剩余额度',
    creditValue: String(input.profile?.free_quota ?? 0),
    balanceValue: String(input.profile?.balance ?? 0),
  }
}
```

- [ ] **Step 7: 跑测试确认展示模型通过**

Run: `cd frontend && npm test -- src/pages/home/entry-cards.test.ts src/pages/home/view-model.test.ts --run`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/home/entry-cards.ts frontend/src/pages/home/entry-cards.test.ts frontend/src/pages/home/view-model.ts frontend/src/pages/home/view-model.test.ts
git commit -m "feat: add home gallery entry card model"
```

### Task 2: 建立叠册卡组的纯交互状态机

**Files:**
- Create: `frontend/src/pages/home/entry-stack.ts`
- Create: `frontend/src/pages/home/entry-stack.test.ts`

- [ ] **Step 1: 先写叠册状态机的失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { buildHomeEntryCards } from './entry-cards'
import {
  STACK_PEEK_COUNT,
  STACK_SWIPE_THRESHOLD,
  clampStackIndex,
  getVisibleHomeEntries,
  resolveStackSwipe,
} from './entry-stack'

const cards = buildHomeEntryCards([
  'portrait',
  'festival',
  'invitation',
  'tshirt',
  'poster',
])

describe('home entry stack', () => {
  it('exposes one active card plus three peek cards', () => {
    expect(STACK_PEEK_COUNT).toBe(3)

    const visible = getVisibleHomeEntries(cards, 1)

    expect(visible.map((item) => `${item.slot}:${item.entry.key}`)).toEqual([
      'active:face-reading',
      'peek-1:festival',
      'peek-2:invitation',
      'peek-3:tshirt',
    ])
  })

  it('clamps indexes into a safe range', () => {
    expect(clampStackIndex(cards.length, -2)).toBe(0)
    expect(clampStackIndex(cards.length, 99)).toBe(cards.length - 1)
  })

  it('moves to the next card on a large upward swipe', () => {
    expect(STACK_SWIPE_THRESHOLD).toBe(72)
    expect(
      resolveStackSwipe({
        count: cards.length,
        activeIndex: 1,
        deltaY: -100,
      }),
    ).toEqual({ nextIndex: 2, consumed: true })
  })

  it('does not consume upward swipes past the last card', () => {
    expect(
      resolveStackSwipe({
        count: cards.length,
        activeIndex: cards.length - 1,
        deltaY: -100,
      }),
    ).toEqual({
      nextIndex: cards.length - 1,
      consumed: false,
    })
  })
})
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd frontend && npm test -- src/pages/home/entry-stack.test.ts --run`
Expected: FAIL，提示 `Failed to resolve import "./entry-stack"` 或 `getVisibleHomeEntries is not exported`

- [ ] **Step 3: 写纯函数实现，固定叠册窗口和滑动阈值**

```ts
// frontend/src/pages/home/entry-stack.ts
import type { HomeEntryCard } from './entry-cards'

export type HomeEntrySlot = 'active' | 'peek-1' | 'peek-2' | 'peek-3'

export interface VisibleHomeEntry {
  entry: HomeEntryCard
  index: number
  slot: HomeEntrySlot
}

export const STACK_PEEK_COUNT = 3
export const STACK_SWIPE_THRESHOLD = 72

export function clampStackIndex(length: number, index: number): number {
  if (length <= 0) return 0
  return Math.max(0, Math.min(index, length - 1))
}

export function getVisibleHomeEntries(
  entries: HomeEntryCard[],
  activeIndex: number,
  peekCount = STACK_PEEK_COUNT,
): VisibleHomeEntry[] {
  if (!entries.length) return []

  const start = clampStackIndex(entries.length, activeIndex)
  const visible = entries.slice(start, start + peekCount + 1)

  return visible.map((entry, offset) => ({
    entry,
    index: start + offset,
    slot: offset === 0 ? 'active' : (`peek-${offset}` as HomeEntrySlot),
  }))
}

export function resolveStackSwipe(input: {
  count: number
  activeIndex: number
  deltaY: number
  threshold?: number
}): { nextIndex: number; consumed: boolean } {
  const threshold = input.threshold ?? STACK_SWIPE_THRESHOLD
  const current = clampStackIndex(input.count, input.activeIndex)

  if (input.deltaY <= -threshold && current < input.count - 1) {
    return { nextIndex: current + 1, consumed: true }
  }

  if (input.deltaY >= threshold && current > 0) {
    return { nextIndex: current - 1, consumed: true }
  }

  return { nextIndex: current, consumed: false }
}
```

- [ ] **Step 4: 跑状态机测试确认通过**

Run: `cd frontend && npm test -- src/pages/home/entry-stack.test.ts --run`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/home/entry-stack.ts frontend/src/pages/home/entry-stack.test.ts
git commit -m "feat: add home entry stack state helpers"
```

### Task 3: 接入叠册组件并替换首页入口区

**Files:**
- Create: `frontend/src/components/home/GalleryEntryStack.vue`
- Modify: `frontend/src/pages/home/index.vue`

- [ ] **Step 1: 先把首页接线改到新组件上，让 type-check 先失败**

```vue
<!-- frontend/src/pages/home/index.vue -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import GalleryEntryStack from '../../components/home/GalleryEntryStack.vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import { getClientConfig, getHistory, getMe, mapHistoryItem } from '../../services/api'
import { useConfigStore } from '../../store/config'
import { useUserStore } from '../../store/user'
import { buildHomeViewModel } from './view-model'
import type { HomeEntryCard } from './entry-cards'

const configStore = useConfigStore()
const userStore = useUserStore()
const historyItems = ref<ReturnType<typeof mapHistoryItem>[]>([])
const loading = ref(true)
const error = ref('')

const model = computed(() =>
  buildHomeViewModel({
    sceneOrder: configStore.clientConfig?.scene_order ?? [],
    historyItems: historyItems.value,
    profile: userStore.profile,
  }),
)

async function loadPage() {
  loading.value = true
  error.value = ''
  try {
    const [configRes, historyRes, meRes] = await Promise.all([
      configStore.clientConfig ? Promise.resolve(configStore.clientConfig) : getClientConfig(),
      getHistory(),
      userStore.profile ? Promise.resolve(userStore.profile) : getMe(),
    ])

    if (!configStore.clientConfig) {
      configStore.setClientConfig(configRes)
    }
    if (!userStore.profile) {
      userStore.setProfile(meRes)
    }

    historyItems.value = (historyRes.items || []).map(mapHistoryItem)
  } catch (err) {
    error.value = '艺廊加载失败，请重试'
  } finally {
    loading.value = false
  }
}

function openEntry(entry: HomeEntryCard) {
  if (entry.kind === 'tool') {
    uni.navigateTo({ url: entry.path })
    return
  }

  uni.reLaunch({ url: entry.path })
}

function openResult(id: number) {
  uni.navigateTo({ url: `/pages/result/index?generation_id=${id}` })
}

function openProfile() {
  uni.reLaunch({ url: '/pages/profile/index' })
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
      <view class="home-page__section">
        <text class="home-page__eyebrow">Curated Collection</text>
        <GalleryEntryStack :entries="model.entryCards" @open="openEntry" />
      </view>

      <view v-if="model.recentWorks.length > 0" class="home-page__section">
        <text class="home-page__eyebrow">Recent Works</text>
        <scroll-view scroll-x class="home-page__recent-list">
          <view
            v-for="item in model.recentWorks"
            :key="item.id"
            class="home-page__recent-item"
            @tap="openResult(item.id)"
          >
            <image class="home-page__recent-image" :src="item.resultUrl" mode="aspectFill" />
          </view>
        </scroll-view>
      </view>

      <view class="home-page__credit-card" @tap="openProfile">
        <text class="home-page__eyebrow">{{ model.creditTitle }}</text>
        <text class="home-page__credit-value">{{ model.creditValue }}</text>
        <text class="home-page__credit-meta">余额 {{ model.balanceValue }}</text>
      </view>
    </view>
  </GalleryPageShell>
</template>
```

- [ ] **Step 2: 跑 type-check，确认因为新组件缺失而失败**

Run: `cd frontend && npm run type-check`
Expected: FAIL，提示 `Cannot find module '../../components/home/GalleryEntryStack.vue'`

- [ ] **Step 3: 创建叠册组件，接入测试过的状态机和高端卡面样式**

```vue
<!-- frontend/src/components/home/GalleryEntryStack.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { HomeEntryCard } from '../../pages/home/entry-cards'
import {
  clampStackIndex,
  getVisibleHomeEntries,
  resolveStackSwipe,
} from '../../pages/home/entry-stack'

const props = defineProps<{
  entries: HomeEntryCard[]
}>()

const emit = defineEmits<{
  (e: 'open', entry: HomeEntryCard): void
}>()

const activeIndex = ref(0)
const touchStartY = ref<number | null>(null)

watch(
  () => props.entries.length,
  (length) => {
    activeIndex.value = clampStackIndex(length, activeIndex.value)
  },
  { immediate: true },
)

const visibleCards = computed(() =>
  getVisibleHomeEntries(props.entries, activeIndex.value),
)

function readTouchY(event: any): number | null {
  return event?.changedTouches?.[0]?.pageY
    ?? event?.touches?.[0]?.pageY
    ?? null
}

function handleTouchStart(event: any) {
  touchStartY.value = readTouchY(event)
}

function handleTouchEnd(event: any) {
  const endY = readTouchY(event)
  if (touchStartY.value == null || endY == null) {
    touchStartY.value = null
    return
  }

  const { nextIndex } = resolveStackSwipe({
    count: props.entries.length,
    activeIndex: activeIndex.value,
    deltaY: endY - touchStartY.value,
  })

  activeIndex.value = nextIndex
  touchStartY.value = null
}

function handleTouchCancel() {
  touchStartY.value = null
}

function handleCardTap(index: number, entry: HomeEntryCard) {
  if (index === activeIndex.value) {
    emit('open', entry)
    return
  }

  activeIndex.value = clampStackIndex(props.entries.length, index)
}
</script>

<template>
  <view
    class="entry-stack"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
    @touchcancel="handleTouchCancel"
  >
    <view
      v-for="card in visibleCards"
      :key="card.entry.key"
      class="entry-stack__card"
      :class="[
        `entry-stack__card--${card.slot}`,
        `entry-stack__card--${card.entry.accent}`,
      ]"
      @tap="handleCardTap(card.index, card.entry)"
    >
      <view class="entry-stack__surface">
        <template v-if="card.slot === 'active'">
          <view class="entry-stack__active-top">
            <view>
              <text class="entry-stack__eyebrow">{{ card.entry.eyebrow }}</text>
              <text class="entry-stack__title">{{ card.entry.title }}</text>
            </view>
            <view class="entry-stack__seal" />
          </view>

          <text class="entry-stack__description">{{ card.entry.description }}</text>

          <view class="entry-stack__footer">
            <view class="entry-stack__tags">
              <text
                v-for="tag in card.entry.tags.slice(0, 3)"
                :key="tag"
                class="entry-stack__tag"
              >
                {{ tag }}
              </text>
            </view>
            <text class="entry-stack__signature">Curated Edition</text>
          </view>
        </template>

        <template v-else>
          <view class="entry-stack__peek">
            <view>
              <text class="entry-stack__peek-eyebrow">{{ card.entry.eyebrow }}</text>
              <text class="entry-stack__peek-title">{{ card.entry.title }}</text>
            </view>
            <view class="entry-stack__peek-mark" />
          </view>
        </template>
      </view>
    </view>
  </view>
</template>

<style scoped>
.entry-stack {
  position: relative;
  height: 620rpx;
}

.entry-stack__card {
  position: absolute;
  left: 0;
  right: 0;
  border-radius: 36rpx;
  overflow: hidden;
  transition: top 180ms ease, transform 180ms ease, box-shadow 180ms ease;
}

.entry-stack__card--active {
  top: 0;
  z-index: 4;
}

.entry-stack__card--peek-1 {
  top: 404rpx;
  z-index: 3;
}

.entry-stack__card--peek-2 {
  top: 472rpx;
  z-index: 2;
}

.entry-stack__card--peek-3 {
  top: 540rpx;
  z-index: 1;
}

.entry-stack__surface {
  position: relative;
  min-height: 96rpx;
  padding: 22rpx 24rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.08);
  box-shadow:
    0 24rpx 44rpx rgba(34, 25, 18, 0.16),
    0 8rpx 18rpx rgba(34, 25, 18, 0.08),
    inset 0 1rpx 0 rgba(255, 255, 255, 0.18);
}

.entry-stack__surface::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(125deg, rgba(255, 255, 255, 0.24) 8%, rgba(255, 255, 255, 0.04) 28%, rgba(255, 255, 255, 0) 52%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0));
  pointer-events: none;
}

.entry-stack__surface::after {
  content: '';
  position: absolute;
  inset: 1rpx;
  border-radius: 36rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.1);
  pointer-events: none;
}

.entry-stack__card--active .entry-stack__surface {
  min-height: 388rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.entry-stack__active-top,
.entry-stack__peek,
.entry-stack__footer,
.entry-stack__tags {
  display: flex;
  align-items: center;
}

.entry-stack__active-top,
.entry-stack__peek,
.entry-stack__footer {
  justify-content: space-between;
}

.entry-stack__tags {
  gap: 10rpx;
  flex-wrap: wrap;
}

.entry-stack__eyebrow,
.entry-stack__peek-eyebrow {
  display: block;
  font-size: 18rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  opacity: 0.72;
}

.entry-stack__title {
  display: block;
  margin-top: 10rpx;
  font-size: 52rpx;
  line-height: 1.08;
  font-weight: 700;
}

.entry-stack__description {
  max-width: 520rpx;
  font-size: 26rpx;
  line-height: 1.6;
  opacity: 0.86;
}

.entry-stack__peek-title {
  display: block;
  margin-top: 6rpx;
  font-size: 34rpx;
  font-weight: 700;
}

.entry-stack__tag {
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  font-size: 18rpx;
  background: rgba(255, 255, 255, 0.12);
  border: 1rpx solid rgba(255, 255, 255, 0.08);
}

.entry-stack__signature {
  font-size: 18rpx;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  opacity: 0.54;
}

.entry-stack__seal,
.entry-stack__peek-mark {
  flex-shrink: 0;
  border-radius: 50%;
  border: 1rpx solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.08);
}

.entry-stack__seal {
  width: 72rpx;
  height: 72rpx;
}

.entry-stack__peek-mark {
  width: 56rpx;
  height: 56rpx;
}

.entry-stack__card--portrait .entry-stack__surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 82% 20%, rgba(235, 212, 185, 0.2), transparent 18%),
    linear-gradient(160deg, #120f0d 0%, #322821 32%, #6b5a4e 100%);
}

.entry-stack__card--analysis .entry-stack__surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 85% 18%, rgba(237, 219, 194, 0.16), transparent 18%),
    linear-gradient(160deg, #181310 0%, #2c231e 34%, #796656 100%);
}

.entry-stack__card--festival .entry-stack__surface {
  color: #fbf5ef;
  background:
    radial-gradient(circle at 84% 20%, rgba(255, 227, 196, 0.16), transparent 18%),
    linear-gradient(160deg, #40241a 0%, #7c4b33 34%, #c99668 100%);
}

.entry-stack__card--invitation .entry-stack__surface {
  color: #221b17;
  background:
    radial-gradient(circle at 84% 22%, rgba(255, 255, 255, 0.5), transparent 16%),
    linear-gradient(160deg, #d8cabd 0%, #e9ddd1 52%, #f7f0e8 100%);
}

.entry-stack__card--tshirt .entry-stack__surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 84% 18%, rgba(206, 217, 229, 0.14), transparent 18%),
    linear-gradient(160deg, #111214 0%, #25303b 36%, #53616f 100%);
}

.entry-stack__card--poster .entry-stack__surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 84% 18%, rgba(226, 214, 203, 0.14), transparent 18%),
    linear-gradient(160deg, #1d1b1a 0%, #433d38 36%, #7c7167 100%);
}

.entry-stack__card--invitation .entry-stack__tag,
.entry-stack__card--invitation .entry-stack__seal,
.entry-stack__card--invitation .entry-stack__peek-mark {
  border-color: rgba(50, 38, 28, 0.12);
  background: rgba(50, 38, 28, 0.06);
}
</style>
```

- [ ] **Step 4: 跑类型检查和首页相关测试确认通过**

Run: `cd frontend && npm run type-check`
Expected: PASS

Run: `cd frontend && npm test -- src/pages/home/entry-cards.test.ts src/pages/home/entry-stack.test.ts src/pages/home/view-model.test.ts --run`
Expected: PASS

- [ ] **Step 5: 手动验证微信小程序交互**

Run: `cd frontend && npm run dev`
Expected: 开发服务启动成功，可在微信开发者工具中打开首页进行真机交互验证。

Manual checks:
- 在首页能看到一组统一叠册卡，而不是旧的 hero/tool/grid 三段结构。
- `面相分析` 是第二张卡，视觉风格和场景卡一致。
- 轻扫不足阈值时，卡片回弹不切换。
- 向上滑时会切到下一张；在最后一张继续上滑时，页面能继续正常向下浏览。
- 露边卡边缘清晰，不会被阴影糊掉；浅色邀请函卡的文字对比仍然足够。
- 如果真机上发现卡组手势持续吞掉最后一张之后的页面滚动，则本任务内立即降级为“首卡滑动切换 + 露边卡点击聚焦”，不要继续堆复杂手势拦截逻辑。

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/home/GalleryEntryStack.vue frontend/src/pages/home/index.vue
git commit -m "feat: replace home gallery cards with stacked entry rail"
```

---

## Verification Summary

在完成三个任务后，执行完整验证：

1. `cd frontend && npm test -- src/pages/home/entry-cards.test.ts src/pages/home/entry-stack.test.ts src/pages/home/view-model.test.ts --run`
Expected: PASS

2. `cd frontend && npm run type-check`
Expected: PASS

3. `cd frontend && npm run dev`
Expected: 开发服务启动，微信开发者工具中可完成叠册卡组滑动与点击验证。
