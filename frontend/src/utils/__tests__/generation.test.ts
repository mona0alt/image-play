import { describe, it, expect } from 'vitest'
import { isGenerationPending, filterHistoryItems, findHistoryItemById } from '../generation'

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
