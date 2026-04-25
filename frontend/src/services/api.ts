// TODO: make API_BASE configurable via env at build time
const API_BASE = 'http://localhost:8080'

interface ApiResponse<T> {
  data: T
  statusCode: number
  header: Record<string, string>
  errMsg: string
}

export function request<T>(options: {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: string | Record<string, unknown> | unknown[]
  headers?: Record<string, string>
}): Promise<T> {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync('access_token') || ''
    const headers: Record<string, string> = {
      ...(options.headers || {}),
    }
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    uni.request({
      url: `${API_BASE}${options.url}`,
      method: options.method || 'GET',
      data: options.data as string | AnyObject | ArrayBuffer,
      header: headers,
      timeout: 15000,
      success: (res) => {
        const response = res as unknown as ApiResponse<T>
        if (response.statusCode === 401) {
          uni.removeStorageSync('access_token')
          uni.reLaunch({ url: '/pages/home/index' })
          reject(new Error('Unauthorized'))
          return
        }
        if (response.statusCode >= 200 && response.statusCode < 300) {
          resolve(response.data)
        } else {
          reject(new Error(response.errMsg || `HTTP ${response.statusCode}`))
        }
      },
      fail: (err: UniNamespace.GeneralCallbackResult) => {
        reject(new Error(err.errMsg || 'Request failed'))
      },
    })
  })
}

export function login(code: string) {
  return request<{ access_token: string; user: { id: number; openid: string; balance: number; free_quota: number } }>({
    url: '/api/auth/login',
    method: 'POST',
    data: { code },
    headers: { 'Content-Type': 'application/json' },
  })
}

export function getMe() {
  return request<{ id: number; openid: string; balance: number; free_quota: number }>({
    url: '/api/me',
    method: 'GET',
  })
}

export function getClientConfig() {
  return request<{ brand_slogan: string; pricing: Record<string, string>; scene_order: string[] }>({
    url: '/api/configs/client',
    method: 'GET',
  })
}

export function getPackages() {
  return request<{ packages: { code: string; title: string; price: string; count: number }[] }>({
    url: '/api/packages',
    method: 'GET',
  })
}

export function createOrder(packageCode: string) {
  return request<{ order_no: string; package_code: string; amount: string; prepay_id: string }>({
    url: '/api/orders',
    method: 'POST',
    data: { package_code: packageCode },
    headers: { 'Content-Type': 'application/json' },
  })
}

export function getHistory() {
  return request<{ items: { id: number; scene_key: string; template_key: string; status: string; result_url: string; created_at: string }[] }>({
    url: '/api/history',
    method: 'GET',
  })
}

export interface SceneTemplateDTO {
  key: string
  name: string
  scene_key: string
  form_schema: { name: string; label: string; type: 'text' | 'textarea' | 'date' | 'select'; required?: boolean; options?: string[] }[]
  sample_image_url?: string
}

export interface HistoryItemDTO {
  id: number
  scene_key: string
  template_key: string
  status: string
  result_url: string
  created_at: string
}

export function mapSceneTemplate(dto: SceneTemplateDTO) {
  return {
    key: dto.key,
    name: dto.name,
    sceneKey: dto.scene_key,
    formSchema: dto.form_schema,
    sampleImageUrl: dto.sample_image_url,
  }
}

export function mapHistoryItem(dto: HistoryItemDTO) {
  return {
    id: dto.id,
    sceneKey: dto.scene_key,
    templateKey: dto.template_key,
    status: dto.status,
    resultUrl: dto.result_url,
    createdAt: dto.created_at,
  }
}

export function getSceneTemplates(sceneKey: string) {
  return request<{ items: SceneTemplateDTO[] }>({
    url: `/api/scenes/${sceneKey}/templates`,
    method: 'GET',
  }).then((res) => res.items.map(mapSceneTemplate))
}

export function createGeneration(payload: {
  client_request_id: string
  scene_key: string
  template_key: string
  fields: Record<string, string>
}) {
  return request<{ generation_id: number }>({
    url: '/api/generations',
    method: 'POST',
    data: payload,
    headers: { 'Content-Type': 'application/json' },
  })
}
