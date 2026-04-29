import { buildScenePresentation, type ScenePresentation } from '../../utils/scene'

export const HOME_DEFAULT_SCENE_ORDER = [
  'portrait',
  'festival',
  'invitation',
  'tshirt',
  'poster',
] as const

export type HomeEntryAccent = ScenePresentation['accent'] | 'analysis'

export interface HomeEntryCard {
  kind: 'scene' | 'tool'
  key: string
  title: string
  description: string
  eyebrow: string
  tags: string[]
  accent: HomeEntryAccent
  path: string
  sceneKey?: string
}

const FACE_READING_ENTRY: HomeEntryCard = {
  kind: 'tool',
  key: 'face-reading',
  title: '面相分析',
  description: '上传照片，AI 解析面部特征与数字命格。',
  eyebrow: 'AI Analysis',
  tags: ['Insight', 'Portrait'],
  accent: 'analysis',
  path: '/pages/face-reading/index',
}

function mapSceneEntry(scene: ScenePresentation): HomeEntryCard {
  return {
    kind: 'scene',
    key: scene.key,
    title: scene.name,
    description: scene.description,
    eyebrow: scene.eyebrow,
    tags: scene.tags,
    accent: scene.accent,
    path: `/pages/scene/index?scene_key=${scene.key}`,
    sceneKey: scene.key,
  }
}

export function buildHomeEntryCards(sceneOrder: string[]): HomeEntryCard[] {
  const normalizedOrder = sceneOrder.length
    ? sceneOrder
    : [...HOME_DEFAULT_SCENE_ORDER]

  const sceneCards = normalizedOrder.map((sceneKey) =>
    mapSceneEntry(buildScenePresentation(sceneKey)),
  )

  const leadCard =
    sceneCards[0] ?? mapSceneEntry(buildScenePresentation('portrait'))

  return [leadCard, FACE_READING_ENTRY, ...sceneCards.slice(1)]
}
