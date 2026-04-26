<script setup lang="ts">
import { ref } from 'vue'
import { updateMe } from '../../services/api'

const nickname = ref('')
const loading = ref(false)

async function handleConfirm() {
  const trimmed = nickname.value.trim()
  if (trimmed.length < 2 || trimmed.length > 20) {
    uni.showToast({ title: '昵称长度需在 2-20 个字符之间', icon: 'none' })
    return
  }

  loading.value = true
  try {
    await updateMe(trimmed)
    uni.reLaunch({ url: '/pages/home/index' })
  } catch (err) {
    console.error('[nickname] update failed:', err)
    uni.showToast({ title: '更新失败，请重试', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function handleSkip() {
  uni.reLaunch({ url: '/pages/home/index' })
}

function goBack() {
  uni.navigateBack()
}
</script>

<template>
  <view class="nickname-page">
    <view class="header">
      <text class="back-arrow" @click="goBack">&#x2190;</text>
    </view>

    <view class="content">
      <text class="title">如何称呼您？</text>
      <text class="subtitle">您可以随时在个人中心修改</text>

      <view class="input-wrapper">
        <input
          v-model="nickname"
          class="nickname-input"
          type="text"
          placeholder="请输入昵称"
          maxlength="20"
        />
        <text class="char-count">{{ nickname.length }}/20</text>
      </view>
    </view>

    <view class="actions">
      <button class="btn-confirm" :disabled="loading" @click="handleConfirm">
        {{ loading ? '保存中...' : '确认' }}
      </button>
      <text class="btn-skip" @click="handleSkip">跳过</text>
    </view>
  </view>
</template>

<style scoped>
.nickname-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 0 64rpx 48rpx;
  background: var(--gallery-bg);
}

.header {
  padding-top: 24rpx;
  height: 88rpx;
  display: flex;
  align-items: center;
}

.back-arrow {
  font-size: 40rpx;
  color: #1c1b1b;
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  margin-top: 48rpx;
}

.title {
  font-size: 48rpx;
  font-weight: 500;
  color: #1c1b1b;
}

.subtitle {
  font-size: 28rpx;
  color: #6d6865;
}

.input-wrapper {
  margin-top: 64rpx;
  position: relative;
}

.nickname-input {
  width: 100%;
  height: 96rpx;
  font-size: 36rpx;
  color: #1c1b1b;
  border-bottom: 1rpx solid #c4c7c7;
  background: transparent;
}

.nickname-input:focus {
  border-bottom-color: #1c1b1b;
}

.char-count {
  position: absolute;
  right: 0;
  bottom: -40rpx;
  font-size: 24rpx;
  color: #6d6865;
}

.actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 32rpx;
  margin-top: 48rpx;
}

.btn-confirm {
  width: 100%;
  height: 96rpx;
  background: #000000;
  color: #ffffff;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
}

.btn-confirm:active {
  transform: scale(0.98);
}

.btn-skip {
  font-size: 28rpx;
  color: #1c1b1b;
  text-decoration: underline;
  text-underline-offset: 8rpx;
}
</style>
