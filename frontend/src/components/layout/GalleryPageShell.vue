<script setup lang="ts">
import { computed } from 'vue'
import type { PrimaryTabKey } from '../../utils/navigation'
import GalleryTabBar from '../navigation/GalleryTabBar.vue'
import { getPageShellBottomPadding } from '../navigation/tab-bar-layout'

const props = defineProps<{
  title?: string
  subtitle?: string
  activeTab?: PrimaryTabKey
  noPadding?: boolean
}>()

const innerStyle = computed(() => {
  if (!props.activeTab) {
    return undefined
  }

  return {
    paddingBottom: getPageShellBottomPadding(Boolean(props.noPadding)),
  }
})
</script>

<template>
  <view class="page-shell">
    <view
      class="page-shell__inner"
      :class="{ 'page-shell__inner--no-padding': props.noPadding }"
      :style="innerStyle"
    >
      <view v-if="props.title" class="page-shell__header">
        <text v-if="props.subtitle" class="page-shell__subtitle">{{ props.subtitle }}</text>
        <text class="page-shell__title">{{ props.title }}</text>
      </view>
      <slot />
    </view>
    <GalleryTabBar v-if="props.activeTab" :active-key="props.activeTab" />
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
  min-height: 0;
  padding: 32rpx 24rpx 32rpx;
}

.page-shell__inner--no-padding {
  padding: 0;
  display: flex;
  flex-direction: column;
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
