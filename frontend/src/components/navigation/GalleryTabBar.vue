<script setup lang="ts">
import { PRIMARY_TABS, type PrimaryTabKey } from '../../utils/navigation'

const props = defineProps<{
  activeKey: PrimaryTabKey
}>()

function go(path: string) {
  uni.reLaunch({ url: path })
}
</script>

<template>
  <view class="tab-bar">
    <view
      v-for="tab in PRIMARY_TABS"
      :key="tab.key"
      class="tab-bar__item"
      :class="{ 'tab-bar__item--active': tab.key === props.activeKey }"
      @click="go(tab.path)"
    >
      <text class="tab-bar__label">{{ tab.label }}</text>
    </view>
  </view>
</template>

<style scoped>
.tab-bar {
  display: flex;
  gap: 12rpx;
  padding: 24rpx 24rpx calc(24rpx + env(safe-area-inset-bottom));
  background: rgba(255, 255, 255, 0.92);
  border-top: 1rpx solid var(--gallery-border);
}

.tab-bar__item {
  flex: 1;
  padding: 16rpx 0;
  border-radius: 999rpx;
  text-align: center;
}

.tab-bar__item--active {
  background: var(--gallery-accent);
}

.tab-bar__label {
  font-size: 22rpx;
  letter-spacing: 0.12em;
  color: var(--gallery-muted);
}

.tab-bar__item--active .tab-bar__label {
  color: #ffffff;
}
</style>
