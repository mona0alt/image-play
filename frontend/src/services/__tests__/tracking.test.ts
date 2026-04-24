import { describe, it, expect } from 'vitest'
import { mapTrackingEvent } from '../tracking'

describe('tracking event mapping', () => {
  it('maps save and share actions to expected event names', () => {
    expect(mapTrackingEvent('save')).toBe('generation_saved')
    expect(mapTrackingEvent('share')).toBe('generation_shared')
  })
})
