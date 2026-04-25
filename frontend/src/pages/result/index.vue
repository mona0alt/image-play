<template>
  <view class="result-page">
    <view v-if="!currentItem" class="empty">未找到生成记录</view>
    <view v-else-if="isGenerationPending(currentItem.status)" class="loading">
      <text>生成中: {{ currentItem.status }}</text>
    </view>
    <view v-else-if="currentItem.status === 'success' && currentItem.resultUrl">
      <ResultPreviewCard
        :image-url="currentItem.resultUrl"
        @save="onSave"
        @share="onShare"
      />
    </view>
    <view v-else-if="currentItem.status === 'failed'" class="empty">
      <text>生成失败</text>
    </view>
    <view v-else class="empty">
      <text>暂无结果</text>
    </view>
    <button class="again-btn" @click="onGenerateAnother">再生成一张</button>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted, watchEffect, computed } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import ResultPreviewCard from '../../components/result/ResultPreviewCard.vue'
import { getHistory, mapHistoryItem } from '../../services/api'
import { isGenerationPending, findHistoryItemById } from '../../utils/generation'
import { mapTrackingEvent, track } from '../../services/tracking'

const generationId = ref<string>('')
const historyItems = ref<ReturnType<typeof mapHistoryItem>[]>([])

const currentItem = computed(() => {
  if (!generationId.value) return undefined
  return findHistoryItemById(historyItems.value, Number(generationId.value))
})

onLoad((query: any) => {
  generationId.value = query.generation_id || ''
})

async function refreshHistory() {
  try {
    const res = await getHistory()
    historyItems.value = (res.items || []).map(mapHistoryItem)
  } catch (e) {
    // ignore
  }
}

onMounted(() => {
  refreshHistory()
})

watchEffect((onCleanup) => {
  const item = currentItem.value
  if (item && isGenerationPending(item.status)) {
    const timer = setTimeout(() => {
      refreshHistory()
    }, 1500)
    onCleanup(() => clearTimeout(timer))
  }
})

async function onSave() {
  try {
    await track(mapTrackingEvent('save'), { generation_id: generationId.value })
  } catch (e) {
    console.error('track save failed:', e)
  }
  uni.showToast({ title: 'Saved', icon: 'success' })
}

async function onShare() {
  try {
    await track(mapTrackingEvent('share'), { generation_id: generationId.value })
  } catch (e) {
    console.error('track share failed:', e)
  }
  uni.showShareMenu({ withShareTicket: true })
}

function onGenerateAnother() {
  uni.navigateTo({ url: '/pages/scene/index' })
}
</script>

<style scoped>
.result-page {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.empty {
  text-align: center;
  color: #999;
  padding: 48px 0;
}
.loading {
  text-align: center;
  color: #666;
  padding: 48px 0;
}
.again-btn {
  width: 100%;
  height: 48px;
  line-height: 48px;
  text-align: center;
  border-radius: 8px;
  background-color: #07c160;
  color: #fff;
  font-size: 16px;
  border: none;
}
</style>
