<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
