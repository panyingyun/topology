<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { FileText, X, Download } from 'lucide-vue-next'
import { auditService } from '../services/auditService'
import type { AuditEntry } from '../services/auditService'

const { t } = useI18n()
const message = useMessage()

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const entries = ref<AuditEntry[]>([])
const loading = ref(false)
const opFilter = ref('')
const limit = 200

const ops = [
  { value: '', label: '—' },
  { value: 'query', label: 'query' },
  { value: 'table_update', label: 'table_update' },
  { value: 'table_delete', label: 'table_delete' },
  { value: 'table_insert', label: 'table_insert' },
  { value: 'table_import', label: 'table_import' },
  { value: 'export', label: 'export' },
  { value: 'backup', label: 'backup' },
  { value: 'restore', label: 'restore' },
]

async function load() {
  loading.value = true
  entries.value = []
  try {
    entries.value = await auditService.query(limit, '', opFilter.value)
  } catch {
    message.error(t('audit.loadFailed'))
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, opFilter.value],
  () => {
    if (props.show) load()
  }
)

function detailTrim(s: string, max: number = 60): string {
  if (!s) return ''
  return s.length <= max ? s : s.slice(0, max) + '…'
}

async function exportLog(format: 'json' | 'csv') {
  try {
    const content = await auditService.exportFormat(format)
    if (!content) {
      message.warning(t('audit.noEntries'))
      return
    }
    const blob = new Blob([content], {
      type: format === 'json' ? 'application/json' : 'text/csv;charset=utf-8',
    })
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `audit.${format}`
    a.click()
    URL.revokeObjectURL(a.href)
    message.success(t('audit.exported'))
  } catch {
    message.error(t('common.error'))
  }
}
</script>

<template>
  <Transition name="fade">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <div class="theme-bg-panel rounded-lg border theme-border w-full max-w-5xl max-h-[90vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b theme-border flex items-center justify-between">
          <h2 class="text-lg font-semibold theme-text flex items-center gap-2">
            <FileText :size="20" class="text-[#1677ff]" />
            {{ t('audit.title') }}
          </h2>
          <div class="flex items-center gap-2">
            <select
              v-model="opFilter"
              class="theme-bg-input theme-text rounded border theme-border px-2 py-1 text-xs"
            >
              <option v-for="o in ops" :key="o.value || 'all'" :value="o.value">
                {{ o.label || t('audit.allOps') }}
              </option>
            </select>
            <button
              class="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
              @click="exportLog('csv')"
            >
              <Download :size="14" />
              CSV
            </button>
            <button
              class="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
              @click="exportLog('json')"
            >
              <Download :size="14" />
              JSON
            </button>
            <button class="p-1.5 theme-text-muted-hover rounded" @click="emit('close')">
              <X :size="20" />
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-auto p-4">
          <div v-if="loading" class="flex justify-center py-12">
            <div class="w-8 h-8 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin" />
          </div>
          <div v-else-if="entries.length === 0" class="py-12 text-center theme-text-muted text-sm">
            {{ t('audit.noEntries') }}
          </div>
          <div v-else class="overflow-auto max-h-[60vh] rounded border theme-border">
            <table class="w-full text-xs border-collapse">
              <thead class="theme-bg-input sticky top-0">
                <tr>
                  <th class="text-left px-2 py-1.5 border-b theme-border font-medium">{{ t('audit.at') }}</th>
                  <th class="text-left px-2 py-1.5 border-b theme-border font-medium">{{ t('audit.op') }}</th>
                  <th class="text-left px-2 py-1.5 border-b theme-border font-medium">{{ t('audit.detail') }}</th>
                  <th class="text-left px-2 py-1.5 border-b theme-border font-medium">{{ t('audit.connectionId') }}</th>
                  <th class="text-left px-2 py-1.5 border-b theme-border font-medium">{{ t('audit.database') }}</th>
                  <th class="text-left px-2 py-1.5 border-b theme-border font-medium">{{ t('audit.table') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(e, i) in entries"
                  :key="i"
                  class="border-b theme-border theme-bg-hover"
                >
                  <td class="px-2 py-1.5 theme-text-muted whitespace-nowrap">{{ e.at }}</td>
                  <td class="px-2 py-1.5 font-mono">{{ e.op }}</td>
                  <td class="px-2 py-1.5 max-w-xs truncate" :title="e.detail">{{ detailTrim(e.detail) }}</td>
                  <td class="px-2 py-1.5 theme-text-muted">{{ e.connectionId || '—' }}</td>
                  <td class="px-2 py-1.5 theme-text-muted">{{ e.database || '—' }}</td>
                  <td class="px-2 py-1.5 theme-text-muted">{{ e.table || '—' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
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
