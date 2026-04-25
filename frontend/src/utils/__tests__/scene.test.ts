import { describe, expect, it } from 'vitest'
import {
  buildScenePresentation,
  getDefaultSceneKey,
  getGalleryScenes,
  getHeroScene,
} from '../scene'

describe('scene presentation', () => {
  it('uses the first configured scene as hero and the rest as gallery', () => {
    const order = ['portrait', 'festival', 'invitation']

    expect(getHeroScene(order).key).toBe('portrait')
    expect(getGalleryScenes(order).map((scene) => scene.key)).toEqual([
      'festival',
      'invitation',
    ])
  })

  it('falls back to portrait when order is empty', () => {
    expect(getDefaultSceneKey([])).toBe('portrait')
  })

  it('builds presentation metadata for a known scene', () => {
    const scene = buildScenePresentation('poster')

    expect(scene.name).toBe('商业海报')
    expect(scene.tags).toEqual(['Editorial', 'Minimal'])
    expect(scene.eyebrow).toBe('Curated Collection')
  })
})
