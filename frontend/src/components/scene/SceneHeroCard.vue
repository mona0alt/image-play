<script setup lang="ts">
import type { Scene } from '../../types/scene'

interface Props {
  scene: Scene
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
  <view class="hero-card" @click="handleTap">
    <text class="hero-icon">{{ scene.icon }}</text>
    <text class="hero-name">{{ scene.name }}</text>
    <text class="hero-desc">{{ scene.description }}</text>
  </view>
</template>

<style scoped>
.hero-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48rpx;
  margin: 24rpx;
  background-color: #f8f9fa;
  border-radius: 24rpx;
}
.hero-icon {
  font-size: 80rpx;
  margin-bottom: 16rpx;
}
.hero-name {
  font-size: 40rpx;
  font-weight: bold;
  margin-bottom: 8rpx;
}
.hero-desc {
  font-size: 28rpx;
  color: #666;
}
</style>
