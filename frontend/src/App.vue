<script setup lang="ts">
import { onLaunch } from '@dcloudio/uni-app'

onLaunch(() => {
  const token = uni.getStorageSync('access_token') as string | undefined
  if (!token) {
    uni.reLaunch({ url: '/pages/login/index' })
  }

  // 导航拦截：未登录时阻止跳转到非登录页
  const methods = ['navigateTo', 'redirectTo', 'switchTab', 'reLaunch'] as const
  methods.forEach((method) => {
    uni.addInterceptor(method, {
      invoke(args) {
        const token = uni.getStorageSync('access_token') as string | undefined
        const url = (args as any).url || ''
        if (!token && !url.includes('/pages/login')) {
          uni.reLaunch({ url: '/pages/login/index' })
          return false
        }
      },
    })
  })
})
</script>

<style>
page {
  --gallery-bg: #fdf8f8;
  --gallery-surface: #ffffff;
  --gallery-surface-soft: #f4efed;
  --gallery-border: rgba(28, 27, 27, 0.08);
  --gallery-text: #1c1b1b;
  --gallery-muted: #6d6865;
  --gallery-accent: #111111;
  background: var(--gallery-bg);
  color: var(--gallery-text);
}

view,
text,
button,
image,
scroll-view,
input,
textarea {
  box-sizing: border-box;
}

button {
  border: none;
}

button::after {
  border: none;
}
</style>