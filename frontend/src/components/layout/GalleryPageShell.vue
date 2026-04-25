<script setup lang="ts">
import type { PrimaryTabKey } from '../../utils/navigation'
import GalleryTabBar from '../navigation/GalleryTabBar.vue'

defineProps<{
  title?: string
  subtitle?: string
  activeTab?: PrimaryTabKey
}>()
</script>

<template>
  <view class="page-shell">
    <view class="page-shell__inner">
      <view v-if="title" class="page-shell__header">
        <text v-if="subtitle" class="page-shell__subtitle">{{ subtitle }}</text>
        <text class="page-shell__title">{{ title }}</text>
      </view>
      <slot />
    </view>
    <GalleryTabBar v-if="activeTab" :active-key="activeTab" />
  </view>
</template>

<style scoped>
.page-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--gallery-bg);
}

.page-shell__inner {
  flex: 1;
  padding: 32rpx 24rpx 32rpx;
}

.page-shell__header {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-bottom: 32rpx;
}

.page-shell__subtitle {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.page-shell__title {
  font-size: 48rpx;
  line-height: 1.2;
  font-weight: 600;
  color: var(--gallery-text);
}
</style>
