import { describe, it, expect } from 'vitest'
import {
  filterHistoryItems,
  findHistoryItemById,
  formatHistoryDate,
  getResultViewState,
  getStatusMeta,
  isGenerationPending,
  takeRecentSuccessItems,
} from '../generation'

describe('isGenerationPending', () => {
  it('returns true for queued', () => {
    expect(isGenerationPending('queued')).toBe(true)
  })
  it('returns true for running', () => {
    expect(isGenerationPending('running')).toBe(true)
  })
  it('returns true for result_auditing', () => {
    expect(isGenerationPending('result_auditing')).toBe(true)
  })
  it('returns false for success', () => {
    expect(isGenerationPending('success')).toBe(false)
  })
  it('returns false for failed', () => {
    expect(isGenerationPending('failed')).toBe(false)
  })
})

describe('filterHistoryItems', () => {
  const items = [
    { id: 1, status: 'success' },
    { id: 2, status: 'queued' },
    { id: 3, status: 'failed' },
    { id: 4, status: 'queued' },
  ]

  it('returns all items when status is all', () => {
    expect(filterHistoryItems(items, 'all')).toEqual(items)
  })

  it('filters by exact status match for success', () => {
    expect(filterHistoryItems(items, 'success')).toEqual([{ id: 1, status: 'success' }])
  })

  it('filters by exact status match for queued', () => {
    expect(filterHistoryItems(items, 'queued')).toEqual([
      { id: 2, status: 'queued' },
      { id: 4, status: 'queued' },
    ])
  })
})

describe('generation display helpers', () => {
  it('maps running-like statuses to pending tone', () => {
    expect(getStatusMeta('queued')).toEqual({ label: '排队中', tone: 'pending' })
    expect(getStatusMeta('result_auditing')).toEqual({ label: '审核中', tone: 'pending' })
  })

  it('maps result records into result page states', () => {
    expect(getResultViewState(undefined)).toBe('missing')
    expect(getResultViewState({ status: 'running', resultUrl: '' })).toBe('pending')
    expect(getResultViewState({ status: 'success', resultUrl: 'https://x/y.png' })).toBe('success')
    expect(getResultViewState({ status: 'failed', resultUrl: '' })).toBe('failed')
  })

  it('returns recent successful items only', () => {
    const items = [
      { id: 1, status: 'failed', resultUrl: '' },
      { id: 2, status: 'success', resultUrl: 'https://x/2.png' },
      { id: 3, status: 'success', resultUrl: 'https://x/3.png' },
    ]

    expect(takeRecentSuccessItems(items, 2).map((item) => item.id)).toEqual([2, 3])
  })

  it('formats unix-second timestamps into yyyy.mm.dd', () => {
    expect(formatHistoryDate('1714003200')).toBe('2024.04.25')
  })
})

describe('findHistoryItemById', () => {
  const items = [
    { id: 1, status: 'success' },
    { id: 2, status: 'queued' },
  ]

  it('finds the correct item by id', () => {
    expect(findHistoryItemById(items, 2)).toEqual({ id: 2, status: 'queued' })
  })

  it('returns undefined when id is not found', () => {
    expect(findHistoryItemById(items, 999)).toBeUndefined()
  })
})
