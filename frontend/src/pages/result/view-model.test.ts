import { describe, expect, it } from 'vitest'
import { buildResultViewModel } from './view-model'

describe('result view model', () => {
  const items = [
    { id: 1, sceneKey: 'portrait', templateKey: 'office-pro', status: 'running', resultUrl: '', createdAt: '1714003200' },
    { id: 2, sceneKey: 'poster', templateKey: 'concert', status: 'success', resultUrl: 'https://x/2.png', createdAt: '1714089600' },
    { id: 3, sceneKey: 'festival', templateKey: 'spring', status: 'failed', resultUrl: '', createdAt: '1714176000' },
  ]

  it('builds pending state for in-progress generations', () => {
    expect(buildResultViewModel(items, 1).state).toBe('pending')
  })

  it('builds success summary and recommendations', () => {
    const vm = buildResultViewModel(items, 2)
    expect(vm.state).toBe('success')
    expect(vm.title).toBe('商业海报')
    expect(vm.summary).toContain('concert')
    expect(vm.recommendations.map((item) => item.id)).toEqual([2])
  })

  it('returns missing when generation is absent', () => {
    expect(buildResultViewModel(items, 999).state).toBe('missing')
  })
})
