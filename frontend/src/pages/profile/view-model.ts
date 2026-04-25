import { takeRecentSuccessItems } from '../../utils/generation'
import { buildScenePresentation } from '../../utils/scene'

export function buildProfileViewModel(input: {
  profile: { balance: number; free_quota: number } | null
  packages: { code: string; title: string; price: string; count: number }[]
  historyItems: { id: number; sceneKey: string; templateKey: string; status: string; resultUrl: string; createdAt: string }[]
  sceneOrder: string[]
}) {
  const recentWorks = takeRecentSuccessItems(input.historyItems, 4)

  return {
    accountTitle: '我的作品室',
    balance: String(input.profile?.balance ?? 0),
    freeQuota: String(input.profile?.free_quota ?? 0),
    recentWorks,
    quickScenes: input.sceneOrder.slice(0, 3).map((key) => buildScenePresentation(key)),
    packages: input.packages.map((item) => ({
      ...item,
      actionLabel: '购买',
    })),
    historyEntry: {
      thumbnails: recentWorks.map((w) => w.resultUrl).slice(0, 3),
      totalCount: input.historyItems.length,
    },
  }
}
