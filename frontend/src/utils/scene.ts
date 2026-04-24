export function getHero(order: string[]): string | undefined {
  return order[0]
}

export function getGallery(order: string[]): string[] {
  return order.slice(1)
}
