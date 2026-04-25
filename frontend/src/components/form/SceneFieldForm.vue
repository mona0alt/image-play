<script setup lang="ts">
import { ref, watch } from 'vue'
import type { FormField } from '../../types/scene'

interface Props {
  schema: FormField[]
  modelValue: Record<string, string>
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'update:modelValue', values: Record<string, string>): void
}>()

const localValues = ref<Record<string, string>>({ ...props.modelValue })

watch(
  () => props.schema,
  () => {
    const next: Record<string, string> = {}
    for (const field of props.schema) {
      next[field.name] = localValues.value[field.name] ?? ''
    }
    localValues.value = next
    emit('update:modelValue', next)
  },
  { immediate: true },
)

watch(
  () => props.modelValue,
  (val) => {
    localValues.value = { ...val }
  },
  { deep: true },
)

function updateField(name: string, value: string) {
  localValues.value[name] = value
  emit('update:modelValue', { ...localValues.value })
}
</script>

<template>
  <view class="scene-field-form">
    <view v-for="field in schema" :key="field.name" class="form-item">
      <text class="label">{{ field.label }}{{ field.required ? ' *' : '' }}</text>
      <input
        v-if="field.type === 'text' || field.type === 'date'"
        :type="field.type === 'date' ? 'date' : 'text'"
        class="input"
        :value="localValues[field.name] ?? ''"
        :placeholder="field.label"
        @input="updateField(field.name, ($event as any).detail.value)"
      />
      <textarea
        v-else-if="field.type === 'textarea'"
        class="textarea"
        :value="localValues[field.name] ?? ''"
        :placeholder="field.label"
        @input="updateField(field.name, ($event as any).detail.value)"
      />
      <picker
        v-else-if="field.type === 'select' && field.options"
        mode="selector"
        :range="field.options"
        :value="Math.max(0, field.options.indexOf(localValues[field.name] ?? ''))"
        @change="updateField(field.name, field.options![($event as any).detail.value] as string)"
      >
        <view class="picker">{{ localValues[field.name] || '请选择' }}</view>
      </picker>
    </view>
  </view>
</template>

<style scoped>
.scene-field-form {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}
.form-item {
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}
.label {
  display: block;
  font-size: 24rpx;
  color: var(--gallery-muted);
}
.input,
.textarea,
.picker {
  width: 100%;
  padding: 24rpx;
  background: var(--gallery-surface);
  border-radius: 20rpx;
  border: 1rpx solid var(--gallery-border);
  font-size: 28rpx;
  color: var(--gallery-text);
}
.textarea {
  min-height: 180rpx;
}
.picker {
  color: var(--gallery-text);
}
</style>
