<script setup lang="ts">
import { ref, nextTick } from 'vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import MarkdownRenderer from '../../components/common/MarkdownRenderer.vue'
import { faceReadingStream } from '../../services/api'

const imageBase64 = ref('')
const result = ref('')
const loading = ref(false)
const scanning = ref(false)
const error = ref('')

function chooseImage() {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (res: any) => {
      const tempPath = res.tempFilePaths[0] as string
      const fs = uni.getFileSystemManager()
      fs.readFile({
        filePath: tempPath,
        encoding: 'base64',
        success: (readRes: any) => {
          const ext = tempPath.split('.').pop()?.toLowerCase() || 'jpeg'
          const mime = ext === 'png' ? 'image/png' : ext === 'gif' ? 'image/gif' : ext === 'webp' ? 'image/webp' : 'image/jpeg'
          imageBase64.value = `data:${mime};base64,${readRes.data}`
          result.value = ''
          error.value = ''
          scanning.value = false
        },
        fail: () => {
          uni.showToast({ title: '读取图片失败', icon: 'none' })
        },
      })
    },
    fail: () => {
      uni.showToast({ title: '选择图片失败', icon: 'none' })
    },
  })
}

function reset() {
  imageBase64.value = ''
  result.value = ''
  error.value = ''
  scanning.value = false
}

function scrollToBottom() {
  uni.pageScrollTo({
    scrollTop: 999999,
    duration: 0,
  })
}

async function submit() {
  if (!imageBase64.value) {
    uni.showToast({ title: '请先选择一张照片', icon: 'none' })
    return
  }
  if (imageBase64.value.length > 7 * 1024 * 1024) {
    uni.showToast({ title: '图片过大，请选择较小的图片', icon: 'none' })
    return
  }

  scanning.value = true
  loading.value = true
  error.value = ''
  result.value = ''

  try {
    await faceReadingStream(imageBase64.value, (chunk) => {
      result.value += chunk
      nextTick(() => {
        scrollToBottom()
      })
    })
  } catch (e: any) {
    console.error('[face-reading] submit error:', e.message || e)
    error.value = '分析服务暂时繁忙，请稍后重试'
  } finally {
    loading.value = false
  }
}

function copyResult() {
  if (!result.value) return
  uni.setClipboardData({
    data: result.value,
    success: () => {
      uni.showToast({ title: '已复制', icon: 'success' })
    },
  })
}
</script>

<template>
  <GalleryPageShell active-tab="gallery">
    <view class="face-page">
      <!-- Header -->
      <view class="face-page__header">
        <text class="face-page__eyebrow">AI 面相分析</text>
        <text class="face-page__title">赛博相面</text>
        <text class="face-page__subtitle">上传面部影像，神经网络将解析你的数字命格</text>
      </view>

      <!-- Upload State -->
      <view
        v-if="!imageBase64"
        class="face-page__upload-card"
        @click="chooseImage"
      >
        <view class="face-page__upload-inner">
          <view class="face-page__upload-icon-wrap">
            <text class="face-page__upload-icon">&#x2295;</text>
          </view>
          <text class="face-page__upload-label">上传面部影像</text>
          <text class="face-page__upload-hint">支持拍照或从相册选择</text>
        </view>
        <view class="face-page__upload-scan" />
      </view>

      <!-- Preview / Scanning State -->
      <view v-else class="face-page__preview-card">
        <view
          class="face-page__preview-frame"
          :class="{ 'face-page__preview-frame--scanning': scanning }"
        >
          <image
            class="face-page__preview-image"
            :src="imageBase64"
            mode="aspectFill"
          />

          <!-- Scanning overlay -->
          <view v-if="scanning" class="face-page__scan-overlay">
            <view class="face-page__scan-line" />
            <view class="face-page__scan-grid" />
            <view class="face-page__scan-ring" />
            <view class="face-page__scan-ring face-page__scan-ring--delay" />
          </view>
        </view>

        <!-- Action bar -->
        <view v-if="!scanning" class="face-page__action-bar">
          <button class="face-page__btn-secondary" @click="reset">
            重新选择
          </button>
          <button class="face-page__btn-primary" @click="submit">
            开始扫描
          </button>
        </view>

        <!-- Scanning status -->
        <view v-if="scanning && !result" class="face-page__scan-status">
          <view class="face-page__scan-spinner">
            <view class="face-page__scan-spinner-ring" />
          </view>
          <text class="face-page__scan-status-label">神经网络解析中…</text>
        </view>
      </view>

      <!-- Error -->
      <view v-if="error" class="face-page__error">
        <text class="face-page__error-text">{{ error }}</text>
      </view>

      <!-- Result Panel -->
      <view v-if="result" class="face-page__result-card">
        <view class="face-page__result-header">
          <text class="face-page__result-title">解析报告</text>
          <view class="face-page__copy-btn" @click="copyResult">
            <text class="face-page__copy-icon">&#x25A2;</text>
            <text class="face-page__copy-label">复制</text>
          </view>
        </view>

        <view class="face-page__result-body">
          <MarkdownRenderer :source="result" />
          <view v-if="loading" class="face-page__cursor-wrap">
            <view class="face-page__cursor" />
          </view>
        </view>
      </view>
    </view>
  </GalleryPageShell>
