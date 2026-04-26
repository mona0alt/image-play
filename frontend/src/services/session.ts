export function ensureSession(): Promise<string> {
  const existing = uni.getStorageSync('access_token') as string | undefined
  if (existing) {
    return Promise.resolve(existing)
  }
  uni.reLaunch({ url: '/pages/login/index' })
  return Promise.reject(new Error('Unauthorized'))
}
