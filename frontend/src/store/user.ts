import { defineStore } from 'pinia'

export interface UserProfile {
  id: number
  openid: string
  balance: number
  free_quota: number
}

export const useUserStore = defineStore('user', {
  state: () => ({
    token: uni.getStorageSync('access_token') || '',
    profile: null as UserProfile | null,
  }),
  actions: {
    setToken(token: string) {
      this.token = token
      uni.setStorageSync('access_token', token)
    },
    setProfile(profile: UserProfile) {
      this.profile = profile
    },
    clear() {
      this.token = ''
      this.profile = null
      uni.removeStorageSync('access_token')
    },
  },
})
