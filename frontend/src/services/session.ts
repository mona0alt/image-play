import { login } from './api'

export async function ensureMockSession(): Promise<string> {
  const existing = uni.getStorageSync('access_token') as string | undefined
  if (existing) {
    return existing
  }

  const code = `mock-code-${Date.now()}`
  const res = await login(code)
  uni.setStorageSync('access_token', res.access_token)
  return res.access_token
}
