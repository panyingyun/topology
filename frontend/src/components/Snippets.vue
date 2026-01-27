<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { FileCode, Search, X, Trash2 } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { snippetService } from '../services/snippetService'
import type { Snippet } from '../types'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  /** Current editor SQL to save as snippet */
  currentSql?: string
  /** Increment to trigger list refresh (e.g. after save) */
  refreshTrigger?: number
}>()

const emit = defineEmits<{
  (e: 'select', sql: string): void
  (e: 'save', alias: string): void
  (e: 'close'): void
}>()

const snippets = ref<Snippet[]>([])
const searchTerm = ref('')
const isLoading = ref(false)
const aliasToSave = ref('')
const isSaving = ref(false)

const loadSnippets = async () => {
  isLoading.value = true
  try {
    snippets.value = await snippetService.getSnippets()
  } catch (error) {
    console.error('Failed to load snippets:', error)
  } finally {
    isLoading.value = false
  }
}

const filteredSnippets = computed(() => {
  if (!searchTerm.value) return snippets.value
  const term = searchTerm.value.toLowerCase()
  return snippets.value.filter((s) => s.alias.toLowerCase().includes(term) || s.sql.toLowerCase().includes(term))
})

const handleSelect = (item: Snippet) => {
  emit('select', item.sql)
  emit('close')
}

const handleDelete = async (e: Event, id: string) => {
  e.stopPropagation()
  if (!confirm(t('common.confirm') + ' ' + t('snippets.delete') + '?')) return
  try {
    await snippetService.deleteSnippet(id)
    await loadSnippets()
  } catch (error) {
    console.error('Failed to delete snippet:', error)
  }
}

const handleSaveCurrent = () => {
  const alias = aliasToSave.value.trim()
  if (!alias) {
    alert(t('snippets.aliasRequired'))
    return
  }
  isSaving.value = true
  emit('save', alias)
  aliasToSave.value = ''
  isSaving.value = false
}

watch(() => props.show, (newVal) => {
  if (newVal) loadSnippets()
})

watch(() => props.refreshTrigger, () => {
  if (props.show) loadSnippets()
})
</script>

<template>
  <Transition name="slide">
    <div
      v-if="show"
      class="fixed right-0 top-0 bottom-0 w-96 theme-bg-panel border-l theme-border z-50 flex flex-col shadow-2xl"
    >
      <div class="h-12 flex items-center justify-between px-4 border-b theme-border theme-bg-footer">
        <div class="flex items-center gap-2">
          <FileCode :size="16" class="text-[#1677ff]" />
          <span class="text-sm font-semibold theme-text">{{ t('snippets.title') }}</span>
        </div>
        <button
          @click="emit('close')"
          class="p-1.5 theme-bg-hover rounded transition-colors"
        >
          <X :size="16" class="theme-text-muted" />
        </button>
      </div>

      <!-- Save current SQL as snippet -->
      <div v-if="currentSql?.trim()" class="p-3 border-b theme-border space-y-2">
        <div class="text-xs theme-text-muted">{{ t('snippets.saveCurrent') }}</div>
        <div class="flex gap-2">
          <input
            v-model="aliasToSave"
            type="text"
            :placeholder="t('snippets.alias')"
            class="flex-1 theme-input text-xs px-3 py-2 rounded border transition-all"
            @keyup.enter="handleSaveCurrent"
          />
          <button
            @click="handleSaveCurrent"
            :disabled="isSaving || !aliasToSave.trim()"
            class="px-3 py-2 rounded text-xs font-medium bg-[#1677ff] hover:bg-[#4096ff] disabled:opacity-50 text-white"
          >
            {{ t('snippets.save') }}
          </button>
        </div>
      </div>

      <div class="p-3 border-b theme-border">
        <div class="relative">
          <Search :size="14" class="absolute left-2 top-1/2 -translate-y-1/2 theme-text-muted" />
          <input
            v-model="searchTerm"
            type="text"
            :placeholder="t('snippets.search')"
            class="w-full theme-input text-xs pl-8 pr-3 py-2 rounded border transition-all"
          />
        </div>
      </div>

      <div class="flex-1 overflow-y-auto custom-scrollbar">
        <div v-if="isLoading" class="flex items-center justify-center h-32">
          <div class="w-6 h-6 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin"></div>
        </div>
        <div v-else-if="filteredSnippets.length === 0" class="flex flex-col items-center justify-center h-32 theme-text-muted">
          <FileCode :size="32" class="mb-2 opacity-50" />
          <p class="text-xs">{{ t('snippets.noSnippets') }}</p>
        </div>
        <div v-else class="divide-y divide-[var(--border)]">
          <div
            v-for="item in filteredSnippets"
            :key="item.id"
            role="button"
            tabindex="0"
            class="w-full text-left px-4 py-3 theme-bg-hover transition-colors group flex items-start gap-2 cursor-pointer"
            @click="handleSelect(item)"
            @keydown.enter.prevent="handleSelect(item)"
            @keydown.space.prevent="handleSelect(item)"
          >
            <div class="flex-1 min-w-0">
              <div class="text-sm font-medium text-[#1677ff] mb-1">{{ item.alias }}</div>
              <p class="text-xs theme-text-muted font-mono line-clamp-2 opacity-90 group-hover:opacity-100">
                {{ item.sql }}
              </p>
            </div>
            <button
              type="button"
              @click.stop="handleDelete($event, item.id)"
              class="p-1 opacity-0 group-hover:opacity-100 theme-bg-hover rounded transition-all flex-shrink-0"
              :title="t('snippets.delete')"
            >
              <Trash2 :size="14" class="theme-text-muted hover:text-red-400" />
            </button>
          </div>
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
