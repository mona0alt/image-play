import { defineStore } from 'pinia'
import { createGeneration } from '../services/api'

export const useGenerationStore = defineStore('generation', {
  state: () => ({
    prompt: '',
    isSubmitting: false,
    submitError: '',
    lastGenerationId: null as number | null,
  }),
  actions: {
    setPrompt(prompt: string) {
      this.prompt = prompt
    },
    async submitGeneration() {
      if (!this.prompt.trim()) {
        throw new Error('请输入提示词')
      }
      this.isSubmitting = true
      this.submitError = ''
      try {
        const res = await createGeneration({
          client_request_id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
          prompt: this.prompt.trim(),
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
