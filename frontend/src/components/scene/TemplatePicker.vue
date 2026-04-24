<script setup lang="ts">
import type { Template } from '../../types/scene'

interface Props {
  templates: Template[]
  selectedKey?: string
}

defineProps<Props>()

const emit = defineEmits<{
  (e: 'select', template: Template): void
}>()

function handleSelect(template: Template) {
  emit('select', template)
}
</script>

<template>
  <view class="template-picker">
    <text class="picker-title">选择模板</text>
    <view class="template-list">
      <view
        v-for="t in templates"
        :key="t.key"
        class="template-card"
        :class="{ active: selectedKey === t.key }"
        @click="handleSelect(t)"
      >
        <text class="template-name">{{ t.name }}</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.template-picker {
  padding: 24rpx;
}
.picker-title {
  font-size: 32rpx;
  font-weight: bold;
  margin-bottom: 16rpx;
}
.template-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}
.template-card {
  padding: 24rpx 32rpx;
  background-color: #f1f3f5;
  border-radius: 12rpx;
  border: 2rpx solid transparent;
}
.template-card.active {
  border-color: #007aff;
  background-color: #e6f2ff;
}
.template-name {
  font-size: 28rpx;
  color: #333;
}
</style>
