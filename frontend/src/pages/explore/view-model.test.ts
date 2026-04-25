import { describe, it, expect } from 'vitest'
import { buildExploreViewModel } from './view-model'

describe('buildExploreViewModel', () => {
  const mockItems = [
    {
      id: 1,
      user: { id: 'u1', nickname: 'User1', avatarUrl: 'https://a.com/1.jpg' },
      imageUrl: 'https://img.com/1.jpg',
      thumbnailUrl: 'https://img.com/1_t.jpg',
      prompt: 'prompt1',
      sceneKey: 'portrait',
      likeCount: 10,
      isLiked: false,
      description: 'desc1',
      createdAt: '2026-04-20T10:00:00Z',
    },
  ]

  it('builds cards from items', () => {
    const vm = buildExploreViewModel({
      items: mockItems,
      pagination: { page: 1, pageSize: 10, total: 1, hasMore: false },
    })

    expect(vm.cards).toHaveLength(1)
    expect(vm.cards[0]!.id).toBe(1)
    expect(vm.cards[0]!.user.nickname).toBe('User1')
    expect(vm.cards[0]!.likeCount).toBe(10)
    expect(vm.cards[0]!.isLiked).toBe(false)
  })

  it('returns empty cards when no items', () => {
    const vm = buildExploreViewModel({
      items: [],
      pagination: { page: 1, pageSize: 10, total: 0, hasMore: false },
    })
    expect(vm.cards).toHaveLength(0)
    expect(vm.empty).toBe(true)
  })
})
