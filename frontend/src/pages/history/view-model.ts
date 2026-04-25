import { filterHistoryItems, formatHistoryDate, getStatusMeta } from '../../utils/generation'
import { buildScenePresentation } from '../../utils/scene'

interface HistoryLike {
  id: number
  sceneKey: string
  templateKey: string
  status: string
  resultUrl: string
  createdAt: string
}

export function buildHistoryViewModel(input: { items: HistoryLike[]; filter: string }) {
  return {
    filters: ['all', 'queued', 'running', 'result_auditing', 'success', 'failed'],
    cards: filterHistoryItems(input.items, input.filter).map((item) => ({
      id: item.id,
      title: buildScenePresentation(item.sceneKey).name,
      subtitle: `模板 · ${item.templateKey}`,
      date: formatHistoryDate(item.createdAt),
      imageUrl: item.resultUrl,
      status: getStatusMeta(item.status),
    })),
  }
}
