import { defineStore } from 'pinia'
import type { Template } from '../types/scene'
import { getSceneTemplates, createGeneration } from '../services/api'

export const useGenerationStore = defineStore('generation', {
  state: () => ({
    selectedScene: '',
    selectedTemplate: null as Template | null,
    formValues: {} as Record<string, string>,
    advancedPrompt: '',
    templates: [] as Template[],
    isSubmitting: false,
    submitError: '',
    lastGenerationId: null as number | null,
  }),
  actions: {
    setScene(key: string) {
      this.selectedScene = key
    },
    setTemplate(template: Template | null) {
      this.selectedTemplate = template
    },
    setFormValues(values: Record<string, string>) {
      this.formValues = values
    },
    setAdvancedPrompt(prompt: string) {
      this.advancedPrompt = prompt
    },
    async loadTemplates(sceneKey: string) {
      const templates = await getSceneTemplates(sceneKey)
      this.selectedScene = sceneKey
      this.templates = templates
      this.selectedTemplate = templates[0] ?? null
      this.formValues = {}
    },
    async submitGeneration() {
      if (!this.selectedScene || !this.selectedTemplate) {
        throw new Error('未选择场景或模板')
      }
      this.isSubmitting = true
      this.submitError = ''
      try {
        const res = await createGeneration({
          client_request_id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
          scene_key: this.selectedScene,
          template_key: this.selectedTemplate.key,
          fields: this.formValues,
        })
        this.lastGenerationId = res.generation_id
        return res.generation_id
      } catch (e: any) {
        this.submitError = e.message || '提交失败'
        throw e
      } finally {
        this.isSubmitting = false
      }
    },
  },
})
