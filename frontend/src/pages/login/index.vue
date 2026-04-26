<script setup lang="ts">
import { ref } from 'vue'
import { login } from '../../services/api'

const agreed = ref(false)
const loading = ref(false)

async function handleWechatLogin() {
  if (!agreed.value) {
    uni.vibrateShort()
    uni.showToast({ title: '请先同意用户协议', icon: 'none' })
    return
  }

  loading.value = true
  try {
    const loginRes = await uni.login({ provider: 'weixin' })
    if (!loginRes.code) {
      throw new Error('WeChat login failed: no code')
    }
    const res = await login(loginRes.code)
    uni.setStorageSync('access_token', res.access_token)
    uni.reLaunch({ url: '/pages/nickname-setup/index' })
  } catch (err) {
    console.error('[login] failed:', err)
    uni.showToast({ title: '微信登录失败，请重试', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function showToast(msg: string) {
  uni.showToast({ title: msg, icon: 'none' })
}
</script>

<template>
  <view class="login-page">
    <view class="brand-section">
      <view class="logo">
        <view class="logo-square"></view>
        <view class="logo-inner"></view>
        <text class="logo-icon">&#x25A1;</text>
      </view>
      <text class="brand-title">精品场景馆</text>
      <text class="brand-subtitle">开启您的艺术创作之旅</text>
    </view>

    <view class="actions">
      <button class="btn-wechat" :disabled="loading" @click="handleWechatLogin">
        <text class="wechat-icon">&#x263A;</text>
        <text>{{ loading ? '登录中...' : '微信一键登录' }}</text>
      </button>

      <view class="gallery-preview">
        <view class="preview-item">
          <view class="preview-placeholder"></view>
        </view>
        <view class="preview-item shifted">
          <view class="preview-placeholder"></view>
        </view>
        <view class="preview-item">
          <view class="preview-placeholder"></view>
        </view>
      </view>
    </view>

    <view class="footer">
      <view class="checkbox-row" @click="agreed = !agreed">
        <view class="custom-checkbox" :class="{ checked: agreed }"></view>
        <text class="terms-text">
          我已阅读并同意
          <text class="link" @click.stop="showToast('协议内容即将上线')">《用户协议》</text>
          和
          <text class="link" @click.stop="showToast('协议内容即将上线')">《隐私政策》</text>
        </text>
      </view>
    </view>
  </view>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 0 64rpx 48rpx;
  background: var(--gallery-bg);
}

.brand-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24rpx;
}

.logo {
  width: 192rpx;
  height: 192rpx;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-square {
  position: absolute;
  inset: 0;
  border: 1rpx solid #1c1b1b;
  transform: rotate(45deg);
}

.logo-inner {
  position: absolute;
  inset: 32rpx;
  border: 1rpx solid rgba(28, 27, 27, 0.2);
  transform: rotate(-12deg);
}

.logo-icon {
  font-size: 64rpx;
  color: #1c1b1b;
  z-index: 1;
}

.brand-title {
  font-size: 64rpx;
  font-weight: 600;
  color: #1c1b1b;
  letter-spacing: 0.03em;
}

.brand-subtitle {
  font-size: 24rpx;
  color: #6d6865;
  letter-spacing: 0.1em;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 48rpx;
  margin-bottom: 48rpx;
}

.btn-wechat {
  height: 112rpx;
  background: #000000;
  color: #ffffff;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  font-size: 32rpx;
  font-weight: 600;
}

.btn-wechat:active {
  transform: scale(0.98);
}

.wechat-icon {
  font-size: 40rpx;
}

.gallery-preview {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16rpx;
  opacity: 0.4;
}

.preview-item {
  padding: 8rpx;
  border: 1rpx solid #e5e2e1;
}

.preview-item.shifted {
  margin-top: 32rpx;
}

.preview-placeholder {
  height: 200rpx;
  background: #e5e2e1;
}

.footer {
  padding-top: 24rpx;
}

.checkbox-row {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
}

.custom-checkbox {
  width: 32rpx;
  height: 32rpx;
  border: 2rpx solid #1c1b1b;
  flex-shrink: 0;
  margin-top: 4rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.custom-checkbox.checked::after {
  content: '';
  width: 16rpx;
  height: 16rpx;
  background: #1c1b1b;
}

.terms-text {
  font-size: 22rpx;
  color: #6d6865;
  line-height: 1.6;
}

.link {
  color: #1c1b1b;
  text-decoration: underline;
  text-underline-offset: 8rpx;
}
</style>
