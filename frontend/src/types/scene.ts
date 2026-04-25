export interface Scene {
  key: string
  name: string
  description: string
  icon: string
}

export interface Template {
  key: string
  name: string
  sceneKey: string
  formSchema: FormField[]
  sampleImageUrl?: string | undefined
}

export interface FormField {
  name: string
  label: string
  type: 'text' | 'textarea' | 'date' | 'select'
  required?: boolean
  options?: string[]
}
