<template>
  <view class="result-page">
    <ResultPreviewCard
      v-if="resultUrl"
      :image-url="resultUrl"
      @save="onSave"
      @share="onShare"
    />
    <view v-else class="empty">No result available</view>
    <button class="again-btn" @click="onGenerateAnother">Generate Another</button>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import ResultPreviewCard from '../../components/result/ResultPreviewCard.vue'
import { mapTrackingEvent, track } from '../../services/tracking'

const generationId = ref<string>('')
const resultUrl = ref<string>('')

onLoad((query: any) => {
  generationId.value = query.generation_id || ''
  resultUrl.value = query.result_url || ''
})

function onSave() {
  track(mapTrackingEvent('save'), { generation_id: generationId.value })
  uni.showToast({ title: 'Saved', icon: 'success' })
}

function onShare() {
  track(mapTrackingEvent('share'), { generation_id: generationId.value })
  uni.showShareMenu({ withShareTicket: true })
}

function onGenerateAnother() {
  uni.navigateTo({ url: '/pages/scene/index' })
}
</script>

<style scoped>
.result-page {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.empty {
  text-align: center;
  color: #999;
  padding: 48px 0;
}
.again-btn {
  width: 100%;
  height: 48px;
  line-height: 48px;
  text-align: center;
  border-radius: 8px;
  background-color: #07c160;
  color: #fff;
  font-size: 16px;
  border: none;
}
</style>
