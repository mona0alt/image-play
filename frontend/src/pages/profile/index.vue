<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import { createOrder, getClientConfig, getHistory, getMe, getPackages, mapHistoryItem } from '../../services/api'
import { useConfigStore } from '../../store/config'
import { useUserStore } from '../../store/user'
import { buildProfileViewModel } from './view-model'

const configStore = useConfigStore()
const userStore = useUserStore()
const packagesList = ref<{ code: string; title: string; price: string; count: number }[]>([])
const historyItems = ref<ReturnType<typeof mapHistoryItem>[]>([])
const loading = ref(false)

const model = computed(() => buildProfileViewModel({
  profile: userStore.profile,
  packages: packagesList.value,
  historyItems: historyItems.value,
  sceneOrder: configStore.clientConfig?.scene_order ?? ['portrait', 'festival', 'invitation'],
}))

async function loadProfilePage() {
  try {
    const [profile, packagesRes, historyRes] = await Promise.all([
      userStore.profile ? Promise.resolve(userStore.profile) : getMe(),
      getPackages(),
      getHistory(),
    ])
    if (!userStore.profile) {
      userStore.setProfile(profile)
    }
    if (!configStore.clientConfig) {
      configStore.setClientConfig(await getClientConfig())
    }
    packagesList.value = packagesRes.packages || []
    historyItems.value = (historyRes.items || []).map(mapHistoryItem)
  } catch (e) {
    uni.showToast({ title: '个人页加载失败', icon: 'none' })
  }
}

async function handleBuy(packageCode: string) {
  loading.value = true
  try {
    const order = await createOrder(packageCode)
    uni.showModal({
      title: '模拟支付',
      content: `订单号: ${order.order_no}\n金额: ${order.amount}`,
      showCancel: false,
    })
  } catch (e: any) {
    uni.showToast({ title: e.message || '下单失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function openResult(id: number) {
  uni.navigateTo({ url: `/pages/result/index?generation_id=${id}` })
}

function openScene(sceneKey: string) {
  uni.reLaunch({ url: `/pages/scene/index?scene_key=${sceneKey}` })
}

onMounted(loadProfilePage)
</script>

<template>
  <GalleryPageShell active-tab="profile" title="我的作品室" subtitle="Profile">
    <view class="profile-page">
      <view class="profile-page__hero">
        <text class="profile-page__balance">{{ model.balance }}</text>
        <text class="profile-page__meta">余额</text>
        <text class="profile-page__quota">免费额度 {{ model.freeQuota }}</text>
      </view>

      <view v-if="model.recentWorks.length > 0" class="profile-page__section">
        <text class="profile-page__eyebrow">Recent Works</text>
        <scroll-view scroll-x class="profile-page__recent-list">
          <view
            v-for="item in model.recentWorks"
            :key="item.id"
            class="profile-page__recent-item"
            @click="openResult(item.id)"
          >
            <image class="profile-page__recent-image" :src="item.resultUrl" mode="aspectFill" />
          </view>
        </scroll-view>
      </view>

      <view class="profile-page__section">
        <text class="profile-page__eyebrow">Quick Scenes</text>
        <view class="profile-page__quick-scenes">
          <view
            v-for="scene in model.quickScenes"
            :key="scene.key"
            class="profile-page__quick-scene"
            @click="openScene(scene.key)"
          >
            <text>{{ scene.name }}</text>
          </view>
        </view>
      </view>

      <view class="profile-page__section">
        <text class="profile-page__eyebrow">Packages</text>
        <view
          v-for="pkg in model.packages"
          :key="pkg.code"
          class="profile-page__package"
        >
          <view>
            <text class="profile-page__package-title">{{ pkg.title }}</text>
            <text class="profile-page__package-meta">¥{{ pkg.price }} / {{ pkg.count }} 次</text>
          </view>
          <button :disabled="loading" @click="handleBuy(pkg.code)">{{ pkg.actionLabel }}</button>
        </view>
      </view>
    </view>
  </GalleryPageShell>
</template>

<style scoped>
.profile-page {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.profile-page__hero {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  padding: 36rpx 28rpx;
  border-radius: 32rpx;
  background: linear-gradient(155deg, #171717 0%, #5b514b 100%);
  color: #ffffff;
}

.profile-page__balance {
  font-size: 56rpx;
  font-weight: 700;
}

.profile-page__meta,
.profile-page__quota {
  font-size: 24rpx;
}

.profile-page__section {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.profile-page__eyebrow {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.profile-page__recent-list {
  white-space: nowrap;
}

.profile-page__recent-item {
  display: inline-flex;
  width: 160rpx;
  height: 160rpx;
  margin-right: 16rpx;
  border-radius: 24rpx;
  overflow: hidden;
}

.profile-page__recent-image {
  width: 100%;
  height: 100%;
}

.profile-page__quick-scenes {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16rpx;
}

.profile-page__quick-scene,
.profile-page__package {
  padding: 22rpx;
  border-radius: 24rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.profile-page__quick-scene {
  text-align: center;
}

.profile-page__package {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.profile-page__package-title {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
}

.profile-page__package-meta {
  display: block;
  margin-top: 6rpx;
  font-size: 24rpx;
  color: var(--gallery-muted);
}
</style>
