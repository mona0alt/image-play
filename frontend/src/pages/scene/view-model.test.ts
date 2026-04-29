import { describe, expect, it } from 'vitest'
import { buildSubmitLabel } from './view-model'

describe('scene view model', () => {
  it('returns submitting label when loading', () => {
    expect(buildSubmitLabel(true)).toBe('生成中...')
  })

  it('returns default label when not loading', () => {
    expect(buildSubmitLabel(false)).toBe('开始生成')
  })
})
