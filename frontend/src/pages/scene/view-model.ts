import type { Template } from '../../types/scene'
import { buildScenePresentation, getDefaultSceneKey } from '../../utils/scene'

export function resolveSceneKey(sceneOrder: string[], requestedKey = ''): string {
  return sceneOrder.includes(requestedKey) ? requestedKey : getDefaultSceneKey(sceneOrder)
}

export function getSceneSubmitError(template: Template | null, formValues: Record<string, string>): string | null {
  if (!template) return '请选择模板'
  const missingField = template.formSchema.find((field) => field.required && !formValues[field.name])
  return missingField ? `请填写: ${missingField.label}` : null
}

export function buildScenePageModel(input: {
  sceneOrder: string[]
  currentSceneKey: string
  templates: Template[]
  selectedTemplateKey: string
  isSubmitting: boolean
}) {
  return {
    sceneChoices: input.sceneOrder.map((key) => buildScenePresentation(key)),
    currentScene: buildScenePresentation(input.currentSceneKey),
    submitLabel: input.isSubmitting ? '提交中...' : '开始生成作品',
    hasTemplates: input.templates.length > 0,
    selectedTemplateKey: input.selectedTemplateKey,
  }
}
