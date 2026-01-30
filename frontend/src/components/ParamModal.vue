<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Key } from 'lucide-vue-next'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  /** Named placeholders e.g. [:id, :name] */
  named: string[]
  /** Number of ? placeholders */
  positional: number
  /** Last used values per key (e.g. :id -> ['1'], ?1 -> ['foo']) */
  history?: Record<string, string[]>
}>()

const emit = defineEmits<{
  (e: 'execute', values: Record<string, string>): void
  (e: 'close'): void
}>()

const values = ref<Record<string, string>>({})

const placeholders = ref<{ key: string; label: string }[]>([])

watch(
  () => [props.show, props.named, props.positional] as const,
  ([show, named, pos]) => {
    if (!show) return
    const list: { key: string; label: string }[] = []
    named.forEach((n) => list.push({ key: n, label: n }))
    for (let i = 0; i < pos; i++) list.push({ key: `?${i + 1}`, label: `? ${i + 1}` })
    placeholders.value = list
    const next: Record<string, string> = {}
    list.forEach((p) => {
      const hist = props.history?.[p.key]
      next[p.key] = hist?.[0] ?? ''
    })
    values.value = next
  },
  { immediate: true }
)

function handleExecute() {
  emit('execute', { ...values.value })
}

function handleClose() {
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="show"
        class="fixed inset-0 z-[9999] flex items-center justify-center bg-black/50"
        @click.self="handleClose"
      >
        <div class="theme-bg-panel rounded-lg border theme-border shadow-xl w-full max-w-md p-4" @click.stop>
          <div class="flex items-center gap-2 mb-3">
            <Key :size="18" class="text-[#1677ff]" />
            <h3 class="text-sm font-semibold theme-text">{{ t('paramModal.title') }}</h3>
          </div>
          <p class="text-xs theme-text-muted mb-3">{{ t('paramModal.hint') }}</p>
          <div class="space-y-3 max-h-64 overflow-y-auto custom-scrollbar">
            <div v-for="p in placeholders" :key="p.key" class="flex flex-col gap-1">
              <label class="text-xs theme-text-muted">{{ p.label }}</label>
              <input
                v-model="values[p.key]"
                type="text"
                class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                :placeholder="t('paramModal.value')"
                @keydown.enter.prevent="handleExecute"
              />
              <div v-if="history?.[p.key]?.length" class="flex flex-wrap gap-1 mt-0.5">
                <button
                  v-for="(v, i) of (history?.[p.key] || []).slice(0, 5)"
                  :key="i"
                  type="button"
                  class="px-2 py-0.5 text-[10px] rounded theme-bg-input theme-text-muted hover:theme-bg-hover"
                  @click="values[p.key] = v"
                >
                  {{ String(v).slice(0, 20) }}{{ String(v).length > 20 ? '…' : '' }}
                </button>
              </div>
            </div>
          </div>
          <div class="flex justify-end gap-2 mt-4">
            <button
              type="button"
              class="px-3 py-1.5 rounded text-sm theme-bg-input theme-bg-input-hover theme-text"
              @click="handleClose"
            >
              {{ t('common.cancel') }}
            </button>
            <button
              type="button"
              class="px-3 py-1.5 rounded text-sm bg-green-600 hover:bg-green-500 text-white"
              @click="handleExecute"
            >
              {{ t('query.execute') }}
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
