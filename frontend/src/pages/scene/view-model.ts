export function buildSubmitLabel(isSubmitting: boolean): string {
  return isSubmitting ? '生成中...' : '开始生成'
}
