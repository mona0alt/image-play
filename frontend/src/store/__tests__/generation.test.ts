import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useGenerationStore } from '../generation'
import * as api from '../../services/api'

describe('generation store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('has default state', () => {
    const store = useGenerationStore()
    expect(store.selectedScene).toBe('')
    expect(store.selectedTemplate).toBeNull()
    expect(store.formValues).toEqual({})
    expect(store.advancedPrompt).toBe('')
    expect(store.templates).toEqual([])
    expect(store.isSubmitting).toBe(false)
    expect(store.submitError).toBe('')
    expect(store.lastGenerationId).toBeNull()
  })

  it('loads templates for a scene', async () => {
    const mockTemplates = [
      { key: 't1', name: 'Template 1', sceneKey: 'scene1', formSchema: [], sampleImageUrl: undefined },
      { key: 't2', name: 'Template 2', sceneKey: 'scene1', formSchema: [], sampleImageUrl: undefined },
    ]
    vi.spyOn(api, 'getSceneTemplates').mockResolvedValue(mockTemplates)

    const store = useGenerationStore()
    await store.loadTemplates('scene1')

    expect(api.getSceneTemplates).toHaveBeenCalledWith('scene1')
    expect(store.selectedScene).toBe('scene1')
    expect(store.templates).toEqual(mockTemplates)
    expect(store.selectedTemplate).toEqual(mockTemplates[0])
    expect(store.formValues).toEqual({})
  })

  it('submits generation and stores generation id', async () => {
    vi.spyOn(api, 'createGeneration').mockResolvedValue({ generation_id: 42 })

    const store = useGenerationStore()
    store.selectedScene = 'scene1'
    store.selectedTemplate = { key: 't1', name: 'T1', sceneKey: 'scene1', formSchema: [] }
    store.formValues = { name: 'test' }

    const id = await store.submitGeneration()

    expect(id).toBe(42)
    expect(store.lastGenerationId).toBe(42)
    expect(store.isSubmitting).toBe(false)
    expect(api.createGeneration).toHaveBeenCalledWith(
      expect.objectContaining({
        scene_key: 'scene1',
        template_key: 't1',
        fields: { name: 'test' },
      })
    )
  })

  it('handles submit error', async () => {
    vi.spyOn(api, 'createGeneration').mockRejectedValue(new Error('network error'))

    const store = useGenerationStore()
    store.selectedScene = 'scene1'
    store.selectedTemplate = { key: 't1', name: 'T1', sceneKey: 'scene1', formSchema: [] }

    await expect(store.submitGeneration()).rejects.toThrow('network error')
    expect(store.submitError).toBe('network error')
    expect(store.isSubmitting).toBe(false)
  })
})
