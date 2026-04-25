<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  liked: boolean
  count: number
}>()

const emit = defineEmits<{
  toggle: []
}>()

const localLiked = ref(props.liked)
const localCount = ref(props.count)

watch(() => props.liked, (v) => { localLiked.value = v })
watch(() => props.count, (v) => { localCount.value = v })

function handleToggle() {
  localLiked.value = !localLiked.value
  localCount.value += localLiked.value ? 1 : -1
  emit('toggle')
}
</script>

<template>
  <view class="like-btn" @click.stop="handleToggle">
    <view
      class="like-btn__icon"
      :class="{ 'like-btn__icon--active': localLiked }"
    >
      <text class="like-btn__symbol">{{ localLiked ? '♥' : '♡' }}</text>
    </view>
    <text class="like-btn__label">点赞</text>
  </view>
</template>

<style scoped>
.like-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.like-btn__icon {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s ease;
}

.like-btn__icon:active {
  transform: scale(0.9);
}

.like-btn__icon--active {
  background: rgba(255, 100, 100, 0.3);
}

.like-btn__symbol {
  font-size: 20px;
  color: #ffffff;
}

.like-btn__label {
  font-size: 9px;
  color: #ffffff;
  letter-spacing: 0.1em;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
}
</style>
