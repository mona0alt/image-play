import { login } from './api'

export async function ensureSession(): Promise<string> {
  const existing = uni.getStorageSync('access_token') as string | undefined
  if (existing) {
    return existing
  }

  const loginRes = await uni.login({ provider: 'weixin' })
  if (!loginRes.code) {
    throw new Error('WeChat login failed: no code returned')
  }

  const res = await login(loginRes.code)
  uni.setStorageSync('access_token', res.access_token)
  return res.access_token
}
