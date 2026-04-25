import { describe, expect, it } from 'vitest'
import { PRIMARY_TABS, findPrimaryTab } from '../navigation'

describe('navigation', () => {
  it('defines the four primary tabs in gallery order', () => {
    expect(PRIMARY_TABS.map((tab) => tab.key)).toEqual([
      'gallery',
      'create',
      'history',
      'profile',
    ])
    expect(PRIMARY_TABS.map((tab) => tab.path)).toEqual([
      '/pages/home/index',
      '/pages/scene/index',
      '/pages/history/index',
      '/pages/profile/index',
    ])
  })

  it('returns the requested primary tab', () => {
    expect(findPrimaryTab('create')?.label).toBe('创作')
    expect(findPrimaryTab('profile')?.path).toBe('/pages/profile/index')
  })
})
