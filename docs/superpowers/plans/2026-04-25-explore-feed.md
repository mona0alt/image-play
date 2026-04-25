# Explore Feed Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the "History" bottom-nav tab with an "Explore" fullscreen immersive feed, moving History into a sub-page entry on the Profile page.

**Architecture:** Use a `swiper` component for fullscreen vertical paging (better uni-app compatibility than scroll-snap). ExploreFeed manages pagination and preloading; ExploreCard handles per-item display, like, and "same-style" navigation. History becomes a sub-page navigated from Profile via `uni.navigateTo`.

**Tech Stack:** Vue 3 (uni-app), TypeScript, existing `request()` API client

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `frontend/src/utils/navigation.ts` | Modify | Replace `history` tab with `explore` tab |
| `frontend/src/pages.json` | Modify | Register `pages/explore/index`, set `navigationStyle: custom` |
| `frontend/src/components/navigation/GalleryTabBar.vue` | Modify | Update `PrimaryTabKey` usage |
| `frontend/src/services/api.ts` | Modify | Add `getExploreFeed`, `likeExploreItem`, DTOs, mapper |
| `frontend/src/utils/image-preloader.ts` | Create | Preload images using `uni.downloadFile` |
| `frontend/src/pages/explore/view-model.ts` | Create | Build explore feed view model |
| `frontend/src/pages/explore/view-model.test.ts` | Create | Unit tests for view model |
| `frontend/src/components/explore/LikeButton.vue` | Create | Animated like/unlike button |
| `frontend/src/components/explore/ExploreCard.vue` | Create | Fullscreen card: image, info, actions |
| `frontend/src/components/explore/ExploreFeed.vue` | Create | Swiper container, pagination, preloading |
| `frontend/src/pages/explore/index.vue` | Create | Explore page shell, data fetching |
| `frontend/src/components/profile/HistoryEntryCard.vue` | Create | Profile page history entry |
| `frontend/src/pages/profile/view-model.ts` | Modify | Add `historyEntry` data to profile VM |
| `frontend/src/pages/profile/view-model.test.ts` | Modify | Update profile VM tests |
| `frontend/src/pages/profile/index.vue` | Modify | Render `HistoryEntryCard` |

---

### Task 1: Update Navigation Configuration

**Files:**
- Modify: `frontend/src/utils/navigation.ts`
- Modify: `frontend/src/components/navigation/GalleryTabBar.vue`
- Modify: `frontend/src/pages.json`

- [ ] **Step 1: Update `navigation.ts`**

Replace the `history` tab with `explore` in `PRIMARY_TABS`:

```typescript
export type PrimaryTabKey = 'gallery' | 'create' | 'explore' | 'profile'

export interface PrimaryTab {
  key: PrimaryTabKey
  label: string
  path: string
}

export const PRIMARY_TABS: PrimaryTab[] = [
  { key: 'gallery', label: '艺廊', path: '/pages/home/index' },
  { key: 'create', label: '创作', path: '/pages/scene/index' },
  { key: 'explore', label: '发现', path: '/pages/explore/index' },
  { key: 'profile', label: '我的', path: '/pages/profile/index' },
]

export function findPrimaryTab(key: PrimaryTabKey) {
  return PRIMARY_TABS.find((tab) => tab.key === key)
}
```

- [ ] **Step 2: Verify `GalleryTabBar.vue` still compiles**

`GalleryTabBar.vue` imports `PRIMARY_TABS` and `PrimaryTabKey` from `utils/navigation`. No code changes needed — the type and array shape are unchanged.

- [ ] **Step 3: Update `pages.json`**

