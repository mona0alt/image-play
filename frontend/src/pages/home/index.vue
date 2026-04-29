<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import { getClientConfig, getHistory, mapHistoryItem } from '../../services/api'
import { useConfigStore } from '../../store/config'
import type { HomeEntryCard } from './entry-cards'
import { buildHomeViewModel } from './view-model'

const configStore = useConfigStore()
const historyItems = ref<ReturnType<typeof mapHistoryItem>[]>([])
const loading = ref(true)
const error = ref('')

const model = computed(() =>
  buildHomeViewModel({
    sceneOrder: configStore.clientConfig?.scene_order ?? [],
    historyItems: historyItems.value,
  }),
)

async function loadPage() {
  loading.value = true
  error.value = ''
  try {
    const [configRes, historyRes] = await Promise.all([
      configStore.clientConfig ? Promise.resolve(configStore.clientConfig) : getClientConfig(),
      getHistory(),
    ])

    if (!configStore.clientConfig) {
      configStore.setClientConfig(configRes)
    }
    historyItems.value = (historyRes.items || []).map(mapHistoryItem)
  } catch (err) {
    error.value = '艺廊加载失败，请重试'
  } finally {
    loading.value = false
  }
}

function openEntry(entry: HomeEntryCard) {
  if (entry.kind === 'tool') {
    uni.navigateTo({ url: entry.path })
    return
  }

  uni.reLaunch({ url: entry.path })
}

function openResult(id: number) {
  uni.navigateTo({ url: `/pages/result/index?generation_id=${id}` })
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
      <view class="home-page__section">
        <text class="home-page__eyebrow">Curated Collection</text>
        <view class="home-page__entries">
          <view
            v-for="entry in model.entryCards"
            :key="entry.key"
            class="home-page__entry-card"
            :class="`home-page__entry-card--${entry.accent}`"
            @tap="openEntry(entry)"
          >
            <view class="home-page__entry-surface">
              <view class="home-page__entry-top">
                <view>
                  <text class="home-page__entry-eyebrow">{{ entry.eyebrow }}</text>
                  <text class="home-page__entry-title">{{ entry.title }}</text>
                </view>
                <view class="home-page__entry-seal" />
              </view>

              <text class="home-page__entry-description">{{ entry.description }}</text>

              <view class="home-page__entry-footer">
                <view class="home-page__entry-tags">
                  <text
                    v-for="tag in entry.tags.slice(0, 3)"
                    :key="tag"
                    class="home-page__entry-tag"
                  >
                    {{ tag }}
                  </text>
                </view>
                <text class="home-page__entry-signature">Curated Edition</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <view v-if="model.recentWorks.length > 0" class="home-page__section">
        <text class="home-page__eyebrow">Recent Works</text>
        <scroll-view scroll-x class="home-page__recent-list">
          <view
            v-for="item in model.recentWorks"
            :key="item.id"
            class="home-page__recent-item"
            @tap="openResult(item.id)"
          >
            <image class="home-page__recent-image" :src="item.resultUrl" mode="aspectFill" />
          </view>
        </scroll-view>
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

.home-page__entries {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.home-page__entry-card {
  border-radius: 36rpx;
  overflow: hidden;
}

.home-page__entry-surface {
  position: relative;
  min-height: 260rpx;
  padding: 28rpx 28rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.08);
  box-shadow:
    0 24rpx 44rpx rgba(34, 25, 18, 0.16),
    0 8rpx 18rpx rgba(34, 25, 18, 0.08),
    inset 0 1rpx 0 rgba(255, 255, 255, 0.18);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.home-page__entry-surface::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(125deg, rgba(255, 255, 255, 0.24) 8%, rgba(255, 255, 255, 0.04) 28%, rgba(255, 255, 255, 0) 52%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0));
  pointer-events: none;
}

.home-page__entry-surface::after {
  content: '';
  position: absolute;
  inset: 1rpx;
  border-radius: 36rpx;
  border: 1rpx solid rgba(255, 255, 255, 0.1);
  pointer-events: none;
}

.home-page__entry-top,
.home-page__entry-footer,
.home-page__entry-tags {
  display: flex;
  align-items: center;
}

.home-page__entry-top,
.home-page__entry-footer {
  justify-content: space-between;
}

.home-page__entry-tags {
  gap: 10rpx;
  flex-wrap: wrap;
}

.home-page__entry-eyebrow {
  display: block;
  font-size: 18rpx;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  opacity: 0.72;
}

.home-page__entry-title {
  display: block;
  margin-top: 10rpx;
  font-size: 44rpx;
  line-height: 1.08;
  font-weight: 700;
}

.home-page__entry-description {
  max-width: 520rpx;
  margin-top: 18rpx;
  font-size: 26rpx;
  line-height: 1.6;
  opacity: 0.86;
}

.home-page__entry-tag {
  padding: 8rpx 14rpx;
  border-radius: 999rpx;
  font-size: 18rpx;
  background: rgba(255, 255, 255, 0.12);
  border: 1rpx solid rgba(255, 255, 255, 0.08);
}

.home-page__entry-signature {
  font-size: 18rpx;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  opacity: 0.54;
}

.home-page__entry-seal {
  flex-shrink: 0;
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  border: 1rpx solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.08);
}

.home-page__entry-card--portrait .home-page__entry-surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 82% 20%, rgba(235, 212, 185, 0.2), transparent 18%),
    linear-gradient(160deg, #120f0d 0%, #322821 32%, #6b5a4e 100%);
}

.home-page__entry-card--analysis .home-page__entry-surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 85% 18%, rgba(237, 219, 194, 0.16), transparent 18%),
    linear-gradient(160deg, #181310 0%, #2c231e 34%, #796656 100%);
}

.home-page__entry-card--festival .home-page__entry-surface {
  color: #fbf5ef;
  background:
    radial-gradient(circle at 84% 20%, rgba(255, 227, 196, 0.16), transparent 18%),
    linear-gradient(160deg, #40241a 0%, #7c4b33 34%, #c99668 100%);
}

.home-page__entry-card--invitation .home-page__entry-surface {
  color: #221b17;
  background:
    radial-gradient(circle at 84% 22%, rgba(255, 255, 255, 0.5), transparent 16%),
    linear-gradient(160deg, #d8cabd 0%, #e9ddd1 52%, #f7f0e8 100%);
}

.home-page__entry-card--invitation .home-page__entry-tag,
.home-page__entry-card--invitation .home-page__entry-seal {
  border-color: rgba(50, 38, 28, 0.12);
  background: rgba(50, 38, 28, 0.06);
}

.home-page__entry-card--tshirt .home-page__entry-surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 84% 18%, rgba(206, 217, 229, 0.14), transparent 18%),
    linear-gradient(160deg, #111214 0%, #25303b 36%, #53616f 100%);
}

.home-page__entry-card--poster .home-page__entry-surface {
  color: #f8f3ed;
  background:
    radial-gradient(circle at 84% 18%, rgba(226, 214, 203, 0.14), transparent 18%),
    linear-gradient(160deg, #1d1b1a 0%, #433d38 36%, #7c7167 100%);
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
</style>
