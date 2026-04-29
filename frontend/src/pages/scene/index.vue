<script setup lang="ts">
import { computed, onMounted, ref, watchEffect } from 'vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import { getHistory, mapHistoryItem } from '../../services/api'
import { useGenerationStore } from '../../store/generation'
import { buildSubmitLabel } from './view-model'

const generationStore = useGenerationStore()
const historyItems = ref<ReturnType<typeof mapHistoryItem>[]>([])
const currentGenerationId = ref<number | null>(null)

const submitLabel = computed(() => buildSubmitLabel(generationStore.isSubmitting))

const currentItem = computed(() => {
  if (!currentGenerationId.value) return null
  return historyItems.value.find((item) => item.id === currentGenerationId.value)
})

const resultState = computed(() => {
  if (!currentItem.value) return 'idle'
  if (currentItem.value.status === 'success') return 'success'
  if (currentItem.value.status === 'failed') return 'failed'
  return 'pending'
})

async function handleSubmit() {
  try {
    const id = await generationStore.submitGeneration()
    currentGenerationId.value = id
    await refreshHistory()
  } catch (e: any) {
    uni.showToast({ title: e.message || '提交失败', icon: 'none' })
  }
}

async function refreshHistory() {
  try {
    const res = await getHistory()
    historyItems.value = (res.items || []).map(mapHistoryItem)
  } catch (e) {
    // ignore
  }
}

watchEffect((cleanup) => {
  if (resultState.value === 'pending') {
    const timer = setTimeout(() => {
      void refreshHistory()
    }, 1500)
    cleanup(() => clearTimeout(timer))
  }
})

onMounted(refreshHistory)
</script>

<template>
  <GalleryPageShell active-tab="create">
    <view class="scene-page">
      <view class="scene-page__header">
        <text class="scene-page__title">创作</text>
        <text class="scene-page__subtitle">开启您的 AI 视觉之旅</text>
      </view>

      <view class="scene-page__input-area">
        <textarea
          v-model="generationStore.prompt"
          class="scene-page__textarea"
          placeholder="描述您脑海中的画面..."
          :disabled="generationStore.isSubmitting"
        />
        <view class="scene-page__input-actions">
          <text class="scene-page__action-icon">&#10022;</text>
          <text class="scene-page__action-icon">&#9673;</text>
        </view>
      </view>

      <button
        class="scene-page__submit"
        :disabled="generationStore.isSubmitting || !generationStore.prompt.trim()"
        @click="handleSubmit"
      >
        {{ submitLabel }}
      </button>

      <text class="scene-page__hint">预计生成时间 30 秒 · 消耗 2 积分</text>

      <view v-if="resultState === 'pending'" class="scene-page__result">
        <text class="scene-page__status">作品正在生成中，请稍候...</text>
      </view>

      <view v-else-if="resultState === 'success' && currentItem" class="scene-page__result">
        <image
          class="scene-page__image"
          :src="currentItem.resultUrl"
          mode="widthFix"
        />
      </view>

      <view v-else-if="resultState === 'failed'" class="scene-page__result">
        <text class="scene-page__status">生成失败，请修改描述后重试</text>
      </view>
    </view>
  </GalleryPageShell>
</template>

<style scoped>
.scene-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 48rpx;
}

.scene-page__header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 48rpx;
}

.scene-page__title {
  font-size: 56rpx;
  font-weight: 700;
  color: var(--gallery-text);
}

.scene-page__subtitle {
  font-size: 26rpx;
  color: var(--gallery-muted);
}

.scene-page__input-area {
  position: relative;
  width: 100%;
  margin-bottom: 40rpx;
}

.scene-page__textarea {
  width: 100%;
  min-height: 360rpx;
  padding: 32rpx;
  padding-bottom: 72rpx;
  background: #ffffff;
  border-radius: 32rpx;
  border: 1rpx solid rgba(28, 27, 27, 0.12);
  font-size: 30rpx;
  color: var(--gallery-text);
  line-height: 1.6;
  box-shadow: 0 4rpx 20rpx rgba(0, 0, 0, 0.04);
}

.scene-page__input-actions {
  position: absolute;
  right: 24rpx;
  bottom: 24rpx;
  display: flex;
  gap: 20rpx;
}

.scene-page__action-icon {
  font-size: 32rpx;
  color: var(--gallery-muted);
  opacity: 0.6;
}

.scene-page__submit {
  width: 280rpx;
  height: 88rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--gallery-accent);
  color: #ffffff;
  border-radius: 999rpx;
  font-size: 30rpx;
  font-weight: 600;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.12);
}

.scene-page__submit[disabled] {
  opacity: 0.4;
}

.scene-page__hint {
  margin-top: 32rpx;
  font-size: 22rpx;
  color: var(--gallery-muted);
  opacity: 0.72;
}

.scene-page__result {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  margin-top: 40rpx;
}

.scene-page__status {
  font-size: 28rpx;
  color: var(--gallery-muted);
}

.scene-page__image {
  width: 100%;
  border-radius: 32rpx;
  box-shadow: 0 8rpx 24rpx rgba(0, 0, 0, 0.08);
}
</style>
