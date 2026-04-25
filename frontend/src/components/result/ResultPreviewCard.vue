<script setup lang="ts">
const props = defineProps<{
  imageUrl: string
  title: string
  summary: string
  chips: string[]
}>()

const emit = defineEmits<{
  (e: 'save'): void
  (e: 'share'): void
}>()

function onPreview() {
  uni.previewImage({ urls: [props.imageUrl], current: props.imageUrl })
}
</script>

<template>
  <view class="result-card">
    <image class="result-card__image" :src="imageUrl" mode="aspectFit" @click="onPreview" />
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

<style scoped>
.result-card {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.result-card__image {
  width: 100%;
  min-height: 640rpx;
  border-radius: 32rpx;
  background: var(--gallery-surface);
}

.result-card__body {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.result-card__label {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.result-card__title {
  font-size: 40rpx;
  font-weight: 600;
}

.result-card__summary {
  font-size: 24rpx;
  line-height: 1.6;
  color: var(--gallery-muted);
}

.result-card__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.result-card__chip {
  padding: 10rpx 16rpx;
  border-radius: 999rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
  color: var(--gallery-muted);
  font-size: 20rpx;
}

.result-card__primary,
.result-card__secondary {
  border-radius: 999rpx;
}

.result-card__primary {
  background: var(--gallery-accent);
  color: #ffffff;
}

.result-card__secondary {
  background: transparent;
  color: var(--gallery-text);
  border: 1rpx solid var(--gallery-border);
}
</style>
