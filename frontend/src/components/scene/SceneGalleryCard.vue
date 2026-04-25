<script setup lang="ts">
import type { ScenePresentation } from '../../utils/scene'

interface Props {
  scene: ScenePresentation
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'tap', key: string): void
}>()

function handleTap() {
  emit('tap', props.scene.key)
}
</script>

<template>
  <view class="gallery-card" @click="handleTap">
    <text class="gallery-card__eyebrow">{{ scene.eyebrow }}</text>
    <text class="gallery-card__name">{{ scene.name }}</text>
    <text class="gallery-card__description">{{ scene.description }}</text>
  </view>
</template>

<style scoped>
.gallery-card {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  min-height: 260rpx;
  padding: 28rpx 24rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.gallery-card__eyebrow {
  font-size: 20rpx;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.gallery-card__name {
  font-size: 32rpx;
  font-weight: 600;
  color: var(--gallery-text);
}

.gallery-card__description {
  font-size: 24rpx;
  line-height: 1.6;
  color: var(--gallery-muted);
}
</style>
