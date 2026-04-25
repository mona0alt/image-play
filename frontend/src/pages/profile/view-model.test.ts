import { describe, expect, it } from 'vitest'
import { buildProfileViewModel } from './view-model'

describe('profile view model', () => {
  it('builds quick scenes, recent works and package cards', () => {
    const vm = buildProfileViewModel({
      profile: { balance: 8, free_quota: 3 },
      packages: [{ code: 'pack10', title: '10次包', price: '8.00', count: 10 }],
      historyItems: [
        {
          id: 1,
          sceneKey: 'poster',
          templateKey: 'concert',
          status: 'success',
          resultUrl: 'https://x/1.png',
          createdAt: '1714003200',
        },
      ],
      sceneOrder: ['portrait', 'poster'],
    })

    expect(vm.accountTitle).toBe('我的作品室')
    expect(vm.quickScenes.map((scene) => scene.key)).toEqual(['portrait', 'poster'])
    expect(vm.recentWorks).toHaveLength(1)
    expect(vm.packages[0]?.actionLabel).toBe('购买')
  })
})
