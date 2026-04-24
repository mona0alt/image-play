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
  <view class="gallery-card" @click="handleTap">
    <text class="gallery-icon">{{ scene.icon }}</text>
    <text class="gallery-name">{{ scene.name }}</text>
  </view>
</template>

<style scoped>
.gallery-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32rpx;
  margin: 16rpx;
  background-color: #f1f3f5;
  border-radius: 16rpx;
  flex: 1;
}
.gallery-icon {
  font-size: 56rpx;
  margin-bottom: 8rpx;
}
.gallery-name {
  font-size: 28rpx;
  color: #333;
}
</style>
