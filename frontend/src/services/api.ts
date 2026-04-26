const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080'

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
  return request<{ access_token: string; user: { id: number; nickname: string; balance: number; free_quota: number } }>({
    url: '/api/auth/login',
    method: 'POST',
    data: { code },
    headers: { 'Content-Type': 'application/json' },
  })
}

export function getMe() {
  return request<{ id: number; nickname: string; balance: number; free_quota: number }>({
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

export interface ExploreUserDTO {
  id: string
  nickname: string
  avatar_url: string
}

export interface ExploreItemDTO {
  id: number
  user: ExploreUserDTO
  image_url: string
  thumbnail_url: string
  prompt: string
  scene_key: string
  like_count: number
  is_liked: boolean
  description: string
  created_at: string
}

export interface ExploreFeedResponse {
  items: ExploreItemDTO[]
  pagination: {
    page: number
    page_size: number
    total: number
    has_more: boolean
  }
}

export function mapExploreItem(dto: ExploreItemDTO) {
  return {
    id: dto.id,
    user: {
      id: dto.user.id,
      nickname: dto.user.nickname,
      avatarUrl: dto.user.avatar_url,
    },
    imageUrl: dto.image_url,
    thumbnailUrl: dto.thumbnail_url,
    prompt: dto.prompt,
    sceneKey: dto.scene_key,
    likeCount: dto.like_count,
    isLiked: dto.is_liked,
    description: dto.description,
    createdAt: dto.created_at,
  }
}

export type ExploreItem = ReturnType<typeof mapExploreItem>

export function getExploreFeed(page = 1, pageSize = 10) {
  return request<ExploreFeedResponse>({
    url: `/api/explore/feed?page=${page}&page_size=${pageSize}`,
    method: 'GET',
  }).then((res) => ({
    items: (res.items || []).map(mapExploreItem),
    pagination: res.pagination,
  }))
}

export function likeExploreItem(generationId: number, action: 'like' | 'unlike') {
  return request<{ success: boolean; like_count: number }>({
    url: '/api/explore/like',
    method: 'POST',
    data: { generation_id: generationId, action },
    headers: { 'Content-Type': 'application/json' },
  })
}
