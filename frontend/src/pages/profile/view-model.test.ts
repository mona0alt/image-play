import { describe, expect, it } from 'vitest'
import { buildProfileViewModel } from './view-model'

const baseInput = {
  nickname: '小明',
  profile: { balance: 8, free_quota: 3 },
  packages: [{ code: 'pack10', title: '10次包', price: '8.00', count: 10 }],
  historyItems: [
    {
      id: 1,
      sceneKey: 'poster',
      templateKey: 'concert',
      status: 'success',
      resultUrl: 'https://img.com/1.jpg',
      createdAt: '1714003200',
    },
    {
      id: 2,
      sceneKey: 'portrait',
      templateKey: 'graduation',
      status: 'success',
      resultUrl: 'https://img.com/2.jpg',
      createdAt: '1714003201',
    },
    {
      id: 3,
      sceneKey: 'poster',
      templateKey: 'sale',
      status: 'success',
      resultUrl: 'https://img.com/3.jpg',
      createdAt: '1714003202',
    },
    {
      id: 4,
      sceneKey: 'portrait',
      templateKey: 'wedding',
      status: 'success',
      resultUrl: 'https://img.com/4.jpg',
      createdAt: '1714003203',
    },
  ],
  sceneOrder: ['portrait', 'poster'],
}

describe('profile view model', () => {
  it('builds quick scenes, recent works and package cards', () => {
    const vm = buildProfileViewModel({
      nickname: '小明',
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

    expect(vm.accountTitle).toBe('小明的作品室')
    expect(vm.quickScenes.map((scene) => scene.key)).toEqual(['portrait', 'poster'])
    expect(vm.recentWorks).toHaveLength(1)
    expect(vm.packages[0]?.actionLabel).toBe('购买')
  })

  it('includes history entry with thumbnails and total count', () => {
    const vm = buildProfileViewModel(baseInput)

    expect(vm.historyEntry).toBeDefined()
    expect(vm.historyEntry.thumbnails).toHaveLength(3)
    expect(vm.historyEntry.thumbnails[0]).toBe('https://img.com/1.jpg')
    expect(vm.historyEntry.totalCount).toBe(4)
  })

  it('handles empty history', () => {
    const vm = buildProfileViewModel({ ...baseInput, historyItems: [] })
    expect(vm.historyEntry.thumbnails).toHaveLength(0)
    expect(vm.historyEntry.totalCount).toBe(0)
  })
})
