<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import SceneGalleryCard from '../../components/scene/SceneGalleryCard.vue'
import SceneHeroCard from '../../components/scene/SceneHeroCard.vue'
import { getClientConfig, getHistory, getMe, mapHistoryItem } from '../../services/api'
import { useConfigStore } from '../../store/config'
import { useUserStore } from '../../store/user'
import { buildHomeViewModel } from './view-model'

const configStore = useConfigStore()
const userStore = useUserStore()
const historyItems = ref<ReturnType<typeof mapHistoryItem>[]>([])
const loading = ref(true)
const error = ref('')

const model = computed(() => buildHomeViewModel({
  sceneOrder: configStore.clientConfig?.scene_order ?? [],
  historyItems: historyItems.value,
  profile: userStore.profile,
}))

async function loadPage() {
  loading.value = true
  error.value = ''
  try {
    const [configRes, historyRes, meRes] = await Promise.all([
      configStore.clientConfig ? Promise.resolve(configStore.clientConfig) : getClientConfig(),
      getHistory(),
      userStore.profile ? Promise.resolve(userStore.profile) : getMe(),
    ])

    if (!configStore.clientConfig) {
      configStore.setClientConfig(configRes)
    }
    if (!userStore.profile) {
      userStore.setProfile(meRes)
    }
    historyItems.value = (historyRes.items || []).map(mapHistoryItem)
  } catch (err) {
    error.value = '艺廊加载失败，请重试'
  } finally {
    loading.value = false
  }
}

function openCreate(sceneKey: string) {
  uni.reLaunch({ url: `/pages/scene/index?scene_key=${sceneKey}` })
}

function openResult(id: number) {
  uni.navigateTo({ url: `/pages/result/index?generation_id=${id}` })
}

function openProfile() {
  uni.reLaunch({ url: '/pages/profile/index' })
}

function openFaceReading() {
  uni.navigateTo({ url: '/pages/face-reading/index' })
}

onMounted(loadPage)
</script>

<template>
  <GalleryPageShell active-tab="gallery">
    <EmptyStateCard
      v-if="error"
      title="艺廊暂时不可用"
      :description="error"
      action-label="重新加载"
      @action="loadPage"
    />

    <view v-else-if="!loading" class="home-page">
      <SceneHeroCard :scene="model.heroScene" @tap="openCreate" />

      <view class="home-page__section">
        <text class="home-page__eyebrow">AI Tools</text>
        <view class="home-page__tool-card" @click="openFaceReading">
          <view class="home-page__tool-info">
            <text class="home-page__tool-title">面相分析</text>
            <text class="home-page__tool-desc">上传照片，AI 解析面相运势</text>
          </view>
          <text class="home-page__tool-arrow">›</text>
        </view>
      </view>

      <view class="home-page__section">
        <text class="home-page__eyebrow">Curated Collection</text>
        <view class="home-page__gallery">
          <SceneGalleryCard
            v-for="scene in model.galleryScenes"
            :key="scene.key"
            :scene="scene"
            @tap="openCreate"
          />
        </view>
      </view>

      <view v-if="model.recentWorks.length > 0" class="home-page__section">
        <text class="home-page__eyebrow">Recent Works</text>
        <scroll-view scroll-x class="home-page__recent-list">
          <view
            v-for="item in model.recentWorks"
            :key="item.id"
            class="home-page__recent-item"
            @click="openResult(item.id)"
          >
            <image class="home-page__recent-image" :src="item.resultUrl" mode="aspectFill" />
          </view>
        </scroll-view>
      </view>

      <view class="home-page__credit-card" @click="openProfile">
        <text class="home-page__eyebrow">{{ model.creditTitle }}</text>
        <text class="home-page__credit-value">{{ model.creditValue }}</text>
        <text class="home-page__credit-meta">余额 {{ model.balanceValue }}</text>
      </view>
    </view>
  </GalleryPageShell>
</template>

<style scoped>
.home-page {
  display: flex;
  flex-direction: column;
  gap: 32rpx;
}

.home-page__section {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

.home-page__eyebrow {
  font-size: 20rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--gallery-muted);
}

.home-page__gallery {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20rpx;
}

.home-page__recent-list {
  white-space: nowrap;
}

.home-page__recent-item {
  display: inline-flex;
  width: 180rpx;
  height: 180rpx;
  margin-right: 16rpx;
  border-radius: 24rpx;
  overflow: hidden;
  background: var(--gallery-surface);
}

.home-page__recent-image {
  width: 100%;
  height: 100%;
}

.home-page__credit-card {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  padding: 32rpx 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.home-page__credit-value {
  font-size: 48rpx;
  font-weight: 700;
}

.home-page__credit-meta {
  font-size: 24rpx;
  color: var(--gallery-muted);
}

.home-page__tool-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface);
  border: 1rpx solid var(--gallery-border);
}

.home-page__tool-info {
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.home-page__tool-title {
  font-size: 32rpx;
  font-weight: 600;
}

.home-page__tool-desc {
  font-size: 24rpx;
  color: var(--gallery-muted);
}

.home-page__tool-arrow {
  font-size: 40rpx;
  color: var(--gallery-muted);
}
</style>
