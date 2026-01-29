<script setup lang="ts">
import { ref, watch } from 'vue'
import { Key, X, Copy } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { queryService } from '../services/queryService'
import type { IndexSuggestion } from '../types'

const { t } = useI18n()
const message = useMessage()

const props = defineProps<{
  show: boolean
  connectionId: string
  tabId: string
  sql: string
  driver?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const suggestions = ref<IndexSuggestion[]>([])
const error = ref<string | null>(null)
const isLoading = ref(false)

const load = async () => {
  if (!props.sql.trim() || !props.connectionId) return
  isLoading.value = true
  suggestions.value = []
  error.value = null
  try {
    const res = await queryService.getIndexSuggestions(
      props.connectionId,
      props.tabId || '',
      props.sql
    )
    suggestions.value = res.suggestions ?? []
    error.value = res.error ?? null
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Unknown error'
  } finally {
    isLoading.value = false
  }
}

watch(
  () => [props.show, props.sql, props.connectionId],
  () => {
    if (props.show && props.sql?.trim() && props.connectionId) load()
  }
)

async function copySql(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    message.success(t('indexSuggestions.copied'))
  } catch {}
}
</script>

<template>
  <Transition name="fade">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <div class="theme-bg-panel rounded-lg border theme-border w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b theme-border flex items-center justify-between">
          <h2 class="text-lg font-semibold theme-text flex items-center gap-2">
            <Key :size="20" class="text-amber-500" />
            {{ t('indexSuggestions.title') }}
          </h2>
          <button
            @click="emit('close')"
            class="theme-text-muted-hover transition-colors p-1"
          >
            <X :size="20" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <div v-if="isLoading" class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="w-8 h-8 border-2 border-amber-500 border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
              <p class="text-sm theme-text-muted">{{ t('indexSuggestions.loading') }}</p>
            </div>
          </div>

          <div v-else-if="error" class="p-4 bg-red-500/10 rounded border border-red-500/50">
            <p class="text-sm text-red-400">{{ error }}</p>
            <p v-if="error?.includes('MySQL and PostgreSQL')" class="text-xs theme-text-muted mt-2">{{ t('indexSuggestions.unsupportedDriver') }}</p>
          </div>

          <div v-else-if="suggestions.length === 0" class="text-center py-12 theme-text-muted text-sm">
            {{ t('indexSuggestions.noSuggestions') }}
          </div>

          <div v-else class="space-y-4">
            <div
              v-for="(s, i) in suggestions"
              :key="i"
              class="p-4 rounded-lg border theme-border theme-bg-panel"
            >
              <p class="text-xs font-medium text-amber-400 mb-2">{{ s.reason }}</p>
              <div class="flex items-start gap-2">
                <pre class="flex-1 text-xs font-mono theme-text overflow-x-auto rounded bg-black/20 px-3 py-2 break-all">{{ s.createIndex }}</pre>
                <button
                  v-if="!s.createIndex.startsWith('--')"
                  type="button"
                  class="flex-shrink-0 p-2 rounded theme-bg-input theme-bg-input-hover theme-text"
                  :aria-label="t('indexSuggestions.copy')"
                  @click="copySql(s.createIndex)"
                >
                  <Copy :size="14" />
                </button>
              </div>
              <p v-if="s.columns?.length" class="text-[10px] theme-text-muted mt-2">
                {{ s.columns.join(', ') }}
              </p>
            </div>
          </div>
        </div>

        <div class="px-6 py-4 border-t theme-border flex justify-end theme-bg-footer">
          <button
            @click="emit('close')"
            class="px-6 py-2 rounded text-xs font-semibold bg-[#1677ff] hover:bg-[#4096ff] text-white"
          >
            {{ t('common.close') }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
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
