<template>
  <view class="history-page">
    <view class="filters">
      <text
        v-for="f in filters"
        :key="f"
        class="filter-item"
        :class="{ active: filter === f }"
        @click="filter = f"
      >{{ f }}</text>
    </view>
    <view v-if="loading" class="loading">Loading...</view>
    <view v-else-if="filteredItems.length === 0" class="empty">No history</view>
    <view v-else class="list">
      <view
        v-for="item in filteredItems"
        :key="item.id"
        class="history-item"
        @click="goToResult(item)"
      >
        <image class="thumb" :src="item.resultUrl || '/static/placeholder.png'" mode="aspectFill" />
        <view class="info">
          <text class="scene">{{ item.sceneKey }}</text>
          <text class="status" :class="item.status">{{ item.status }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getHistory, mapHistoryItem } from '../../services/api'
import { filterHistoryItems } from '../../utils/generation'

interface HistoryItem {
  id: number
  sceneKey: string
  templateKey: string
  status: string
  resultUrl: string
  createdAt: string
}

const items = ref<HistoryItem[]>([])
const loading = ref(false)
const filter = ref('all')
const filters = ['all', 'queued', 'running', 'result_auditing', 'success', 'failed']

const filteredItems = computed(() => {
  return filterHistoryItems(items.value, filter.value)
})

async function fetchHistory() {
  loading.value = true
  try {
    const res = await getHistory()
    items.value = (res.items || []).map(mapHistoryItem)
  } catch (e) {
    uni.showToast({ title: 'Failed to load history', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function goToResult(item: HistoryItem) {
  uni.navigateTo({
    url: `/pages/result/index?generation_id=${item.id}`,
  })
}

onMounted(() => {
  fetchHistory()
})
</script>

<style scoped>
.history-page {
  padding: 12px;
}
.filters {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}
.filter-item {
  padding: 6px 12px;
  border-radius: 16px;
  background-color: #f0f0f0;
  font-size: 14px;
  color: #666;
}
.filter-item.active {
  background-color: #07c160;
  color: #fff;
}
.loading,
.empty {
  text-align: center;
  padding: 48px 0;
  color: #999;
}
.list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.history-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
}
.thumb {
  width: 80px;
  height: 80px;
  border-radius: 6px;
  background-color: #f5f5f5;
}
.info {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 6px;
}
.scene {
  font-size: 16px;
  font-weight: 500;
  color: #333;
}
.status {
  font-size: 13px;
  color: #999;
}
.status.success {
  color: #07c160;
}
.status.failed {
  color: #fa5151;
}
</style>
