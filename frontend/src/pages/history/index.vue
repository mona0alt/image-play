<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import StatusBadge from '../../components/common/StatusBadge.vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
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
  } catch (e) {
    error.value = '历史记录加载失败，请重试'
  } finally {
    loading.value = false
  }
}

function openResult(id: number) {
  uni.navigateTo({
    url: `/pages/result/index?generation_id=${id}`,
  })
}

function openCreate() {
  uni.reLaunch({ url: '/pages/scene/index' })
}

onMounted(() => {
  loadHistory()
})
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
        @action="openCreate"
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

<style scoped>
.history-page {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.history-page__filters {
  white-space: nowrap;
}

.history-page__filter {
  display: inline-flex;
  margin-right: 12rpx;
  padding: 12rpx 18rpx;
  border-radius: 999rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
  color: var(--gallery-muted);
}

.history-page__filter--active {
  background: var(--gallery-accent);
  color: #ffffff;
}

.history-page__list {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

.history-page__card {
  display: flex;
  gap: 18rpx;
  padding: 18rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.history-page__image {
  width: 180rpx;
  height: 180rpx;
  flex-shrink: 0;
  border-radius: 20rpx;
  background: var(--gallery-surface-soft);
}

.history-page__image--placeholder {
  background: linear-gradient(160deg, #ece7e5 0%, #f8f4f2 100%);
}

.history-page__info {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  justify-content: center;
}

.history-page__title {
  font-size: 30rpx;
  font-weight: 600;
}

.history-page__subtitle,
.history-page__date {
  font-size: 24rpx;
  color: var(--gallery-muted);
}
</style>
