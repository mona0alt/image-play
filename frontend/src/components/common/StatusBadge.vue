<script setup lang="ts">
defineProps<{
  label: string
  tone: 'neutral' | 'pending' | 'success' | 'danger'
}>()
</script>

<template>
  <view class="status-badge" :class="`status-badge--${tone}`">
    <text class="status-badge__text">{{ label }}</text>
  </view>
</template>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  background: #ece7e5;
}

.status-badge__text {
  font-size: 20rpx;
  letter-spacing: 0.08em;
  color: var(--gallery-muted);
}

.status-badge--pending {
  background: #ede9e0;
}

.status-badge--success {
  background: #e2ece7;
}

.status-badge--danger {
  background: #f4e2e2;
}
</style>
