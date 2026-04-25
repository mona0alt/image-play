import { describe, expect, it } from 'vitest'
import { buildScenePageModel, getSceneSubmitError, resolveSceneKey } from './view-model'

describe('scene view model', () => {
  it('uses requested scene when it exists in scene order', () => {
    expect(resolveSceneKey(['portrait', 'festival'], 'festival')).toBe('festival')
    expect(resolveSceneKey(['portrait', 'festival'], 'unknown')).toBe('portrait')
  })

  it('returns the first required field error before submit', () => {
    expect(getSceneSubmitError({
      key: 'office-pro',
      name: '通勤职业',
      sceneKey: 'portrait',
      formSchema: [
        { name: 'subject_name', label: '拍摄对象', type: 'text', required: true },
      ],
    }, {})).toBe('请填写: 拍摄对象')
  })

  it('builds a submit label from the loading state', () => {
    expect(buildScenePageModel({
      sceneOrder: ['portrait'],
      currentSceneKey: 'portrait',
      templates: [],
      selectedTemplateKey: '',
      isSubmitting: true,
    }).submitLabel).toBe('提交中...')
  })
})
