<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import EmptyStateCard from '../../components/common/EmptyStateCard.vue'
import SceneFieldForm from '../../components/form/SceneFieldForm.vue'
import GalleryPageShell from '../../components/layout/GalleryPageShell.vue'
import TemplatePicker from '../../components/scene/TemplatePicker.vue'
import { getClientConfig } from '../../services/api'
import { useConfigStore } from '../../store/config'
import { useGenerationStore } from '../../store/generation'
import { buildScenePageModel, getSceneSubmitError, resolveSceneKey } from './view-model'

const configStore = useConfigStore()
const generationStore = useGenerationStore()
const requestedSceneKey = ref('')
const pageLoading = ref(true)
const pageError = ref('')

const currentSceneKey = computed(() =>
  resolveSceneKey(configStore.clientConfig?.scene_order ?? [], requestedSceneKey.value),
)

const model = computed(() => buildScenePageModel({
  sceneOrder: configStore.clientConfig?.scene_order ?? [],
  currentSceneKey: currentSceneKey.value,
  templates: generationStore.templates,
  selectedTemplateKey: generationStore.selectedTemplate?.key ?? '',
  isSubmitting: generationStore.isSubmitting,
}))

onLoad((query: any) => {
  requestedSceneKey.value = query.scene_key || ''
})

async function loadScene() {
  pageLoading.value = true
  pageError.value = ''
  try {
    if (!configStore.clientConfig) {
      configStore.setClientConfig(await getClientConfig())
    }
    await generationStore.loadTemplates(currentSceneKey.value)
  } catch (err) {
    pageError.value = '创作页加载失败，请重试'
  } finally {
    pageLoading.value = false
  }
}

async function switchScene(sceneKey: string) {
  requestedSceneKey.value = sceneKey
  await generationStore.loadTemplates(sceneKey)
}

async function handleSubmit() {
  const error = getSceneSubmitError(generationStore.selectedTemplate, generationStore.formValues)
  if (error) {
    uni.showToast({ title: error, icon: 'none' })
    return
  }
  try {
    const generationId = await generationStore.submitGeneration()
    uni.navigateTo({ url: `/pages/result/index?generation_id=${generationId}` })
  } catch (e: any) {
    uni.showToast({ title: e.message || '提交失败', icon: 'none' })
  }
}

onMounted(loadScene)
</script>

<template>
  <GalleryPageShell
    active-tab="create"
    :title="model.currentScene.name"
    :subtitle="model.currentScene.eyebrow"
  >
    <EmptyStateCard
      v-if="pageError"
      title="创作页暂时不可用"
      :description="pageError"
      action-label="重新加载"
      @action="loadScene"
    />

    <view v-else-if="!pageLoading" class="scene-page">
      <scroll-view scroll-x class="scene-page__scene-strip">
        <view
          v-for="scene in model.sceneChoices"
          :key="scene.key"
          class="scene-page__scene-pill"
          :class="{ 'scene-page__scene-pill--active': scene.key === currentSceneKey }"
          @click="switchScene(scene.key)"
        >
          <text>{{ scene.name }}</text>
        </view>
      </scroll-view>

      <view class="scene-page__hero">
        <text class="scene-page__hero-title">{{ model.currentScene.name }}</text>
        <text class="scene-page__hero-desc">{{ model.currentScene.description }}</text>
      </view>

      <TemplatePicker
        :scene="model.currentScene"
        :templates="generationStore.templates"
        :selected-key="generationStore.selectedTemplate?.key ?? ''"
        @select="generationStore.setTemplate"
      />

      <SceneFieldForm
        v-if="generationStore.selectedTemplate"
        :schema="generationStore.selectedTemplate.formSchema"
        :model-value="generationStore.formValues"
        @update:model-value="generationStore.setFormValues"
      />

      <button class="scene-page__submit" :disabled="generationStore.isSubmitting" @click="handleSubmit">
        {{ model.submitLabel }}
      </button>
    </view>
  </GalleryPageShell>
</template>

<style scoped>
.scene-page {
  display: flex;
  flex-direction: column;
  gap: 24rpx;
}

.scene-page__scene-strip {
  white-space: nowrap;
}

.scene-page__scene-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 12rpx;
  padding: 14rpx 22rpx;
  border-radius: 999rpx;
  background: var(--gallery-surface);
  color: var(--gallery-muted);
  border: 1rpx solid var(--gallery-border);
}

.scene-page__scene-pill--active {
  background: var(--gallery-accent);
  color: #ffffff;
}

.scene-page__hero {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
  padding: 28rpx;
  border-radius: 28rpx;
  background: var(--gallery-surface-soft);
}

.scene-page__hero-title {
  font-size: 40rpx;
  font-weight: 600;
}

.scene-page__hero-desc {
  font-size: 24rpx;
  line-height: 1.6;
  color: var(--gallery-muted);
}

.scene-page__submit {
  background: var(--gallery-accent);
  color: #ffffff;
  border-radius: 999rpx;
  margin-top: 8rpx;
}
</style>
