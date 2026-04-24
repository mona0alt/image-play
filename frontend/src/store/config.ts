import { defineStore } from 'pinia'

export interface ClientConfig {
  brand_slogan: string
  pricing: Record<string, string>
  scene_order: string[]
}

export const useConfigStore = defineStore('config', {
  state: () => ({
    clientConfig: null as ClientConfig | null,
  }),
  actions: {
    setClientConfig(config: ClientConfig) {
      this.clientConfig = config
    },
  },
})
