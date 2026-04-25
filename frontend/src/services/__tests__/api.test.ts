import { describe, it, expect } from 'vitest'
import { mapSceneTemplate, mapHistoryItem } from '../api'
import type { SceneTemplateDTO, HistoryItemDTO } from '../api'

describe('mapSceneTemplate', () => {
  it('correctly maps snake_case fields to camelCase', () => {
    const dto: SceneTemplateDTO = {
      key: 'template-1',
      name: 'Template One',
      scene_key: 'scene-a',
      form_schema: [
        { name: 'field1', label: 'Field 1', type: 'text', required: true },
      ],
      sample_image_url: 'https://example.com/sample.png',
    }

    const result = mapSceneTemplate(dto)

    expect(result.key).toBe('template-1')
    expect(result.name).toBe('Template One')
    expect(result.sceneKey).toBe('scene-a')
    expect(result.formSchema).toEqual([
      { name: 'field1', label: 'Field 1', type: 'text', required: true },
    ])
    expect(result.sampleImageUrl).toBe('https://example.com/sample.png')
  })

  it('handles optional sample_image_url when missing', () => {
    const dto: SceneTemplateDTO = {
      key: 'template-2',
      name: 'Template Two',
      scene_key: 'scene-b',
      form_schema: [],
    }

    const result = mapSceneTemplate(dto)

    expect(result.sampleImageUrl).toBeUndefined()
  })
})

describe('mapHistoryItem', () => {
  it('correctly maps snake_case fields to camelCase', () => {
    const dto: HistoryItemDTO = {
      id: 1,
      scene_key: 'scene-a',
      template_key: 'template-1',
      status: 'completed',
      result_url: 'https://example.com/result.png',
      created_at: '2026-04-25T10:00:00Z',
    }

    const result = mapHistoryItem(dto)

    expect(result.id).toBe(1)
    expect(result.sceneKey).toBe('scene-a')
    expect(result.templateKey).toBe('template-1')
    expect(result.status).toBe('completed')
    expect(result.resultUrl).toBe('https://example.com/result.png')
    expect(result.createdAt).toBe('2026-04-25T10:00:00Z')
  })
})
