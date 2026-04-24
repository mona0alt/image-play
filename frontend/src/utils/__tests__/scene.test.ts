import { describe, it, expect } from 'vitest'
import { getHero, getGallery } from '../scene'

describe('scene layout', () => {
  it('renders portrait as hero and remaining scenes as gallery cards', () => {
    const order = ['portrait', 'festival', 'invitation', 'tshirt', 'poster']
    expect(getHero(order)).toBe('portrait')
    expect(getGallery(order)).toEqual(['festival', 'invitation', 'tshirt', 'poster'])
  })
})
