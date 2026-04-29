import type { HomeEntryCard } from './entry-cards'

export type HomeEntrySlot = 'active' | 'peek-1' | 'peek-2' | 'peek-3'

export interface VisibleHomeEntry {
  entry: HomeEntryCard
  index: number
  slot: HomeEntrySlot
}

export const STACK_PEEK_COUNT = 3
export const STACK_SWIPE_THRESHOLD = 72

export function clampStackIndex(length: number, index: number): number {
  if (length <= 0) return 0
  return Math.max(0, Math.min(index, length - 1))
}

export function getVisibleHomeEntries(
  entries: HomeEntryCard[],
  activeIndex: number,
  peekCount = STACK_PEEK_COUNT,
): VisibleHomeEntry[] {
  if (!entries.length) return []

  const start = clampStackIndex(entries.length, activeIndex)
  const visible = entries.slice(start, start + peekCount + 1)

  return visible.map((entry, offset) => ({
    entry,
    index: start + offset,
    slot: offset === 0 ? 'active' : (`peek-${offset}` as HomeEntrySlot),
  }))
}

export function resolveStackSwipe(input: {
  count: number
  activeIndex: number
  deltaY: number
  threshold?: number
}): { nextIndex: number; consumed: boolean } {
  const threshold = input.threshold ?? STACK_SWIPE_THRESHOLD
  const current = clampStackIndex(input.count, input.activeIndex)

  if (input.deltaY <= -threshold && current < input.count - 1) {
    return { nextIndex: current + 1, consumed: true }
  }

  if (input.deltaY >= threshold && current > 0) {
    return { nextIndex: current - 1, consumed: true }
  }

  return { nextIndex: current, consumed: false }
}
