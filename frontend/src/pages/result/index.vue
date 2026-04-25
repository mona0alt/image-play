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
  try {
    const res = await getHistory()
    historyItems.value = (res.items || []).map(mapHistoryItem)
  } catch (e) {
    // ignore
  }
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
  try {
    await track(mapTrackingEvent('save'), { generation_id: generationId.value })
  } catch (e) {
    console.error('track save failed:', e)
  }
  uni.showToast({ title: '已保存', icon: 'success' })
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
  const sceneKey = model.value.currentItem?.sceneKey ?? 'portrait'
  uni.reLaunch({ url: `/pages/scene/index?scene_key=${sceneKey}` })
}

function openResult(id: number) {
  uni.navigateTo({ url: `/pages/result/index?generation_id=${id}` })
}

onMounted(refreshHistory)
</script>

<template>
  <GalleryPageShell :title="model.title" subtitle="Generated Artwork" active-tab="gallery">
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

    <view v-else-if="model.state === 'success' && model.currentItem" class="result-page">
      <ResultPreviewCard
        :image-url="model.currentItem.resultUrl"
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

      <view v-if="model.recommendations.length > 0" class="result-page__section">
        <text class="result-page__eyebrow">Recent Works</text>
        <scroll-view scroll-x class="result-page__recent-list">
          <view
            v-for="item in model.recommendations"
            :key="item.id"
            class="result-page__recent-item"
            @click="openResult(item.id)"
          >
            <image class="result-page__recent-image" :src="item.resultUrl" mode="aspectFill" />
          </view>
        </scroll-view>
      </view>
    </view>
  </GalleryPageShell>
</template>

<style scoped>
.result-page {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.result-page__state {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  padding: 48rpx 32rpx;
  border-radius: 32rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.result-page__pending-title {
  font-size: 38rpx;
  font-weight: 600;
}

.result-page__pending-desc,
.result-page__eyebrow {
  font-size: 24rpx;
  color: var(--gallery-muted);
}

.result-page__eyebrow {
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.result-page__section {
  display: flex;
  flex-direction: column;
  gap: 14rpx;
}

.result-page__again {
  background: var(--gallery-accent);
  color: #ffffff;
  border-radius: 999rpx;
}

.result-page__recent-list {
  white-space: nowrap;
}

.result-page__recent-item {
  display: inline-flex;
  width: 180rpx;
  height: 180rpx;
  margin-right: 16rpx;
  border-radius: 24rpx;
  overflow: hidden;
}

.result-page__recent-image {
  width: 100%;
  height: 100%;
}
</style>
