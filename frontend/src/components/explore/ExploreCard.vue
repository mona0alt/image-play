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

  </view>
</template>

<style scoped>
.card {
  position: relative;
  width: 100%;
  height: 100%;
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

</style>
