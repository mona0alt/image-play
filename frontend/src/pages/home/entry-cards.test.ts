import { describe, expect, it } from 'vitest'
import { buildHomeEntryCards, HOME_DEFAULT_SCENE_ORDER } from './entry-cards'

describe('home entry cards', () => {
  it('inserts face reading right after the lead scene', () => {
    const cards = buildHomeEntryCards(['portrait', 'festival', 'invitation'])

    expect(cards.map((card) => card.key)).toEqual([
      'portrait',
      'face-reading',
      'festival',
      'invitation',
    ])
    expect(cards[1]).toMatchObject({
      kind: 'tool',
      title: '面相分析',
      accent: 'analysis',
      path: '/pages/face-reading/index',
    })
  })

  it('falls back to the static gallery order when scene_order is empty', () => {
    expect(HOME_DEFAULT_SCENE_ORDER).toEqual([
      'portrait',
      'festival',
      'invitation',
      'tshirt',
      'poster',
    ])
    expect(buildHomeEntryCards([]).map((card) => card.key)).toEqual([
      'portrait',
      'face-reading',
      'festival',
      'invitation',
      'tshirt',
      'poster',
    ])
  })
})
