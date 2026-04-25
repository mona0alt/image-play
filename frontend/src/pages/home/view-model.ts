import type { UserProfile } from '../../store/user'
import { takeRecentSuccessItems } from '../../utils/generation'
import { getGalleryScenes, getHeroScene } from '../../utils/scene'

interface HistoryLike {
  id: number
  sceneKey: string
  templateKey: string
  status: string
  resultUrl: string
  createdAt: string
}

export function buildHomeViewModel(input: {
  sceneOrder: string[]
  historyItems: HistoryLike[]
  profile: Pick<UserProfile, 'balance' | 'free_quota'> | null
}) {
  return {
    heroScene: getHeroScene(input.sceneOrder),
    galleryScenes: getGalleryScenes(input.sceneOrder),
    recentWorks: takeRecentSuccessItems(input.historyItems, 3),
    creditTitle: '剩余额度',
    creditValue: String(input.profile?.free_quota ?? 0),
    balanceValue: String(input.profile?.balance ?? 0),
  }
}
