// TODO: make API_BASE configurable via env at build time
const API_BASE = 'http://localhost:8080'

interface ApiResponse<T> {
  data: T
  statusCode: number
  header: Record<string, string>
  errMsg: string
}

function request<T>(options: {
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
          uni.reLaunch({ url: '/pages/login/login' })
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
