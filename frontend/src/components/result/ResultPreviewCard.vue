<template>
  <view class="result-preview-card">
    <image
      class="result-image"
      :src="imageUrl"
      mode="aspectFit"
      @click="onPreview"
    />
    <view class="actions">
      <button class="action-btn save" @click="onSave">Save</button>
      <button class="action-btn share" @click="onShare">Share</button>
    </view>
  </view>
</template>

<script setup lang="ts">
const props = defineProps<{
  imageUrl: string
}>()

const emit = defineEmits<{
  (e: 'save'): void
  (e: 'share'): void
}>()

function onPreview() {
  uni.previewImage({ urls: [props.imageUrl], current: props.imageUrl })
}

function onSave() {
  emit('save')
}

function onShare() {
  emit('share')
}
</script>

<style scoped>
.result-preview-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 16px;
}
.result-image {
  width: 100%;
  max-height: 60vh;
  border-radius: 12px;
  background-color: #f5f5f5;
}
.actions {
  display: flex;
  gap: 12px;
  width: 100%;
}
.action-btn {
  flex: 1;
  height: 44px;
  line-height: 44px;
  text-align: center;
  border-radius: 8px;
  font-size: 16px;
  color: #fff;
  border: none;
}
.save {
  background-color: #07c160;
}
.share {
  background-color: #576b95;
}
</style>
