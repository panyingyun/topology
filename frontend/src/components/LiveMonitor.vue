<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { X } from 'lucide-vue-next'
import { startMonitor, stopMonitor } from '../services/monitorService'
import type { LiveStatsPayload } from '../types'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  connectionId: string
  connectionName: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const liveStats = ref<LiveStatsPayload | null>(null)
let unsubscribe: (() => void) | null = null
const SLOW_THRESHOLD_SEC = 5

function start() {
  if (!props.connectionId) return
  liveStats.value = null
  startMonitor(props.connectionId).then((res) => {
    if (res.error) {
      liveStats.value = { connectionId: props.connectionId, threadsConnected: 0, processList: [], error: res.error }
      return
    }
    unsubscribe = EventsOn('live-stats', (data: string) => {
      try {
        const payload = JSON.parse(data) as LiveStatsPayload
        if (payload.connectionId === props.connectionId) {
          liveStats.value = payload
        }
      } catch {
        // ignore
      }
    })
  })
}

function stop() {
  if (props.connectionId) {
    stopMonitor(props.connectionId).catch(() => {})
  }
  if (unsubscribe) {
    unsubscribe()
    unsubscribe = null
  }
  liveStats.value = null
}

watch(
  () => [props.show, props.connectionId] as const,
  ([show, id]) => {
    if (show && id) {
      start()
    } else {
      stop()
    }
  },
  { immediate: true }
)

onUnmounted(() => {
  stop()
})

const processList = () => liveStats.value?.processList ?? []
const isSlow = (time: number) => time >= SLOW_THRESHOLD_SEC
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="show"
        class="fixed inset-0 z-[200] flex items-center justify-center p-4 theme-bg-content/90"
        @click.self="emit('close')"
      >
        <div
          class="w-full max-w-4xl max-h-[85vh] flex flex-col theme-bg-panel border theme-border rounded-lg shadow-xl overflow-hidden"
          @click.stop
        >
          <div class="flex items-center justify-between px-4 py-3 border-b theme-border">
            <h2 class="text-sm font-semibold theme-text">
              {{ t('monitor.title') }} - {{ connectionName || connectionId }}
            </h2>
            <button
              type="button"
              class="p-1.5 rounded hover:bg-[#37373d] theme-text transition-colors"
              @click="emit('close')"
            >
              <X :size="18" />
            </button>
          </div>

          <div class="flex-1 overflow-auto p-4 space-y-4">
            <p v-if="liveStats?.error" class="text-sm text-red-400">
              {{ liveStats.error }}
            </p>
            <template v-else>
              <div class="flex items-center gap-6">
                <div class="flex items-center gap-2">
                  <span class="text-xs theme-text opacity-80">{{ t('monitor.threadsConnected') }}:</span>
                  <span class="text-sm font-mono font-semibold text-[#1677ff]">{{ liveStats?.threadsConnected ?? '–' }}</span>
                </div>
              </div>

              <div>
                <h3 class="text-xs font-medium theme-text opacity-90 mb-2">{{ t('monitor.processList') }}</h3>
                <div class="border theme-border rounded overflow-x-auto">
                  <table class="w-full text-xs border-collapse">
                    <thead>
                      <tr class="theme-bg-panel border-b theme-border">
                        <th class="text-left px-3 py-2 theme-text font-medium">{{ t('monitor.id') }}</th>
                        <th class="text-left px-3 py-2 theme-text font-medium">{{ t('monitor.user') }}</th>
                        <th class="text-left px-3 py-2 theme-text font-medium">{{ t('monitor.host') }}</th>
                        <th class="text-left px-3 py-2 theme-text font-medium">{{ t('monitor.db') }}</th>
                        <th class="text-left px-3 py-2 theme-text font-medium">{{ t('monitor.command') }}</th>
                        <th class="text-left px-3 py-2 theme-text font-medium">{{ t('monitor.time') }}</th>
                        <th class="text-left px-3 py-2 theme-text font-medium">{{ t('monitor.state') }}</th>
                        <th class="text-left px-3 py-2 theme-text font-medium max-w-[200px] truncate">{{ t('monitor.info') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-if="processList().length === 0" class="theme-text opacity-70">
                        <td colspan="8" class="px-3 py-4 text-center">{{ t('monitor.noData') }}</td>
                      </tr>
                      <tr
                        v-for="p in processList()"
                        :key="p.id"
                        class="border-b theme-border hover:bg-[#2a2a2e]"
                        :class="{ 'bg-red-500/10': isSlow(p.time) }"
                      >
                        <td class="px-3 py-2 font-mono">{{ p.id }}</td>
                        <td class="px-3 py-2">{{ p.user }}</td>
                        <td class="px-3 py-2 font-mono text-[11px]">{{ p.host }}</td>
                        <td class="px-3 py-2">{{ p.db || '–' }}</td>
                        <td class="px-3 py-2">{{ p.command }}</td>
                        <td class="px-3 py-2 font-mono" :class="{ 'text-red-400 font-semibold': isSlow(p.time) }">
                          {{ p.time }}
                          <span v-if="isSlow(p.time)" class="ml-1 text-[10px] text-red-400">({{ t('monitor.slowQuery') }})</span>
                        </td>
                        <td class="px-3 py-2 opacity-90">{{ p.state || '–' }}</td>
                        <td class="px-3 py-2 font-mono text-[11px] max-w-[200px] truncate" :title="p.info">{{ p.info || '–' }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </template>
          </div>

          <div class="px-4 py-2 border-t theme-border flex justify-end">
            <button
              type="button"
              class="text-xs px-3 py-1.5 rounded bg-[#1677ff] hover:bg-[#4096ff] text-white transition-colors"
              @click="emit('close')"
            >
              {{ t('monitor.close') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
