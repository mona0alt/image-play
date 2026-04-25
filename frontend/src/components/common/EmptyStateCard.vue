<script setup lang="ts">
defineProps<{
  title: string
  description: string
  actionLabel?: string
}>()

const emit = defineEmits<{
  (e: 'action'): void
}>()
</script>

<template>
  <view class="empty-state">
    <text class="empty-state__title">{{ title }}</text>
    <text class="empty-state__description">{{ description }}</text>
    <button v-if="actionLabel" class="empty-state__action" @click="emit('action')">{{ actionLabel }}</button>
  </view>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 40rpx 32rpx;
  border-radius: 32rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
  text-align: center;
}

.empty-state__title {
  font-size: 34rpx;
  font-weight: 600;
}

.empty-state__description {
  font-size: 24rpx;
  line-height: 1.6;
  color: var(--gallery-muted);
}

.empty-state__action {
  margin-top: 12rpx;
  background: var(--gallery-accent);
  color: #ffffff;
  border-radius: 999rpx;
}
</style>
