<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useGenerationStore } from '../../store/generation'
import TemplatePicker from '../../components/scene/TemplatePicker.vue'
import SceneFieldForm from '../../components/form/SceneFieldForm.vue'
import type { Template } from '../../types/scene'

const generationStore = useGenerationStore()

const sceneKey = ref('')
const templates = ref<Template[]>([])
const showAdvanced = ref(false)

// TODO: replace mockTemplates with API call
const mockTemplates: Record<string, Template[]> = {
  portrait: [
    {
      key: 'portrait-cyber',
      name: '赛博朋克',
      sceneKey: 'portrait',
      formSchema: [
        { name: 'subject', label: '拍摄对象', type: 'text', required: true },
        { name: 'mood', label: '氛围', type: 'select', options: ['冷峻', '热情', '神秘'], required: false },
      ],
    },
    {
      key: 'portrait-vintage',
      name: '复古胶片',
      sceneKey: 'portrait',
      formSchema: [
        { name: 'subject', label: '拍摄对象', type: 'text', required: true },
        { name: 'era', label: '年代', type: 'select', options: ['80年代', '90年代', '民国'], required: false },
      ],
    },
  ],
  festival: [
    {
      key: 'festival-spring',
      name: '春节',
      sceneKey: 'festival',
      formSchema: [
        { name: 'title', label: '标题', type: 'text', required: true },
        { name: 'date', label: '日期', type: 'date', required: true },
      ],
    },
    {
      key: 'festival-midautumn',
      name: '中秋',
      sceneKey: 'festival',
      formSchema: [
        { name: 'title', label: '标题', type: 'text', required: true },
        { name: 'blessing', label: '祝福语', type: 'textarea', required: false },
      ],
    },
  ],
  invitation: [
    {
      key: 'invitation-wedding',
      name: '婚礼请柬',
      sceneKey: 'invitation',
      formSchema: [
        { name: 'groom', label: '新郎', type: 'text', required: true },
        { name: 'bride', label: '新娘', type: 'text', required: true },
        { name: 'date', label: '婚礼日期', type: 'date', required: true },
      ],
    },
  ],
  tshirt: [
    {
      key: 'tshirt-anime',
      name: '动漫风',
      sceneKey: 'tshirt',
      formSchema: [
        { name: 'theme', label: '主题', type: 'text', required: true },
        { name: 'style', label: '风格描述', type: 'textarea', required: false },
      ],
    },
  ],
  poster: [
    {
      key: 'poster-sale',
      name: '促销海报',
      sceneKey: 'poster',
      formSchema: [
        { name: 'product', label: '商品名称', type: 'text', required: true },
        { name: 'discount', label: '折扣信息', type: 'text', required: true },
      ],
    },
  ],
}

onMounted(() => {
  const pages = getCurrentPages()
  const page = pages[pages.length - 1] as any
  const key = page?.$route?.query?.scene_key ?? page?.options?.scene_key ?? ''
  sceneKey.value = key
  generationStore.setScene(key)
  templates.value = mockTemplates[key] ?? []
})

function handleSelectTemplate(template: Template) {
  generationStore.setTemplate(template)
}

function handleUpdateForm(values: Record<string, string>) {
  generationStore.setFormValues(values)
}

function handleSubmit() {
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
  generationStore.submitGeneration()
  uni.showToast({ title: '提交成功', icon: 'success' })
}
</script>

<template>
  <view class="scene">
    <TemplatePicker
      :templates="templates"
      :selected-key="generationStore.selectedTemplate?.key ?? ''"
      @select="handleSelectTemplate"
    />

    <view v-if="generationStore.selectedTemplate">
      <SceneFieldForm
        :schema="generationStore.selectedTemplate.formSchema"
        :model-value="generationStore.formValues"
        @update:model-value="handleUpdateForm"
      />

      <view class="advanced-toggle" @click="showAdvanced = !showAdvanced">
        <text>{{ showAdvanced ? '收起高级设置' : '展开高级设置' }}</text>
      </view>

      <view v-if="showAdvanced" class="advanced-section">
        <text class="label">自定义提示词</text>
        <textarea
          class="textarea"
          :value="generationStore.advancedPrompt"
          placeholder="输入额外的自定义提示词..."
          @input="generationStore.setAdvancedPrompt(($event as any).detail.value)"
        />
      </view>

      <button class="submit-btn" @click="handleSubmit">开始生成</button>
    </view>
  </view>
</template>

<style scoped>
.scene {
  padding: 24rpx;
}
.advanced-toggle {
  padding: 24rpx;
  text-align: center;
  color: #007aff;
  font-size: 28rpx;
}
.advanced-section {
  padding: 0 24rpx 24rpx;
}
.label {
  display: block;
  font-size: 28rpx;
  color: #333;
  margin-bottom: 8rpx;
}
.textarea {
  padding: 16rpx;
  background-color: #f8f9fa;
  border-radius: 8rpx;
  font-size: 28rpx;
  height: 160rpx;
}
.submit-btn {
  margin: 24rpx;
  background-color: #007aff;
  color: #fff;
  border-radius: 12rpx;
}
</style>
