export interface ScenePresentation {
  key: string
  name: string
  description: string
  eyebrow: string
  icon: string
  tags: string[]
  accent: 'portrait' | 'festival' | 'invitation' | 'tshirt' | 'poster'
}

const SCENE_META: Record<string, Omit<ScenePresentation, 'key'>> = {
  portrait: {
    name: '人像写真',
    description: '打造更适合展示与社交使用的质感人物作品。',
    eyebrow: 'Premium Service',
    icon: '✦',
    tags: ['Portrait', 'Minimal'],
    accent: 'portrait',
  },
  festival: {
    name: '节日海报',
    description: '把节庆氛围浓缩成一张干净、克制的视觉作品。',
    eyebrow: 'Seasonal Edit',
    icon: '✺',
    tags: ['Greeting', 'Warm Light'],
    accent: 'festival',
  },
  invitation: {
    name: '邀请函',
    description: '用更轻盈的排版和留白组织你的邀请信息。',
    eyebrow: 'Paper Studio',
    icon: '✧',
    tags: ['Stationery', 'Elegant'],
    accent: 'invitation',
  },
  tshirt: {
    name: 'T恤图案',
    description: '将主题文案转成适合服饰呈现的图案风格。',
    eyebrow: 'Graphic Lab',
    icon: '✷',
    tags: ['Print', 'Streetwear'],
    accent: 'tshirt',
  },
  poster: {
    name: '商业海报',
    description: '适合活动、餐饮和商业宣传的编辑式画面。',
    eyebrow: 'Curated Collection',
    icon: '✹',
    tags: ['Editorial', 'Minimal'],
    accent: 'poster',
  },
}

export function buildScenePresentation(key: string): ScenePresentation {
  const fallbackKey = key in SCENE_META ? key : 'portrait'
  return { key, ...SCENE_META[fallbackKey] }
}

export function getDefaultSceneKey(order: string[]): string {
  return order[0] ?? 'portrait'
}

export function getHeroScene(order: string[]): ScenePresentation {
  return buildScenePresentation(getDefaultSceneKey(order))
}

export function getGalleryScenes(order: string[]): ScenePresentation[] {
  return order.slice(1).map((key) => buildScenePresentation(key))
}

export function getHero(order: string[]): string | undefined {
  return getDefaultSceneKey(order)
}

export function getGallery(order: string[]): string[] {
  return getGalleryScenes(order).map((scene) => scene.key)
}