</template>

<style scoped>
.face-page {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

/* -------------------------------------------------------------
   Header
   ------------------------------------------------------------- */
.face-page__header {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
  margin-bottom: 8rpx;
}

.face-page__eyebrow {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.face-page__title {
  font-size: 48rpx;
  font-weight: 600;
  line-height: 1.2;
  color: var(--gallery-text);
}

.face-page__subtitle {
  font-size: 24rpx;
  line-height: 1.6;
  color: var(--gallery-muted);
}

/* -------------------------------------------------------------
   Upload Card
   ------------------------------------------------------------- */
.face-page__upload-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  padding: 80rpx 40rpx;
  border-radius: 20rpx;
  background: var(--gallery-surface);
  border: 2rpx dashed var(--gallery-border);
  position: relative;
  overflow: hidden;
}

.face-page__upload-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16rpx;
}

.face-page__upload-icon-wrap {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  background: var(--gallery-surface-soft);
  display: flex;
  align-items: center;
  justify-content: center;
}

.face-page__upload-icon {
  font-size: 44rpx;
  color: var(--gallery-accent);
  line-height: 1;
}

.face-page__upload-label {
  font-size: 32rpx;
  font-weight: 500;
  color: var(--gallery-text);
}

.face-page__upload-hint {
  font-size: 22rpx;
  color: var(--gallery-muted);
}

.face-page__upload-scan {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 3rpx;
  background: linear-gradient(90deg, transparent, var(--gallery-accent), transparent);
  animation: scan-slide 3s linear infinite;
  opacity: 0.15;
}

/* -------------------------------------------------------------
   Preview Card
   ------------------------------------------------------------- */
.face-page__preview-card {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.face-page__preview-frame {
  position: relative;
  width: 100%;
  border-radius: 20rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
  overflow: hidden;
  transition: all 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.face-page__preview-frame--scanning {
  width: 240rpx;
  height: 240rpx;
  border-radius: 50%;
  margin: 0 auto;
  border: 2rpx solid var(--gallery-border);
  box-shadow: 0 0 30rpx rgba(0, 0, 0, 0.06);
}

.face-page__preview-image {
  width: 100%;
  height: 600rpx;
  display: block;
}

.face-page__preview-frame--scanning .face-page__preview-image {
  height: 100%;
}

/* Scanning Overlay */
.face-page__scan-overlay {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
  border-radius: inherit;
}

.face-page__scan-line {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 2rpx;
  background: linear-gradient(90deg, transparent, var(--gallery-accent), transparent);
  animation: scan-slide 2.5s linear infinite;
  opacity: 0.25;
}

.face-page__scan-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(28, 27, 27, 0.03) 1rpx, transparent 1rpx),
    linear-gradient(90deg, rgba(28, 27, 27, 0.03) 1rpx, transparent 1rpx);
  background-size: 40rpx 40rpx;
}

