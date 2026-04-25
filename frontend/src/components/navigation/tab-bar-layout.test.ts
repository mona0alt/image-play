import { describe, expect, it } from 'vitest'
import {
  GALLERY_TAB_BAR_HEIGHT_RPX,
  GALLERY_TAB_BAR_SPACE,
  getPageShellBottomPadding,
} from './tab-bar-layout'

describe('tab bar layout', () => {
  it('uses one shared reserved space for the fixed bottom tab bar', () => {
    expect(GALLERY_TAB_BAR_HEIGHT_RPX).toBe(120)
    expect(GALLERY_TAB_BAR_SPACE).toBe('calc(120rpx + env(safe-area-inset-bottom))')
  })

  it('adds the fixed tab bar space to regular page padding', () => {
    expect(getPageShellBottomPadding(false)).toBe('calc(152rpx + env(safe-area-inset-bottom))')
  })

  it('uses the same reserved space for no-padding pages', () => {
    expect(getPageShellBottomPadding(true)).toBe(
      'calc(120rpx + env(safe-area-inset-bottom))',
    )
  })
})
