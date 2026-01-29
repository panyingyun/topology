import { ref } from 'vue'

export interface GlobalErrorPayload {
  message: string
  stack?: string
}

export const globalError = ref<GlobalErrorPayload | null>(null)

function isBenign(err: unknown): boolean {
  const msg = err instanceof Error ? err.message : String(err)
  const s = (err instanceof Error ? err.stack : '') || ''
  if (/ResizeObserver loop|Script error\.?|ChunkLoadError|Loading chunk \d+ failed/i.test(msg)) return true
  if (/ResizeObserver|Loading CSS chunk/i.test(s)) return true
  return false
}

export function setGlobalError(e: unknown): void {
  if (isBenign(e)) return
  const message = e instanceof Error ? e.message : String(e)
  const stack = e instanceof Error ? e.stack : undefined
  globalError.value = { message, stack }
}

export function clearGlobalError(): void {
  globalError.value = null
}
