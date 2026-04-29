import { takeRecentSuccessItems } from '../../utils/generation'
import { buildHomeEntryCards } from './entry-cards'

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
}) {
  return {
    entryCards: buildHomeEntryCards(input.sceneOrder),
    recentWorks: takeRecentSuccessItems(input.historyItems, 3),
  }
}
