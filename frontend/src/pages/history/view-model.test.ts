import { describe, expect, it } from 'vitest'
import { buildHistoryViewModel } from './view-model'

describe('history view model', () => {
  it('filters archive cards and maps status labels', () => {
    const vm = buildHistoryViewModel({
      items: [
        {
          id: 1,
          sceneKey: 'portrait',
          templateKey: 'office-pro',
          status: 'success',
          resultUrl: 'https://x/1.png',
          createdAt: '1714003200',
        },
        {
          id: 2,
          sceneKey: 'festival',
          templateKey: 'spring',
          status: 'running',
          resultUrl: '',
          createdAt: '1714089600',
        },
      ],
      filter: 'running',
    })

    expect(vm.cards).toHaveLength(1)
    expect(vm.cards[0]?.status.label).toBe('生成中')
    expect(vm.cards[0]?.title).toBe('节日海报')
  })
})
