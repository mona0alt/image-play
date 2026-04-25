export const GALLERY_TAB_BAR_HEIGHT_RPX = 120
const PAGE_SHELL_BASE_BOTTOM_PADDING_RPX = 32

export const GALLERY_TAB_BAR_SPACE =
  `calc(${GALLERY_TAB_BAR_HEIGHT_RPX}rpx + env(safe-area-inset-bottom))`

export function getPageShellBottomPadding(noPadding: boolean) {
  if (noPadding) {
    return GALLERY_TAB_BAR_SPACE
  }

  return `calc(${PAGE_SHELL_BASE_BOTTOM_PADDING_RPX + GALLERY_TAB_BAR_HEIGHT_RPX}rpx + env(safe-area-inset-bottom))`
}
