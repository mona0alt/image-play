<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '../../store/user'
import { getMe, getPackages, createOrder, getHistory, mapHistoryItem } from '../../services/api'
import { isGenerationPending } from '../../utils/generation'

const userStore = useUserStore()
const packagesList = ref<{ code: string; title: string; price: string; count: number }[]>([])
const history = ref<ReturnType<typeof mapHistoryItem>[]>([])
const loading = ref(false)

onMounted(async () => {
  if (!userStore.profile) {
    try {
      const profile = await getMe()
      userStore.setProfile(profile)
    } catch (e) {
      // auth middleware will redirect
    }
  }
  try {
    const res = await getPackages()
    packagesList.value = res.packages || []
  } catch (e) {
    // ignore
  }
  try {
    const res = await getHistory()
    history.value = (res.items || []).map(mapHistoryItem)
  } catch (e) {
    // ignore
  }
})

async function handleBuy(pkgCode: string) {
  loading.value = true
  try {
    const res = await createOrder(pkgCode)
    uni.showModal({
      title: '模拟支付',
      content: `订单号: ${res.order_no}\n金额: ${res.amount}`,
      showCancel: false,
    })
  } catch (e: any) {
    uni.showToast({ title: e.message || '下单失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function formatDate(ts: string) {
  const d = new Date(Number(ts) * 1000)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}
</script>

<template>
  <view class="profile">
    <view class="card">
      <text class="label">余额</text>
      <text class="balance">{{ userStore.profile?.balance ?? 0 }}</text>
      <text class="label">免费额度</text>
      <text class="quota">{{ userStore.profile?.free_quota ?? 0 }}</text>
    </view>

    <view class="section">
      <text class="section-title">充值套餐</text>
      <view v-for="pkg in packagesList" :key="pkg.code" class="pkg-card">
        <view class="pkg-info">
          <text class="pkg-title">{{ pkg.title }}</text>
          <text class="pkg-price">¥{{ pkg.price }}</text>
          <text class="pkg-count">{{ pkg.count }} 次</text>
        </view>
        <button class="buy-btn" :disabled="loading" @click="handleBuy(pkg.code)">购买</button>
      </view>
    </view>

    <view class="section">
      <text class="section-title">生成历史</text>
      <view v-if="history.length === 0" class="empty">暂无记录</view>
      <view v-for="item in history" :key="item.id" class="history-item">
        <text class="history-scene">{{ item.sceneKey }} / {{ item.templateKey }}</text>
        <text class="history-status" :class="{ pending: isGenerationPending(item.status) }">{{ item.status }}</text>
        <text class="history-date">{{ formatDate(item.createdAt) }}</text>
        <image v-if="item.resultUrl" class="history-img" :src="item.resultUrl" mode="aspectFill" />
      </view>
    </view>
  </view>
</template>

<style scoped>
.profile {
  padding: 24rpx;
}
.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 32rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.05);
}
.label {
  display: block;
  font-size: 26rpx;
  color: #666;
  margin-top: 12rpx;
}
.balance {
  display: block;
  font-size: 48rpx;
  font-weight: bold;
  color: #333;
}
.quota {
  display: block;
  font-size: 32rpx;
  color: #007aff;
}
.section {
  margin-bottom: 32rpx;
}
.section-title {
  display: block;
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  margin-bottom: 16rpx;
}
.pkg-card {
  background: #fff;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);
}
.pkg-title {
  display: block;
  font-size: 30rpx;
  font-weight: bold;
}
.pkg-price {
  display: block;
  font-size: 28rpx;
  color: #e64340;
}
.pkg-count {
  display: block;
  font-size: 24rpx;
  color: #999;
}
.buy-btn {
  background: #007aff;
  color: #fff;
  font-size: 28rpx;
  padding: 12rpx 24rpx;
  border-radius: 8rpx;
}
.empty {
  text-align: center;
  color: #999;
  padding: 48rpx;
}
.history-item {
  background: #fff;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.04);
}
.history-scene {
  display: block;
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
}
.history-status {
  display: block;
  font-size: 26rpx;
  color: #666;
  margin-top: 8rpx;
}
.history-status.pending {
  color: #007aff;
}
.history-date {
  display: block;
  font-size: 24rpx;
  color: #999;
  margin-top: 4rpx;
}
.history-img {
  width: 200rpx;
  height: 200rpx;
  margin-top: 12rpx;
  border-radius: 8rpx;
}
</style>
