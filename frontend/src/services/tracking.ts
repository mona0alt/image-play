import { request } from './api'

export function mapTrackingEvent(action: string): string {
  const mapping: Record<string, string> = {
    save: 'generation_saved',
    share: 'generation_shared',
    retry: 'generation_retried',
  }
  return mapping[action] || action
}

export function track(event: string, payload: Record<string, unknown>) {
  return request<void>({
    url: '/api/tracking/events',
    method: 'POST',
    data: { event, payload },
    headers: { 'Content-Type': 'application/json' },
  })
}
