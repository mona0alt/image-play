export function isGenerationPending(status: string): boolean {
  return status === 'queued' || status === 'running' || status === 'result_auditing'
}

export function getStatusMeta(status: string): { label: string; tone: 'neutral' | 'pending' | 'success' | 'danger' } {
  switch (status) {
    case 'queued':
      return { label: '排队中', tone: 'pending' }
    case 'running':
      return { label: '生成中', tone: 'pending' }
    case 'result_auditing':
      return { label: '审核中', tone: 'pending' }
    case 'success':
      return { label: '已完成', tone: 'success' }
    case 'failed':
      return { label: '失败', tone: 'danger' }
    default:
      return { label: status, tone: 'neutral' }
  }
}

export function getResultViewState(item?: { status: string; resultUrl: string }): 'missing' | 'pending' | 'success' | 'failed' | 'empty' {
  if (!item) return 'missing'
  if (isGenerationPending(item.status)) return 'pending'
  if (item.status === 'success' && item.resultUrl) return 'success'
  if (item.status === 'failed') return 'failed'
  return 'empty'
}

export function takeRecentSuccessItems<T extends { status: string; resultUrl: string }>(items: T[], limit = 3): T[] {
  return items.filter((item) => item.status === 'success' && !!item.resultUrl).slice(0, limit)
}

export function formatHistoryDate(ts: string): string {
  const date = new Date(Number(ts) * 1000)
  return `${date.getFullYear()}.${String(date.getMonth() + 1).padStart(2, '0')}.${String(date.getDate()).padStart(2, '0')}`
}

export function filterHistoryItems<T extends { status: string }>(items: T[], status: string): T[] {
  if (status === 'all') return items
  return items.filter((item) => item.status === status)
}

export function findHistoryItemById<T extends { id: number }>(items: T[], id: number): T | undefined {
  return items.find((item) => item.id === id)
}