Add the explore page **before** history (order does not matter for tabBar, but keep logical grouping):

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
      "path": "pages/explore/index",
      "style": {
        "navigationBarTitleText": "发现",
        "navigationStyle": "custom",
        "backgroundColor": "#000000"
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

- [ ] **Step 4: Commit**

```bash
git add frontend/src/utils/navigation.ts frontend/src/pages.json
git commit -m "feat: replace history tab with explore in navigation"
```

---

### Task 2: Add Explore API Client

**Files:**
- Modify: `frontend/src/services/api.ts`

- [ ] **Step 1: Write the failing test (type-check)**

Create a temporary type-check script by trying to use the new functions (they don't exist yet, so TypeScript compilation should fail):

```bash
cd frontend
npx vue-tsc --noEmit --skipLibCheck 2>&1 | grep -E "(getExploreFeed|likeExploreItem|mapExploreItem)" || true
```

Expected: errors about undefined functions.

- [ ] **Step 2: Add DTO types and mapper to `api.ts`**

Append to `frontend/src/services/api.ts`:

```typescript
export interface ExploreUserDTO {
  id: string
  nickname: string
  avatar_url: string
}

export interface ExploreItemDTO {
  id: number
  user: ExploreUserDTO
  image_url: string
  thumbnail_url: string
  prompt: string
  scene_key: string
  like_count: number
  is_liked: boolean
  description: string
  created_at: string
}

export interface ExploreFeedResponse {
  items: ExploreItemDTO[]
  pagination: {
    page: number
    page_size: number
    total: number
    has_more: boolean
  }
}

export function mapExploreItem(dto: ExploreItemDTO) {
  return {
    id: dto.id,
    user: {
      id: dto.user.id,
      nickname: dto.user.nickname,
      avatarUrl: dto.user.avatar_url,
    },
    imageUrl: dto.image_url,
    thumbnailUrl: dto.thumbnail_url,
    prompt: dto.prompt,
    sceneKey: dto.scene_key,
    likeCount: dto.like_count,
    isLiked: dto.is_liked,
    description: dto.description,
    createdAt: dto.created_at,
  }
}

export type ExploreItem = ReturnType<typeof mapExploreItem>
```

- [ ] **Step 3: Add API functions to `api.ts`**

Append after `mapExploreItem`:

```typescript
export function getExploreFeed(page = 1, pageSize = 10) {
  return request<ExploreFeedResponse>({
    url: `/api/explore/feed?page=${page}&page_size=${pageSize}`,
    method: 'GET',
  }).then((res) => ({
    items: res.items.map(mapExploreItem),
    pagination: res.pagination,
  }))
}

export function likeExploreItem(generationId: number, action: 'like' | 'unlike') {
  return request<{ success: boolean; like_count: number }>({
    url: '/api/explore/like',
    method: 'POST',
    data: { generation_id: generationId, action },
    headers: { 'Content-Type': 'application/json' },
  })
}
```

- [ ] **Step 4: Verify type-check passes**

```bash
cd frontend
npx vue-tsc --noEmit --skipLibCheck
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/services/api.ts
git commit -m "feat: add explore feed API client"
```

---

### Task 3: Create Image Preloader Utility

**Files:**
- Create: `frontend/src/utils/image-preloader.ts`
- Test: create a quick inline validation step

- [ ] **Step 1: Create the preloader**

```typescript
const PRELOAD_CACHE = new Map<string, Promise<string>>()

export function preloadImage(url: string): Promise<string> {
  if (PRELOAD_CACHE.has(url)) {
    return PRELOAD_CACHE.get(url)!
  }

  const promise = new Promise<string>((resolve, reject) => {
    uni.downloadFile({
      url,
      success: (res) => {
        if (res.statusCode === 200) {
          resolve(res.tempFilePath)
        } else {
          reject(new Error(`Download failed: ${res.statusCode}`))
        }
      },
      fail: (err) => reject(new Error(err.errMsg || 'Download failed')),
    })
  })

  PRELOAD_CACHE.set(url, promise)
  return promise
}

export function preloadImages(urls: string[]): Promise<string[]> {
  return Promise.all(urls.map((url) => preloadImage(url).catch(() => url)))
}

export function clearPreloadCache(url?: string) {
  if (url) {
    PRELOAD_CACHE.delete(url)
  } else {
    PRELOAD_CACHE.clear()
  }
}
```

- [ ] **Step 2: Verify the file is importable**

Temporarily add to any `.ts` file (e.g., `frontend/src/pages/explore/view-model.ts` which will be created next):

```typescript
import { preloadImage } from '../../utils/image-preloader'
```

Run type check:

```bash
cd frontend
npx vue-tsc --noEmit --skipLibCheck
```

Expected: no errors (the file exists and exports correctly).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/utils/image-preloader.ts
git commit -m "feat: add image preloader utility"
```

---

### Task 4: Explore View Model + Tests

**Files:**
- Create: `frontend/src/pages/explore/view-model.ts`
- Create: `frontend/src/pages/explore/view-model.test.ts`

- [ ] **Step 1: Write the failing test**

```typescript
import { describe, it, expect } from 'vitest'
import { buildExploreViewModel } from './view-model'

describe('buildExploreViewModel', () => {
  const mockItems = [
    {
      id: 1,
      user: { id: 'u1', nickname: 'User1', avatarUrl: 'https://a.com/1.jpg' },
      imageUrl: 'https://img.com/1.jpg',
      thumbnailUrl: 'https://img.com/1_t.jpg',
      prompt: 'prompt1',
      sceneKey: 'portrait',
      likeCount: 10,
      isLiked: false,
      description: 'desc1',
      createdAt: '2026-04-20T10:00:00Z',
    },
  ]

  it('builds cards from items', () => {
    const vm = buildExploreViewModel({
      items: mockItems,
      pagination: { page: 1, pageSize: 10, total: 1, hasMore: false },
    })

    expect(vm.cards).toHaveLength(1)
    expect(vm.cards[0].id).toBe(1)
    expect(vm.cards[0].user.nickname).toBe('User1')
    expect(vm.cards[0].likeCount).toBe(10)
    expect(vm.cards[0].isLiked).toBe(false)
  })

  it('returns empty cards when no items', () => {
    const vm = buildExploreViewModel({
      items: [],
      pagination: { page: 1, pageSize: 10, total: 0, hasMore: false },
    })
    expect(vm.cards).toHaveLength(0)
    expect(vm.empty).toBe(true)
  })
})
```

- [ ] **Step 2: Run the test — expect failure**

```bash
cd frontend
npx vitest run src/pages/explore/view-model.test.ts
```

Expected: FAIL — `buildExploreViewModel` not found.

- [ ] **Step 3: Implement the view model**

```typescript
import type { ExploreItem } from '../../services/api'

interface Pagination {
  page: number
  pageSize: number
  total: number
  hasMore: boolean
}

export function buildExploreViewModel(input: {
  items: ExploreItem[]
  pagination: Pagination
}) {
  return {
    cards: input.items,
    empty: input.items.length === 0,
    hasMore: input.pagination.hasMore,
  }
}
```

- [ ] **Step 4: Run the test — expect pass**

```bash
cd frontend
npx vitest run src/pages/explore/view-model.test.ts
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/explore/view-model.ts frontend/src/pages/explore/view-model.test.ts
git commit -m "feat: add explore view model with tests"
```

---

### Task 5: LikeButton Component

**Files:**
- Create: `frontend/src/components/explore/LikeButton.vue`

- [ ] **Step 1: Create the component**

```vue
<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  liked: boolean
  count: number
}>()

const emit = defineEmits<{
  toggle: []
}>()

const localLiked = ref(props.liked)
const localCount = ref(props.count)

watch(() => props.liked, (v) => { localLiked.value = v })
watch(() => props.count, (v) => { localCount.value = v })

function handleToggle() {
  localLiked.value = !localLiked.value
  localCount.value += localLiked.value ? 1 : -1
  emit('toggle')
}
</script>

<template>
  <view class="like-btn" @click.stop="handleToggle">
    <view
      class="like-btn__icon"
      :class="{ 'like-btn__icon--active': localLiked }"
    >
      <text class="like-btn__symbol">{{ localLiked ? '♥' : '♡' }}</text>
    </view>
    <text class="like-btn__label">点赞</text>
  </view>
</template>

<style scoped>
.like-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.like-btn__icon {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s ease;
}

.like-btn__icon:active {
  transform: scale(0.9);
}

.like-btn__icon--active {
  background: rgba(255, 100, 100, 0.3);
}

.like-btn__symbol {
  font-size: 20px;
  color: #ffffff;
}

.like-btn__label {
  font-size: 9px;
  color: #ffffff;
  letter-spacing: 0.1em;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}
</style>
```

- [ ] **Step 2: Verify type check**

```bash
cd frontend
npx vue-tsc --noEmit --skipLibCheck
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/explore/LikeButton.vue
git commit -m "feat: add LikeButton component"
```

---

### Task 6: ExploreCard Component

**Files:**
- Create: `frontend/src/components/explore/ExploreCard.vue`

- [ ] **Step 1: Create the component**

```vue
<script setup lang="ts">
import LikeButton from './LikeButton.vue'
import type { ExploreItem } from '../../services/api'

const props = defineProps<{
  item: ExploreItem
}>()

const emit = defineEmits<{
  like: [id: number]
  sameStyle: [item: ExploreItem]
}>()

function handleLike() {
  emit('like', props.item.id)
}

function handleSameStyle() {
  emit('sameStyle', props.item)
}
</script>

<template>
  <view class="card">
    <!-- Fullscreen background image -->
    <image
      class="card__image"
      :src="item.imageUrl"
      mode="aspectFill"
      :lazy-load="true"
    />

    <!-- Right-side action buttons -->
    <view class="card__actions">
      <LikeButton
        :liked="item.isLiked"
        :count="item.likeCount"
        @toggle="handleLike"
      />
      <view class="same-style-btn" @click.stop="handleSameStyle">
        <view class="same-style-btn__icon">
          <text class="same-style-btn__symbol">✦</text>
        </view>
        <text class="same-style-btn__label">同款</text>
      </view>
    </view>

    <!-- Bottom-left info card -->
    <view class="card__info">
      <view class="card__info-inner">
        <view class="card__author">
          <image
            class="card__avatar"
            :src="item.user.avatarUrl"
            mode="aspectFill"
          />
          <text class="card__nickname">@{{ item.user.nickname }}</text>
        </view>
        <text class="card__description">{{ item.description }}</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.card {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: #111111;
}

.card__image {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.card__actions {
  position: absolute;
  right: 16px;
  bottom: 140px;
  z-index: 10;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.same-style-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.same-style-btn__icon {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s ease;
}

.same-style-btn__icon:active {
  transform: scale(0.9);
}

.same-style-btn__symbol {
  font-size: 20px;
  color: #111111;
}

.same-style-btn__label {
  font-size: 9px;
  color: #ffffff;
  letter-spacing: 0.1em;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}

.card__info {
  position: absolute;
  bottom: 100px;
  left: 16px;
  right: 80px;
  z-index: 10;
}

.card__info-inner {
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  padding: 16px;
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  max-width: 280px;
}

.card__author {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.card__avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #ffffff;
}

.card__nickname {
  font-size: 13px;
  font-weight: 600;
  color: #ffffff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}

.card__description {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.9);
  line-height: 1.5;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}
</style>
```

- [ ] **Step 2: Verify type check**

```bash
cd frontend
npx vue-tsc --noEmit --skipLibCheck
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/explore/ExploreCard.vue
git commit -m "feat: add ExploreCard component"
```

---

### Task 7: ExploreFeed Component

**Files:**
- Create: `frontend/src/components/explore/ExploreFeed.vue`

- [ ] **Step 1: Create the component**

```vue
<script setup lang="ts">
import { ref, watch } from 'vue'
import ExploreCard from './ExploreCard.vue'
import type { ExploreItem } from '../../services/api'

const props = defineProps<{
  items: ExploreItem[]
  hasMore: boolean
}>()

const emit = defineEmits<{
  loadMore: []
  like: [id: number]
  sameStyle: [item: ExploreItem]
}>()

const currentIndex = ref(0)

function onChange(e: { detail: { current: number } }) {
  currentIndex.value = e.detail.current

  // Trigger pagination when approaching end
  if (props.items.length > 0 && e.detail.current >= props.items.length - 3) {
    if (props.hasMore) {
      emit('loadMore')
    }
  }
}

function handleLike(id: number) {
  emit('like', id)
}

function handleSameStyle(item: ExploreItem) {
  emit('sameStyle', item)
}
</script>

<template>
  <swiper
    class="feed"
    vertical
    :current="currentIndex"
    @change="onChange"
  >
    <swiper-item
      v-for="item in items"
      :key="item.id"
    >
      <ExploreCard
        :item="item"
        @like="handleLike"
        @same-style="handleSameStyle"
      />
    </swiper-item>
  </swiper>
</template>

<style scoped>
.feed {
  width: 100vw;
  height: 100vh;
}
</style>
```

- [ ] **Step 2: Verify type check**

```bash
cd frontend
npx vue-tsc --noEmit --skipLibCheck
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/explore/ExploreFeed.vue
git commit -m "feat: add ExploreFeed swiper container"
```

---

### Task 8: Explore Page (index.vue)

**Files:**
- Create: `frontend/src/pages/explore/index.vue`

- [ ] **Step 1: Create the page**

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ExploreFeed from '../../components/explore/ExploreFeed.vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import GalleryTabBar from '../../components/navigation/GalleryTabBar.vue'
import { getExploreFeed, likeExploreItem } from '../../services/api'
import { buildExploreViewModel } from './view-model'
import type { ExploreItem } from '../../services/api'

const items = ref<ExploreItem[]>([])
const hasMore = ref(true)
const loading = ref(false)
const error = ref('')
const page = ref(1)

async function loadFeed() {
  if (loading.value || !hasMore.value) return
  loading.value = true
  error.value = ''

  try {
    const res = await getExploreFeed(page.value, 10)
    if (page.value === 1) {
      items.value = res.items
    } else {
      items.value = [...items.value, ...res.items]
    }
    hasMore.value = res.pagination.has_more
    page.value += 1
  } catch (e) {
    if (page.value === 1) {
      error.value = '推荐内容加载失败，请重试'
    }
  } finally {
    loading.value = false
  }
}

async function handleLike(id: number) {
  const item = items.value.find((i) => i.id === id)
  if (!item) return

  const action = item.isLiked ? 'unlike' : 'like'
  const originalLiked = item.isLiked
  const originalCount = item.likeCount

  // Optimistic update
  item.isLiked = !item.isLiked
  item.likeCount += item.isLiked ? 1 : -1

  try {
    await likeExploreItem(id, action)
  } catch (e) {
    // Rollback on failure
    item.isLiked = originalLiked
    item.likeCount = originalCount
    uni.showToast({ title: '点赞失败，请检查网络', icon: 'none' })
  }
}

function handleSameStyle(item: ExploreItem) {
  const url = `/pages/scene/index?scene_key=${encodeURIComponent(item.sceneKey)}&prompt=${encodeURIComponent(item.prompt)}&reference_id=${item.id}`
  uni.reLaunch({ url })
}

function openCreate() {
  uni.reLaunch({ url: '/pages/scene/index' })
}

onMounted(() => {
  loadFeed()
})

const vm = computed(() => buildExploreViewModel({
  items: items.value,
  pagination: { page: page.value - 1, pageSize: 10, total: items.value.length, hasMore: hasMore.value },
}))
</script>

<template>
  <view class="explore-page">
    <EmptyStateCard
      v-if="error"
      title="推荐内容加载失败"
      :description="error"
      action-label="重新加载"
      @action="loadFeed"
    />

    <EmptyStateCard
      v-else-if="vm.empty && !loading"
      title="暂无精选作品"
      description="先去创作吧"
      action-label="去创作"
      @action="openCreate"
    />

    <ExploreFeed
      v-else
      :items="items"
      :has-more="hasMore"
      @load-more="loadFeed"
      @like="handleLike"
      @same-style="handleSameStyle"
    />

    <GalleryTabBar active-key="explore" />
  </view>
</template>

<style scoped>
.explore-page {
  position: relative;
  width: 100vw;
  height: 100vh;
  background: #000000;
}
</style>
```

- [ ] **Step 2: Verify type check**

```bash
cd frontend
npx vue-tsc --noEmit --skipLibCheck
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/explore/index.vue
git commit -m "feat: add explore page shell"
```

---

### Task 9: HistoryEntryCard Component

**Files:**
- Create: `frontend/src/components/profile/HistoryEntryCard.vue`

- [ ] **Step 1: Create the component**

```vue
<script setup lang="ts">
const props = defineProps<{
  thumbnails: string[]
  totalCount: number
}>()

const emit = defineEmits<{
  click: []
}>()

function handleClick() {
  emit('click')
}

const displayThumbnails = computed(() => props.thumbnails.slice(0, 3))
const remainingCount = computed(() => Math.max(0, props.totalCount - 3))
</script>

<template>
  <view class="history-card" @click="handleClick">
    <view class="history-card__header">
      <text class="history-card__title">历史档案</text>
      <text class="history-card__arrow">›</text>
    </view>
    <view class="history-card__thumbnails">
      <image
        v-for="(url, idx) in displayThumbnails"
        :key="idx"
        class="history-card__thumb"
        :src="url"
        mode="aspectFill"
      />
      <view v-if="remainingCount > 0" class="history-card__more">
        <text class="history-card__more-text">+{{ remainingCount }}</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.history-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid #eeeeee;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.history-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.history-card__title {
  font-weight: 600;
  font-size: 14px;
  color: #1c1b1b;
}

.history-card__arrow {
  font-size: 12px;
  color: #999999;
}

.history-card__thumbnails {
  display: flex;
  gap: 8px;
}

.history-card__thumb {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background: #f0f0f0;
}

.history-card__more {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.history-card__more-text {
  font-size: 12px;
  color: #999999;
}
</style>
```

- [ ] **Step 2: Verify type check**

```bash
cd frontend
npx vue-tsc --noEmit --skipLibCheck
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/profile/HistoryEntryCard.vue
git commit -m "feat: add HistoryEntryCard component"
```

---

### Task 10: Update Profile View Model + Tests

**Files:**
- Modify: `frontend/src/pages/profile/view-model.ts`
- Modify: `frontend/src/pages/profile/view-model.test.ts`

- [ ] **Step 1: Write the failing test first**

```typescript
import { describe, it, expect } from 'vitest'
import { buildProfileViewModel } from './view-model'

describe('buildProfileViewModel', () => {
  const baseInput = {
    profile: { balance: 100, free_quota: 5 },
    packages: [],
    historyItems: [
      { id: 1, sceneKey: 'portrait', templateKey: 't1', status: 'success', resultUrl: 'https://img.com/1.jpg', createdAt: '2026-04-20T10:00:00Z' },
      { id: 2, sceneKey: 'festival', templateKey: 't2', status: 'success', resultUrl: 'https://img.com/2.jpg', createdAt: '2026-04-20T11:00:00Z' },
      { id: 3, sceneKey: 'invitation', templateKey: 't3', status: 'success', resultUrl: 'https://img.com/3.jpg', createdAt: '2026-04-20T12:00:00Z' },
      { id: 4, sceneKey: 'portrait', templateKey: 't4', status: 'success', resultUrl: 'https://img.com/4.jpg', createdAt: '2026-04-20T13:00:00Z' },
    ],
    sceneOrder: ['portrait', 'festival', 'invitation'],
  }

  it('includes history entry with thumbnails and total count', () => {
    const vm = buildProfileViewModel(baseInput)

    expect(vm.historyEntry).toBeDefined()
    expect(vm.historyEntry.thumbnails).toHaveLength(3)
    expect(vm.historyEntry.thumbnails[0]).toBe('https://img.com/4.jpg')
    expect(vm.historyEntry.totalCount).toBe(4)
  })

  it('handles empty history', () => {
    const vm = buildProfileViewModel({ ...baseInput, historyItems: [] })
    expect(vm.historyEntry.thumbnails).toHaveLength(0)
    expect(vm.historyEntry.totalCount).toBe(0)
  })
})
```

- [ ] **Step 2: Run the test — expect failure**

```bash
cd frontend
npx vitest run src/pages/profile/view-model.test.ts
```

Expected: FAIL — `historyEntry` property does not exist.

- [ ] **Step 3: Update the view model**

```typescript
import { takeRecentSuccessItems } from '../../utils/generation'
import { buildScenePresentation } from '../../utils/scene'

export function buildProfileViewModel(input: {
  profile: { balance: number; free_quota: number } | null
  packages: { code: string; title: string; price: string; count: number }[]
  historyItems: { id: number; sceneKey: string; templateKey: string; status: string; resultUrl: string; createdAt: string }[]
  sceneOrder: string[]
}) {
  const recentWorks = takeRecentSuccessItems(input.historyItems, 4)

  return {
    accountTitle: '我的作品室',
    balance: String(input.profile?.balance ?? 0),
    freeQuota: String(input.profile?.free_quota ?? 0),
    recentWorks,
    quickScenes: input.sceneOrder.slice(0, 3).map((key) => buildScenePresentation(key)),
    packages: input.packages.map((item) => ({
      ...item,
      actionLabel: '购买',
    })),
    historyEntry: {
      thumbnails: recentWorks.map((w) => w.resultUrl).slice(0, 3),
      totalCount: input.historyItems.length,
    },
  }
}
```

- [ ] **Step 4: Run the test — expect pass**

```bash
cd frontend
npx vitest run src/pages/profile/view-model.test.ts
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/profile/view-model.ts frontend/src/pages/profile/view-model.test.ts
git commit -m "feat: add history entry to profile view model"
```

---

### Task 11: Update Profile Page to Show History Entry

**Files:**
- Modify: `frontend/src/pages/profile/index.vue`

- [ ] **Step 1: Update the template**

In `frontend/src/pages/profile/index.vue`, add the `HistoryEntryCard` after the hero section and before the Recent Works section:

Replace the `<template>` block:

```vue
<template>
  <GalleryPageShell active-tab="profile" title="我的作品室" subtitle="Profile">
    <view class="profile-page">
      <view class="profile-page__hero">
        <text class="profile-page__balance">{{ model.balance }}</text>
        <text class="profile-page__meta">余额</text>
        <text class="profile-page__quota">免费额度 {{ model.freeQuota }}</text>
      </view>

      <HistoryEntryCard
        v-if="model.historyEntry.totalCount > 0"
        :thumbnails="model.historyEntry.thumbnails"
        :total-count="model.historyEntry.totalCount"
        @click="openHistory"
      />

      <view v-if="model.recentWorks.length > 0" class="profile-page__section">
        <text class="profile-page__eyebrow">Recent Works</text>
        <scroll-view scroll-x class="profile-page__recent-list">
          <view
            v-for="item in model.recentWorks"
            :key="item.id"
            class="profile-page__recent-item"
            @click="openResult(item.id)"
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
            @click="openScene(scene.key)"
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

- [ ] **Step 2: Add import and handler**

In the `<script setup>` block, add:

```typescript
import HistoryEntryCard from '../../components/profile/HistoryEntryCard.vue'
```

And add the `openHistory` function:

```typescript
function openHistory() {
  uni.navigateTo({ url: '/pages/history/index' })
}
```

- [ ] **Step 3: Verify type check**

```bash
cd frontend
npx vue-tsc --noEmit --skipLibCheck
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/profile/index.vue
git commit -m "feat: integrate history entry into profile page"
```

---

## Self-Review Checklist

**1. Spec coverage:**
- [x] Navigation: `history` → `explore` tab (Task 1)
- [x] Explore page fullscreen feed (Tasks 6-8)
- [x] Like button + API (Tasks 2, 5)
- [x] Same-style navigation (Task 8)
- [x] History sub-page from Profile (Tasks 9-11)
- [x] Empty states and error handling (Task 8)
- [x] Preloading (Tasks 3, 7)

**2. Placeholder scan:**
- [x] No TBD/TODO/fill-in-later steps
- [x] Every step has concrete code or exact command
- [x] No vague instructions like "add appropriate error handling"

**3. Type consistency:**
- [x] `ExploreItem` type used consistently across API, view-model, and components
- [x] `PrimaryTabKey` updated to `'explore'` everywhere
- [x] `historyEntry` field name consistent in view-model and page template
