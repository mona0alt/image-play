<script setup lang="ts">
import { ref, onMounted } from 'vue'
const error = ref('')
import { useConfigStore } from '../../store/config'
import { getHero, getGallery } from '../../utils/scene'
import SceneHeroCard from '../../components/scene/SceneHeroCard.vue'
import SceneGalleryCard from '../../components/scene/SceneGalleryCard.vue'
import { getClientConfig } from '../../services/api'

const configStore = useConfigStore()

const sceneOrder = ref<string[]>([])

const sceneMeta: Record<string, { name: string; description: string; icon: string }> = {
  portrait: { name: '人像写真', description: '打造专属个人写真风格', icon: '📸' },
  festival: { name: '节日海报', description: '快速生成节日主题海报', icon: '🎉' },
  invitation: { name: '邀请函', description: '精美电子邀请函制作', icon: '💌' },
  tshirt: { name: 'T恤图案', description: '创意T恤图案设计', icon: '👕' },
  poster: { name: '商业海报', description: '专业商业宣传海报', icon: '📰' },
}

function buildScene(key: string) {
  const meta = sceneMeta[key]
  if (!meta) return { key, name: key, description: '', icon: '❓' }
  return { key, name: meta.name, description: meta.description, icon: meta.icon }
}

onMounted(async () => {
  if (!configStore.clientConfig) {
    try {
      const config = await getClientConfig()
      configStore.setClientConfig(config)
    } catch (e) {
      error.value = '配置加载失败，请重试'
    }
  }
  sceneOrder.value = configStore.clientConfig?.scene_order ?? []
})

function navigateToScene(key: string) {
  uni.navigateTo({ url: `/pages/scene/index?scene_key=${key}` })
}
</script>

<template>
  <view class="home">
    <view v-if="sceneOrder.length > 0">
      <SceneHeroCard
        v-if="getHero(sceneOrder)"
        :scene="buildScene(getHero(sceneOrder)!)"
        @tap="navigateToScene"
      />
      <view class="gallery">
        <SceneGalleryCard
          v-for="key in getGallery(sceneOrder)"
          :key="key"
          :scene="buildScene(key)"
          @tap="navigateToScene"
        />
      </view>
    </view>
    <view v-else-if="error">{{ error }}</view>
    <view v-else class="loading">
      <text>加载中...</text>
    </view>
  </view>
</template>

<style scoped>
.home {
  padding: 24rpx;
}
.gallery {
  display: flex;
  flex-wrap: wrap;
  margin-top: 16rpx;
}
.loading {
  text-align: center;
  padding: 48rpx;
  color: #999;
}
</style>
