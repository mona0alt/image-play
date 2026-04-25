<script setup lang="ts">
import type { Template } from '../../types/scene'
import type { ScenePresentation } from '../../utils/scene'

interface Props {
  scene: ScenePresentation
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
    <text class="template-picker__eyebrow">{{ scene.eyebrow }}</text>
    <text class="template-picker__title">选择模板</text>
    <view class="template-picker__list">
      <view
        v-for="t in templates"
        :key="t.key"
        class="template-card"
        :class="{ 'template-card--active': selectedKey === t.key }"
        @click="handleSelect(t)"
      >
        <image v-if="t.sampleImageUrl" class="template-card__image" :src="t.sampleImageUrl" mode="aspectFill" />
        <view v-else class="template-card__image template-card__image--fallback">
          <text class="template-card__icon">{{ scene.icon }}</text>
        </view>
        <text class="template-card__name">{{ t.name }}</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.template-picker {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.template-picker__eyebrow {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.template-picker__title {
  font-size: 40rpx;
  font-weight: 600;
}

.template-picker__list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.template-card {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  padding: 16rpx;
  background: var(--gallery-surface);
  border-radius: 24rpx;
  border: 1rpx solid transparent;
}

.template-card--active {
  border-color: var(--gallery-text);
}

.template-card__image {
  width: 100%;
  height: 220rpx;
  border-radius: 18rpx;
  background: var(--gallery-surface-soft);
}

.template-card__image--fallback {
  display: flex;
  align-items: center;
  justify-content: center;
}

.template-card__icon {
  font-size: 42rpx;
  color: var(--gallery-muted);
}

.template-card__name {
  font-size: 26rpx;
  color: var(--gallery-text);
}
</style>
