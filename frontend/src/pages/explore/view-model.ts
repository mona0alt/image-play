import type { ExploreItem } from '../../services/api'

interface Pagination {
  page: number
  pageSize: number
  total: number
  hasMore: boolean
}

export function buildExploreViewModel(input: {
  items: ExploreItem[]
  pagination: Pagination
}) {
  return {
    cards: input.items,
    empty: input.items.length === 0,
    hasMore: input.pagination.hasMore,
  }
}
