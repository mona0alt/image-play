import { describe, expect, it } from 'vitest'
import { buildHomeEntryCards } from './entry-cards'
import {
  STACK_PEEK_COUNT,
  STACK_SWIPE_THRESHOLD,
  clampStackIndex,
  getVisibleHomeEntries,
  resolveStackSwipe,
} from './entry-stack'

const cards = buildHomeEntryCards([
  'portrait',
  'festival',
  'invitation',
  'tshirt',
  'poster',
])

describe('home entry stack', () => {
  it('exposes one active card plus three peek cards', () => {
    expect(STACK_PEEK_COUNT).toBe(3)

    const visible = getVisibleHomeEntries(cards, 1)

    expect(visible.map((item) => `${item.slot}:${item.entry.key}`)).toEqual([
      'active:face-reading',
      'peek-1:festival',
      'peek-2:invitation',
      'peek-3:tshirt',
    ])
  })

  it('clamps indexes into a safe range', () => {
    expect(clampStackIndex(cards.length, -2)).toBe(0)
    expect(clampStackIndex(cards.length, 99)).toBe(cards.length - 1)
  })

  it('moves to the next card on a large upward swipe', () => {
    expect(STACK_SWIPE_THRESHOLD).toBe(72)
    expect(
      resolveStackSwipe({
        count: cards.length,
        activeIndex: 1,
        deltaY: -100,
      }),
    ).toEqual({ nextIndex: 2, consumed: true })
  })

  it('does not consume upward swipes past the last card', () => {
    expect(
      resolveStackSwipe({
        count: cards.length,
        activeIndex: cards.length - 1,
        deltaY: -100,
      }),
    ).toEqual({
      nextIndex: cards.length - 1,
      consumed: false,
    })
  })
})
