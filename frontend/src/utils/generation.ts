export function isGenerationPending(status: string): boolean {
  return status === 'queued' || status === 'running' || status === 'result_auditing'
}

export function filterHistoryItems<T extends { status: string }>(items: T[], status: string): T[] {
  if (status === 'all') return items
  return items.filter((item) => item.status === status)
}

export function findHistoryItemById<T extends { id: number }>(items: T[], id: number): T | undefined {
  return items.find((item) => item.id === id)
}
