import { ensureSession } from './session'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://192.168.0.105:8080'
const LOGIN_URL = '/api/auth/login'

interface ApiResponse<T> {
  data: T
  statusCode: number
  header: Record<string, string>
  errMsg: string
}

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: string | Record<string, unknown> | unknown[]
  headers?: Record<string, string>
  skipAuth?: boolean
}

export async function request<T>(options: RequestOptions): Promise<T> {
  let token = ''
  if (!options.skipAuth) {
    token = await ensureSession()
  }

  const headers: Record<string, string> = {
    ...(options.headers || {}),
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  return new Promise((resolve, reject) => {
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
          uni.reLaunch({ url: '/pages/login/index' })
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
  return request<{ access_token: string; user: { id: number; nickname: string; balance: number; free_quota: number }; is_new: boolean }>({
    url: LOGIN_URL,
    method: 'POST',
    data: { code },
    headers: { 'Content-Type': 'application/json' },
    skipAuth: true,
  })
}

export function getMe() {
  return request<{ id: number; nickname: string; balance: number; free_quota: number }>({
    url: '/api/me',
    method: 'GET',
  })
}

export function updateMe(nickname: string) {
  return request<{ id: number; nickname: string; balance: number; free_quota: number }>({
    url: '/api/me',
    method: 'PUT',
    data: { nickname },
    headers: { 'Content-Type': 'application/json' },
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

export function likeExploreItem(exploreAssetId: number, action: 'like' | 'unlike') {
  return request<{ success: boolean; like_count: number }>({
    url: '/api/explore/like',
    method: 'POST',
    data: { explore_asset_id: exploreAssetId, action },
    headers: { 'Content-Type': 'application/json' },
  })
}

export function faceReading(imageBase64: string) {
  return request<{ result: string }>({
    url: '/api/face-reading',
    method: 'POST',
    data: { image_base64: imageBase64 },
    headers: { 'Content-Type': 'application/json' },
  })
}

class SSEParser {
  private buffer = ''

  parse(chunk: ArrayBuffer | string): string[] {
    let text: string
    if (typeof chunk === 'string') {
      text = chunk
    } else {
      // 微信小程序不支持 TextDecoder，使用兼容方式解码 UTF-8
      const uint8Array = new Uint8Array(chunk)
      const raw = String.fromCharCode.apply(null, uint8Array as any)
      text = decodeURIComponent(escape(raw))
    }
    this.buffer += text

    const messages: string[] = []

    while (true) {
      const index = this.buffer.indexOf('\n\n')
      if (index === -1) break

      const message = this.buffer.substring(0, index)
      this.buffer = this.buffer.substring(index + 2)

      const lines = message.split('\n')
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.substring(6)
          if (data === '[DONE]') continue
          try {
            const parsed = JSON.parse(data)
            if (parsed.chunk) {
              messages.push(parsed.chunk)
            }
          } catch {
            // ignore malformed JSON
          }
        }
      }
    }

    return messages
  }
}

export async function faceReadingStream(
  imageBase64: string,
  onChunk: (chunk: string) => void,
): Promise<void> {
  console.log('[face-reading] ========== faceReadingStream called ==========')

  console.log('[face-reading] calling ensureSession...')
  const token = await ensureSession()
  console.log('[face-reading] token got:', token ? 'yes' : 'no')

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  console.log('[face-reading] headers:', JSON.stringify(headers))

  const parser = new SSEParser()

  return new Promise((resolve, reject) => {
    // @ts-ignore 微信小程序原生全局对象
    const hasWx = typeof wx !== 'undefined' && wx.request
    // @ts-ignore
    const requestFn = hasWx ? wx.request : (uni.request as any)
    console.log('[face-reading] requestFn:', hasWx ? 'wx.request' : 'uni.request')

    let requestTask: any
    let aborted = false

    console.log('[face-reading] about to call requestFn...')
    try {
      requestTask = requestFn({
        url: `${API_BASE}/api/face-reading`,
        method: 'POST',
        data: { image_base64: imageBase64 },
        header: headers,
        enableChunked: true,
        responseType: 'arraybuffer',
        success: (res: any) => {
          console.log('[face-reading] >>> success callback, statusCode:', res.statusCode, 'dataType:', typeof res.data)
          if (aborted) {
            console.log('[face-reading] success ignored because aborted')
            return
          }
          if (res.statusCode === 401) {
            uni.removeStorageSync('access_token')
            uni.reLaunch({ url: '/pages/login/index' })
            reject(new Error('Unauthorized'))
            return
          }
          if (res.statusCode >= 200 && res.statusCode < 300) {
            console.log('[face-reading] resolving promise')
            resolve()
          } else {
            console.log('[face-reading] rejecting with HTTP', res.statusCode)
            reject(new Error(`HTTP ${res.statusCode}`))
          }
        },
        fail: (err: any) => {
          console.log('[face-reading] >>> fail callback:', JSON.stringify(err))
          if (aborted) {
            console.log('[face-reading] fail ignored because aborted')
            return
          }
          reject(new Error(err.errMsg || 'Request failed'))
        },
      })
      console.log('[face-reading] requestTask returned, type:', typeof requestTask)
      if (requestTask && typeof requestTask === 'object') {
        console.log('[face-reading] requestTask keys:', Object.keys(requestTask).join(','))
      }
    } catch (e: any) {
      console.log('[face-reading] >>> request exception:', e.message || e)
      reject(e)
      return
    }

    console.log('[face-reading] checking onChunkReceived...')
    console.log('[face-reading] requestTask?.onChunkReceived:', typeof requestTask?.onChunkReceived)

    if (!requestTask || typeof requestTask.onChunkReceived !== 'function') {
      console.log('[face-reading] onChunkReceived not available, fallback to normal request with 60s timeout')
      aborted = true
      try {
        if (requestTask && typeof requestTask.abort === 'function') {
          requestTask.abort()
        }
      } catch (e: any) {
        console.log('[face-reading] abort error:', e.message || e)
      }

      requestFn({
        url: `${API_BASE}/api/face-reading`,
        method: 'POST',
        data: { image_base64: imageBase64 },
        header: headers,
        timeout: 60000,
        success: (res: any) => {
          console.log('[face-reading] fallback success, status:', res.statusCode)
          if (res.statusCode === 401) {
            uni.removeStorageSync('access_token')
            uni.reLaunch({ url: '/pages/login/index' })
            reject(new Error('Unauthorized'))
            return
          }
          if (res.statusCode >= 200 && res.statusCode < 300) {
            const text = typeof res.data === 'string' ? res.data : ''
            console.log('[face-reading] fallback response text length:', text.length)
            const chunks = parser.parse(text)
            console.log('[face-reading] fallback parsed chunks:', chunks.length)
            for (const chunk of chunks) {
              onChunk(chunk)
            }
            resolve()
          } else {
            reject(new Error(`HTTP ${res.statusCode}`))
          }
        },
        fail: (err: any) => {
          console.log('[face-reading] fallback fail:', JSON.stringify(err))
          reject(new Error(err.errMsg || 'Request failed'))
        },
      })
      return
    }

    console.log('[face-reading] registering onChunkReceived...')
    requestTask.onChunkReceived((res: any) => {
      try {
        console.log('[face-reading] >>> onChunkReceived, data type:', typeof res.data, 'is ArrayBuffer:', res.data instanceof ArrayBuffer)
        const chunks = parser.parse(res.data)
        console.log('[face-reading] parsed chunks count:', chunks.length, 'first chunk:', chunks[0]?.substring(0, 20))
        for (const chunk of chunks) {
          onChunk(chunk)
        }
      } catch (e: any) {
        console.error('[face-reading] SSE parse error:', e.message || e)
      }
    })
    console.log('[face-reading] onChunkRegistered registered')
  })
}
