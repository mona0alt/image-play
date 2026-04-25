<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import ExploreFeed from '../../components/explore/ExploreFeed.vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
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
  const item = items.value.find((i: ExploreItem) => i.id === id)
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
  <GalleryPageShell active-tab="explore" no-padding>
    <view class="explore-page">
      <view
        v-if="error || (vm.empty && !loading)"
        class="explore-page__placeholder"
      >
        <EmptyStateCard
          v-if="error"
          title="推荐内容加载失败"
          :description="error"
          action-label="重新加载"
          @action="loadFeed"
        />
        <EmptyStateCard
          v-else
          title="暂无精选作品"
          description="先去创作吧"
          action-label="去创作"
          @action="openCreate"
        />
      </view>

      <view v-else class="explore-page__feed">
        <ExploreFeed
          :items="items"
          :has-more="hasMore"
          @load-more="loadFeed"
          @like="handleLike"
          @same-style="handleSameStyle"
        />
      </view>
    </view>
  </GalleryPageShell>
</template>

<style scoped>
.explore-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #000000;
}

.explore-page__placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48rpx 32rpx;
  background: var(--gallery-bg);
}

.explore-page__feed {
  flex: 1;
  position: relative;
}
</style>
