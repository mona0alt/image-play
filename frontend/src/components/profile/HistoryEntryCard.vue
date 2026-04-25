<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  thumbnails: string[]
  totalCount: number
}>()

const emit = defineEmits<{
  click: []
}>()

function handleClick() {
  emit('click')
}

const displayThumbnails = computed(() => props.thumbnails.slice(0, 3))
const remainingCount = computed(() => Math.max(0, props.totalCount - 3))
</script>

<template>
  <view class="history-card" @click="handleClick">
    <view class="history-card__header">
      <text class="history-card__title">历史档案</text>
      <text class="history-card__arrow">›</text>
    </view>
    <view class="history-card__thumbnails">
      <image
        v-for="(url, idx) in displayThumbnails"
        :key="idx"
        class="history-card__thumb"
        :src="url"
        mode="aspectFill"
      />
      <view v-if="remainingCount > 0" class="history-card__more">
        <text class="history-card__more-text">+{{ remainingCount }}</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.history-card {
  background: #ffffff;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid #eeeeee;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.history-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.history-card__title {
  font-weight: 600;
  font-size: 14px;
  color: #1c1b1b;
}

.history-card__arrow {
  font-size: 12px;
  color: #999999;
}

.history-card__thumbnails {
  display: flex;
  gap: 8px;
}

.history-card__thumb {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background: #f0f0f0;
}

.history-card__more {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.history-card__more-text {
  font-size: 12px;
  color: #999999;
}
</style>
