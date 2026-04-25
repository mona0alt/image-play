import { describe, expect, it } from 'vitest'
import { buildHomeViewModel } from './view-model'

describe('home view model', () => {
  it('builds hero, gallery and recent work sections', () => {
    const vm = buildHomeViewModel({
      sceneOrder: ['portrait', 'festival', 'invitation'],
      historyItems: [
        {
          id: 1,
          sceneKey: 'festival',
          templateKey: 'spring',
          status: 'success',
          resultUrl: 'https://x/1.png',
          createdAt: '1714003200',
        },
        {
          id: 2,
          sceneKey: 'portrait',
          templateKey: 'office-pro',
          status: 'failed',
          resultUrl: '',
          createdAt: '1714089600',
        },
      ],
      profile: { balance: 5, free_quota: 2 },
    })

    expect(vm.heroScene.key).toBe('portrait')
    expect(vm.galleryScenes.map((scene) => scene.key)).toEqual(['festival', 'invitation'])
    expect(vm.recentWorks.map((item) => item.id)).toEqual([1])
    expect(vm.creditTitle).toBe('剩余额度')
    expect(vm.creditValue).toBe('2')
  })
})
