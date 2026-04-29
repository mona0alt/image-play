<script setup lang="ts">
import { ref } from 'vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import { faceReadingStream } from '../../services/api'

const imageBase64 = ref('')
const result = ref('')
const loading = ref(false)
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

  loading.value = true
  error.value = ''
  result.value = ''
  try {
    await faceReadingStream(imageBase64.value, (chunk) => {
      result.value += chunk
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
  <GalleryPageShell active-tab="gallery" title="面相分析">
    <view class="face-page">
      <view class="face-page__hero">
        <text class="face-page__hero-title">面相分析</text>
        <text class="face-page__hero-desc">上传一张正面清晰照片，AI 将基于传统面相学进行解析</text>
      </view>

      <view v-if="!imageBase64" class="face-page__upload-card" @click="chooseImage">
        <text class="face-page__upload-icon">+</text>
        <text class="face-page__upload-label">选择照片</text>
        <text class="face-page__upload-hint">支持拍照或从相册选择</text>
      </view>

      <view v-else class="face-page__preview-card">
        <image class="face-page__preview-image" :src="imageBase64" mode="aspectFit" />
        <view class="face-page__preview-actions">
          <button class="face-page__btn-secondary" @click="reset">重新选择</button>
          <button class="face-page__btn-primary" :disabled="loading" @click="submit">
            {{ loading ? '正在分析…' : '开始分析' }}
          </button>
        </view>
      </view>

      <view v-if="loading" class="face-page__loading">
        <text class="face-page__loading-text">正在分析面相，请稍候…</text>
      </view>

      <view v-if="error" class="face-page__error">
        <text>{{ error }}</text>
      </view>

      <view v-if="result" class="face-page__result-card">
        <text class="face-page__result-title">分析结果</text>
        <text class="face-page__result-body">{{ result }}</text>
        <button class="face-page__btn-primary" @click="copyResult">复制结果</button>
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

.face-page__hero {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  padding: 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface-soft);
}

.face-page__hero-title {
  font-size: 40rpx;
  font-weight: 600;
}

.face-page__hero-desc {
  font-size: 24rpx;
  line-height: 1.6;
  color: var(--gallery-muted);
}

.face-page__upload-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 80rpx 40rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 2rpx dashed var(--gallery-border);
}

.face-page__upload-icon {
  font-size: 64rpx;
  color: var(--gallery-muted);
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

.face-page__preview-card {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.face-page__preview-image {
  width: 100%;
  height: 400rpx;
  border-radius: 16rpx;
  background: var(--gallery-surface-soft);
}

.face-page__preview-actions {
  display: flex;
  gap: 16rpx;
}

.face-page__btn-primary {
  flex: 1;
  background: var(--gallery-accent);
  color: #ffffff;
  border-radius: 999rpx;
}

.face-page__btn-secondary {
  flex: 1;
  background: var(--gallery-surface-soft);
  color: var(--gallery-text);
  border-radius: 999rpx;
}

.face-page__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx;
}

.face-page__loading-text {
  font-size: 26rpx;
  color: var(--gallery-muted);
}

.face-page__error {
  padding: 24rpx;
  border-radius: 16rpx;
  background: #fff0f0;
  color: #c0392b;
  font-size: 26rpx;
  text-align: center;
}

.face-page__result-card {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  padding: 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.face-page__result-title {
  font-size: 32rpx;
  font-weight: 600;
}

.face-page__result-body {
  font-size: 26rpx;
  line-height: 1.7;
  color: var(--gallery-text);
  white-space: pre-wrap;
}
</style>
