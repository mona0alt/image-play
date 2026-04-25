export type PrimaryTabKey = 'gallery' | 'create' | 'explore' | 'profile'

export interface PrimaryTab {
  key: PrimaryTabKey
  label: string
  path: string
}

export const PRIMARY_TABS: PrimaryTab[] = [
  { key: 'gallery', label: '艺廊', path: '/pages/home/index' },
  { key: 'create', label: '创作', path: '/pages/scene/index' },
  { key: 'explore', label: '发现', path: '/pages/explore/index' },
  { key: 'profile', label: '我的', path: '/pages/profile/index' },
]

export function findPrimaryTab(key: PrimaryTabKey) {
  return PRIMARY_TABS.find((tab) => tab.key === key)
}
