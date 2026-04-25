<script setup lang="ts">
import type { ScenePresentation } from '../../utils/scene'

interface Props {
  scene: ScenePresentation
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
  <view class="hero-card" :class="`hero-card--${scene.accent}`" @click="handleTap">
    <text class="hero-card__eyebrow">{{ scene.eyebrow }}</text>
    <text class="hero-card__title">{{ scene.name }}</text>
    <text class="hero-card__description">{{ scene.description }}</text>
    <view class="hero-card__tags">
      <text v-for="tag in scene.tags" :key="tag" class="hero-card__tag">{{ tag }}</text>
    </view>
    <button class="hero-card__cta">进入创作</button>
  </view>
</template>

<style scoped>
.hero-card {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  padding: 40rpx 32rpx;
  border-radius: 36rpx;
  color: #ffffff;
  background: linear-gradient(155deg, #151515 0%, #7a6a61 100%);
}

.hero-card--festival {
  background: linear-gradient(155deg, #5d3426 0%, #c38f62 100%);
}

.hero-card--invitation {
  background: linear-gradient(155deg, #b9a89d 0%, #efe8e2 100%);
  color: #1c1b1b;
}

.hero-card--tshirt {
  background: linear-gradient(155deg, #111111 0%, #41515f 100%);
}

.hero-card--poster {
  background: linear-gradient(155deg, #272727 0%, #8b7d73 100%);
}

.hero-card__eyebrow {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.hero-card__title {
  font-size: 56rpx;
  line-height: 1.15;
  font-weight: 700;
}

.hero-card__description {
  font-size: 26rpx;
  line-height: 1.6;
}

.hero-card__tags {
  display: flex;
  gap: 12rpx;
  flex-wrap: wrap;
}

.hero-card__tag {
  padding: 10rpx 18rpx;
  border-radius: 999rpx;
  font-size: 20rpx;
  background: rgba(255, 255, 255, 0.14);
}

.hero-card__cta {
  margin-top: 8rpx;
  background: rgba(255, 255, 255, 0.16);
  color: inherit;
  border-radius: 999rpx;
}
</style>
