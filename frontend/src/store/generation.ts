import { defineStore } from 'pinia'
import type { Template } from '../types/scene'

export const useGenerationStore = defineStore('generation', {
  state: () => ({
    selectedScene: '',
    selectedTemplate: null as Template | null,
    formValues: {} as Record<string, string>,
    advancedPrompt: '',
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
    async submitGeneration() {
      // Mock: just console.log for now
      // eslint-disable-next-line no-console
      console.log('submit', this.formValues, this.advancedPrompt)
    },
  },
})