.face-page__scan-ring {
  position: absolute;
  top: -6rpx;
  left: -6rpx;
  right: -6rpx;
  bottom: -6rpx;
  border-radius: 50%;
  border: 2rpx solid rgba(28, 27, 27, 0.15);
  animation: scan-ring-pulse 2s ease-out infinite;
}

.face-page__scan-ring--delay {
  animation-delay: 1s;
}

/* Scanning Status */
.face-page__scan-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16rpx;
  padding: 20rpx 0;
}

.face-page__scan-spinner {
  position: relative;
  width: 56rpx;
  height: 56rpx;
}

.face-page__scan-spinner-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 3rpx solid transparent;
  border-top-color: var(--gallery-accent);
  animation: spin 1s linear infinite;
}

.face-page__scan-status-label {
  font-size: 24rpx;
  color: var(--gallery-muted);
}

/* -------------------------------------------------------------
   Action Bar
   ------------------------------------------------------------- */
.face-page__action-bar {
  display: flex;
  gap: 16rpx;
}

.face-page__btn-primary {
  flex: 1;
  height: 88rpx;
  line-height: 88rpx;
  background: var(--gallery-accent);
  color: #ffffff;
  border-radius: 12rpx;
  font-size: 28rpx;
  font-weight: 600;
  letter-spacing: 0.05em;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
}

.face-page__btn-primary[disabled] {
  opacity: 0.4;
}

.face-page__btn-secondary {
  flex: 1;
  height: 88rpx;
  line-height: 88rpx;
  background: var(--gallery-surface);
  color: var(--gallery-text);
  border-radius: 12rpx;
  font-size: 28rpx;
  font-weight: 500;
  letter-spacing: 0.05em;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1rpx solid var(--gallery-border);
}

/* -------------------------------------------------------------
   Error
   ------------------------------------------------------------- */
.face-page__error {
  padding: 24rpx;
  border-radius: 16rpx;
  background: rgba(255, 42, 109, 0.06);
  border: 1rpx solid rgba(255, 42, 109, 0.15);
}

.face-page__error-text {
  font-size: 26rpx;
  color: #d43a5c;
  text-align: center;
  display: block;
}

/* -------------------------------------------------------------
   Result Card
   ------------------------------------------------------------- */
.face-page__result-card {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  animation: slideUp 0.5s ease-out;
}

.face-page__result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4rpx;
}

.face-page__result-title {
  font-size: 32rpx;
  font-weight: 600;
  color: var(--gallery-text);
}

.face-page__copy-btn {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 10rpx 20rpx;
  border-radius: 10rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.face-page__copy-icon {
  font-size: 22rpx;
  color: var(--gallery-muted);
}

.face-page__copy-label {
  font-size: 22rpx;
  color: var(--gallery-muted);
  font-weight: 500;
}

.face-page__result-body {
  padding: 28rpx;
  border-radius: 20rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.face-page__cursor-wrap {
  display: flex;
  align-items: center;
  margin-top: 8rpx;
  min-height: 1em;
}

.face-page__cursor {
  width: 4rpx;
  height: 1em;
  background: var(--gallery-accent);
  animation: blink 1s step-end infinite;
}

/* -------------------------------------------------------------
   Animations
   ------------------------------------------------------------- */
@keyframes scan-slide {
  0% {
    transform: translateY(0);
  }
  100% {
    transform: translateY(600rpx);
  }
}

@keyframes scan-ring-pulse {
  0% {
    transform: scale(1);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.5);
    opacity: 0;
  }
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes blink {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(16rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
