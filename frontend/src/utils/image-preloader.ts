const PRELOAD_CACHE = new Map<string, Promise<string>>()

export function preloadImage(url: string): Promise<string> {
  if (PRELOAD_CACHE.has(url)) {
    return PRELOAD_CACHE.get(url)!
  }

  const promise = new Promise<string>((resolve, reject) => {
    uni.downloadFile({
      url,
      success: (res) => {
        if (res.statusCode === 200) {
          resolve(res.tempFilePath)
        } else {
          reject(new Error(`Download failed: ${res.statusCode}`))
        }
      },
      fail: (err) => reject(new Error(err.errMsg || 'Download failed')),
    })
  })

  PRELOAD_CACHE.set(url, promise)
  return promise
}

export function preloadImages(urls: string[]): Promise<string[]> {
  return Promise.all(urls.map((url) => preloadImage(url).catch(() => url)))
}

export function clearPreloadCache(url?: string) {
  if (url) {
    PRELOAD_CACHE.delete(url)
  } else {
    PRELOAD_CACHE.clear()
  }
}
