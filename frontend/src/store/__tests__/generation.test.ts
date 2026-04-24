import { describe, it, expect } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useGenerationStore } from '../generation'

describe('generation store', () => {
  it('has default state', () => {
    setActivePinia(createPinia())
    const store = useGenerationStore()
    expect(store.selectedScene).toBe('')
    expect(store.selectedTemplate).toBeNull()
    expect(store.formValues).toEqual({})
    expect(store.advancedPrompt).toBe('')
  })
})
