<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useGenerationStore } from '../../store/generation'
import TemplatePicker from '../../components/scene/TemplatePicker.vue'
import SceneFieldForm from '../../components/form/SceneFieldForm.vue'

const generationStore = useGenerationStore()

const sceneKey = ref('')

onMounted(() => {
  const pages = getCurrentPages()
  const page = pages[pages.length - 1] as any
  const key = page?.$route?.query?.scene_key ?? page?.options?.scene_key ?? ''
  sceneKey.value = key
  generationStore.loadTemplates(key)
})

function handleSelectTemplate(template: any) {
  generationStore.setTemplate(template)
}

function handleUpdateForm(values: Record<string, string>) {
  generationStore.setFormValues(values)
}

async function handleSubmit() {
  if (!generationStore.selectedTemplate) {
    uni.showToast({ title: '请选择模板', icon: 'none' })
    return
  }
  const selected = generationStore.selectedTemplate
  const missing = selected.formSchema
    .filter(f => f.required && !generationStore.formValues[f.name])
  if (missing.length > 0) {
    const firstMissing = missing[0]
    if (firstMissing) {
      uni.showToast({ title: `请填写: ${firstMissing.label}`, icon: 'none' })
      return
    }
  }
  try {
    const generationId = await generationStore.submitGeneration()
    uni.navigateTo({ url: `/pages/result/index?generation_id=${generationId}` })
  } catch (e: any) {
    uni.showToast({ title: e.message || '提交失败', icon: 'none' })
  }
}
</script>

<template>
  <view class="scene">
    <TemplatePicker
      :templates="generationStore.templates"
      :selected-key="generationStore.selectedTemplate?.key ?? ''"
      @select="handleSelectTemplate"
    />

    <view v-if="generationStore.selectedTemplate">
      <SceneFieldForm
        :schema="generationStore.selectedTemplate.formSchema"
        :model-value="generationStore.formValues"
        @update:model-value="handleUpdateForm"
      />

      <button class="submit-btn" :disabled="generationStore.isSubmitting" @click="handleSubmit">
        {{ generationStore.isSubmitting ? '提交中...' : '开始生成' }}
      </button>
    </view>
  </view>
</template>

<style scoped>
.scene {
  padding: 24rpx;
}
.submit-btn {
  margin: 24rpx;
  background-color: #007aff;
  color: #fff;
  border-radius: 12rpx;
}
.submit-btn[disabled] {
  opacity: 0.6;
}
</style>
