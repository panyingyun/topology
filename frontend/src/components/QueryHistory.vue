<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { History, Search, X, Clock, CheckCircle, XCircle, Trash2 } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { getLocale } from '../locales'
import { historyService } from '../services/historyService'
import type { QueryHistory } from '../types'

const { t } = useI18n()

const props = defineProps<{
  connectionId?: string
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'select', sql: string): void
  (e: 'close'): void
}>()

const history = ref<QueryHistory[]>([])
const searchTerm = ref('')
const isLoading = ref(false)

const loadHistory = async () => {
  isLoading.value = true
  try {
    history.value = await historyService.getQueryHistory(
      props.connectionId || '',
      searchTerm.value,
      50
    )
  } catch (error) {
    console.error('Failed to load history:', error)
  } finally {
    isLoading.value = false
  }
}

const handleSelect = (item: QueryHistory) => {
  emit('select', item.sql)
  emit('close')
}

const handleClear = async () => {
  if (confirm(t('history.clearConfirm'))) {
    try {
      await historyService.clearQueryHistory()
      await loadHistory()
    } catch (error) {
      console.error('Failed to clear history:', error)
    }
  }
}

const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return t('history.timeAgo.justNow')
  if (minutes < 60) return t('history.timeAgo.minutesAgo', { n: minutes })
  if (hours < 24) return t('history.timeAgo.hoursAgo', { n: hours })
  if (days < 7) return t('history.timeAgo.daysAgo', { n: days })
  const locale = getLocale()
  return date.toLocaleDateString(locale === 'zh-CN' ? 'zh-CN' : 'en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const filteredHistory = computed(() => {
  if (!searchTerm.value) return history.value
  const term = searchTerm.value.toLowerCase()
  return history.value.filter((h) => h.sql.toLowerCase().includes(term))
})

watch(() => props.show, (newVal) => {
  if (newVal) {
    loadHistory()
  }
})

watch(searchTerm, () => {
  loadHistory()
})

onMounted(() => {
  if (props.show) {
    loadHistory()
  }
})
</script>

<template>
  <Transition name="slide">
    <div
      v-if="show"
      class="fixed right-0 top-0 bottom-0 w-96 theme-bg-panel border-l theme-border z-50 flex flex-col shadow-2xl"
    >
      <!-- Header -->
      <div class="h-12 flex items-center justify-between px-4 border-b theme-border theme-bg-footer">
        <div class="flex items-center gap-2">
          <History :size="16" class="text-[#1677ff]" />
          <span class="text-sm font-semibold theme-text">{{ t('history.title') }}</span>
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="handleClear"
            class="p-1.5 theme-bg-hover rounded transition-colors"
            :title="t('history.clear')"
          >
            <Trash2 :size="14" class="theme-text-muted" />
          </button>
          <button
            @click="emit('close')"
            class="p-1.5 theme-bg-hover rounded transition-colors"
          >
            <X :size="16" class="theme-text-muted" />
          </button>
        </div>
      </div>

      <!-- Search -->
      <div class="p-3 border-b theme-border">
        <div class="relative">
          <Search :size="14" class="absolute left-2 top-1/2 -translate-y-1/2 theme-text-muted" />
          <input
            v-model="searchTerm"
            type="text"
            :placeholder="t('history.search')"
            class="w-full theme-input text-xs pl-8 pr-3 py-2 rounded border transition-all"
          />
        </div>
      </div>

      <!-- History List -->
      <div class="flex-1 overflow-y-auto custom-scrollbar">
        <div v-if="isLoading" class="flex items-center justify-center h-32">
          <div class="w-6 h-6 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin"></div>
        </div>
        <div v-else-if="filteredHistory.length === 0" class="flex flex-col items-center justify-center h-32 theme-text-muted">
          <History :size="32" class="mb-2 opacity-50" />
          <p class="text-xs">{{ t('history.noHistory') }}</p>
        </div>
        <div v-else class="divide-y divide-[var(--border)]">
          <button
            v-for="item in filteredHistory"
            :key="item.id"
            @click="handleSelect(item)"
            class="w-full text-left px-4 py-3 theme-bg-hover transition-colors group"
          >
            <div class="flex items-start gap-2 mb-1">
              <component
                :is="item.success ? CheckCircle : XCircle"
                :size="12"
                :class="item.success ? 'text-green-500 mt-0.5' : 'text-red-500 mt-0.5'"
              />
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1">
                  <span class="text-xs theme-text-muted flex items-center gap-1">
                    <Clock :size="10" />
                    {{ formatTime(item.executedAt) }}
                  </span>
                  <span v-if="item.duration" class="text-xs theme-text-muted opacity-80">
                    {{ item.duration }}ms
                  </span>
                  <span v-if="item.rowCount !== undefined" class="text-xs theme-text-muted opacity-80">
                    {{ item.rowCount }} 行
                  </span>
                </div>
                <p class="text-xs theme-text font-mono line-clamp-2 opacity-90 group-hover:opacity-100 transition-colors">
                  {{ item.sql }}
                </p>
              </div>
            </div>
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.slide-enter-active,
.slide-leave-active {
  transition: transform 0.3s ease;
}

.slide-enter-from {
  transform: translateX(100%);
}

.slide-leave-to {
  transform: translateX(100%);
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
