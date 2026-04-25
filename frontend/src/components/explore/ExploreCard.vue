<script setup lang="ts">
import LikeButton from './LikeButton.vue'
import type { ExploreItem } from '../../services/api'

const props = defineProps<{
  item: ExploreItem
}>()

const emit = defineEmits<{
  like: [id: number]
  sameStyle: [item: ExploreItem]
}>()

function handleLike() {
  emit('like', props.item.id)
}

function handleSameStyle() {
  emit('sameStyle', props.item)
}
</script>

<template>
  <view class="card">
    <!-- Fullscreen background image -->
    <image
      class="card__image"
      :src="item.imageUrl"
      mode="aspectFill"
      :lazy-load="true"
    />

    <!-- Right-side action buttons -->
    <view class="card__actions">
      <LikeButton
        :liked="item.isLiked"
        :count="item.likeCount"
        @toggle="handleLike"
      />
      <view class="same-style-btn" @click.stop="handleSameStyle">
        <view class="same-style-btn__icon">
          <text class="same-style-btn__symbol">✦</text>
        </view>
        <text class="same-style-btn__label">同款</text>
      </view>
    </view>

    <!-- Bottom-left info card -->
    <view class="card__info">
      <view class="card__info-inner">
        <view class="card__author">
          <image
            class="card__avatar"
            :src="item.user.avatarUrl"
            mode="aspectFill"
          />
          <text class="card__nickname">@{{ item.user.nickname }}</text>
        </view>
        <text class="card__description">{{ item.description }}</text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.card {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: #111111;
}

.card__image {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.card__actions {
  position: absolute;
  right: 16px;
  bottom: 140px;
  z-index: 10;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.same-style-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.same-style-btn__icon {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s ease;
}

.same-style-btn__icon:active {
  transform: scale(0.9);
}

.same-style-btn__symbol {
  font-size: 20px;
  color: #111111;
}

.same-style-btn__label {
  font-size: 9px;
  color: #ffffff;
  letter-spacing: 0.1em;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}

.card__info {
  position: absolute;
  bottom: 100px;
  left: 16px;
  right: 80px;
  z-index: 10;
}

.card__info-inner {
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  padding: 16px;
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  max-width: 280px;
}

.card__author {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.card__avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #ffffff;
}

.card__nickname {
  font-size: 13px;
  font-weight: 600;
  color: #ffffff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}

.card__description {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.9);
  line-height: 1.5;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}
</style>
