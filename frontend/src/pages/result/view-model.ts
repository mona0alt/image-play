import { findHistoryItemById, formatHistoryDate, getResultViewState, takeRecentSuccessItems } from '../../utils/generation'
import { buildScenePresentation } from '../../utils/scene'

interface HistoryLike {
  id: number
  sceneKey: string
  templateKey: string
  status: string
  resultUrl: string
  createdAt: string
}

export function buildResultViewModel(items: HistoryLike[], generationId: number) {
  const currentItem = findHistoryItemById(items, generationId)
  const state = getResultViewState(currentItem)

  if (!currentItem) {
    return {
      state,
      title: '未找到生成记录',
      summary: '',
      chips: [] as string[],
      recommendations: [] as HistoryLike[],
      currentItem,
    }
  }

  const scene = buildScenePresentation(currentItem.sceneKey)

  return {
    state,
    title: scene.name,
    summary: `${currentItem.templateKey} · ${formatHistoryDate(currentItem.createdAt)}`,
    chips: scene.tags,
    currentItem,
    recommendations: takeRecentSuccessItems(items, 4),
  }
}
