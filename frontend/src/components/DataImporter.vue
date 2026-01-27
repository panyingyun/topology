<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Upload, X, CheckCircle, AlertCircle, FileText, Database } from 'lucide-vue-next'
import { importService } from '../services/importService'
import type { ImportPreview, ImportFormat, ImportResult } from '../types'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  connectionId: string
  database: string
  tableName: string
  /** Tab/session id for per-tab DB isolation */
  sessionId?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success', result: ImportResult): void
}>()

const selectedFile = ref<File | null>(null)
const filePath = ref('')
const importFormat = ref<ImportFormat>('csv')
const preview = ref<ImportPreview | null>(null)
const columnMapping = ref<Record<string, string>>({})
const isPreviewing = ref(false)
const isImporting = ref(false)
const importResult = ref<ImportResult | null>(null)
const step = ref<'select' | 'preview' | 'mapping' | 'importing' | 'result'>('select')

const handleFileSelect = async (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    selectedFile.value = target.files[0]
    const fileName = target.files[0].name.toLowerCase()
    if (fileName.endsWith('.csv')) {
      importFormat.value = 'csv'
    } else if (fileName.endsWith('.json')) {
      importFormat.value = 'json'
    }
    // For Wails, we need to use file path
    // In a real implementation, we'd need to handle file upload
    filePath.value = target.files[0].name
    step.value = 'preview'
    await loadPreview()
  }
}

const loadPreview = async () => {
  if (!filePath.value) return
  isPreviewing.value = true
  try {
    // Note: In Wails, we need to handle file selection differently
    // For now, we'll use a placeholder path
    preview.value = await importService.previewImport(filePath.value, importFormat.value)
    if (preview.value.error) {
      alert(t('importer.previewFailed') + ': ' + preview.value.error)
      step.value = 'select'
    } else {
      // Initialize column mapping (file column -> table column)
      columnMapping.value = {}
      preview.value.columns.forEach((col) => {
        columnMapping.value[col] = col // Default: same name
      })
      step.value = 'mapping'
    }
  } catch (error) {
    console.error('Failed to load preview:', error)
    alert(t('importer.previewFailed') + ': ' + (error instanceof Error ? error.message : 'Unknown error'))
    step.value = 'select'
  } finally {
    isPreviewing.value = false
  }
}

const handleImport = async () => {
  if (!filePath.value) return
  isImporting.value = true
  step.value = 'importing'
  try {
    const result = await importService.importData(
      props.connectionId,
      props.database,
      props.tableName,
      filePath.value,
      importFormat.value,
      columnMapping.value,
      props.sessionId ?? ''
    )
    importResult.value = result
    if (result.success) {
      step.value = 'result'
      emit('success', result)
    } else {
      alert(t('importer.importFailed') + ': ' + (result.error || 'Unknown error'))
      step.value = 'mapping'
    }
  } catch (error) {
    console.error('Failed to import:', error)
    alert(t('importer.importFailed') + ': ' + (error instanceof Error ? error.message : 'Unknown error'))
    step.value = 'mapping'
  } finally {
    isImporting.value = false
  }
}

const handleClose = () => {
  selectedFile.value = null
  filePath.value = ''
  preview.value = null
  columnMapping.value = {}
  importResult.value = null
  step.value = 'select'
  emit('close')
}

const availableTableColumns = computed(() => {
  // In a real implementation, we'd fetch this from the backend
  // For now, we'll use the preview columns as a placeholder
  return preview.value?.columns || []
})
</script>

<template>
  <Transition name="fade">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      @click.self="handleClose"
    >
      <div class="theme-bg-panel rounded-lg border theme-border w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
        <!-- Header -->
        <div class="px-6 py-4 border-b theme-border flex items-center justify-between">
          <h2 class="text-lg font-semibold theme-text flex items-center gap-2">
            <Database :size="20" class="text-[#1677ff]" />
            {{ t('importer.title', { table: tableName }) }}
          </h2>
          <button
            @click="handleClose"
            class="theme-text-muted-hover transition-colors"
          >
            <X :size="20" />
          </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <!-- Step 1: File Selection -->
          <div v-if="step === 'select'" class="space-y-4">
            <div>
              <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('importer.selectFormat') }}</label>
              <div class="flex gap-4">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    v-model="importFormat"
                    type="radio"
                    value="csv"
                    class="w-4 h-4 text-[#1677ff]"
                  />
                  <span class="text-sm theme-text">CSV</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input
                    v-model="importFormat"
                    type="radio"
                    value="json"
                    class="w-4 h-4 text-[#1677ff]"
                  />
                  <span class="text-sm theme-text">JSON</span>
                </label>
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('importer.selectFile') }}</label>
              <div class="border-2 border-dashed theme-border-strong rounded-lg p-8 text-center hover:border-[#1677ff] transition-colors">
                <input
                  type="file"
                  :accept="importFormat === 'csv' ? '.csv' : '.json'"
                  @change="handleFileSelect"
                  class="hidden"
                  id="file-input"
                />
                <label
                  for="file-input"
                  class="cursor-pointer flex flex-col items-center gap-2"
                >
                  <Upload :size="32" class="theme-text-muted" />
                  <span class="text-sm theme-text">{{ t('importer.dragDrop') }}</span>
                  <span class="text-xs theme-text-muted opacity-80">{{ t('importer.supportedFormat', { format: importFormat.toUpperCase() }) }}</span>
                </label>
              </div>
            </div>
          </div>

          <!-- Step 2: Preview -->
          <div v-if="step === 'preview' && isPreviewing" class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="w-8 h-8 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
              <p class="text-sm theme-text-muted">{{ t('common.loading') }}</p>
            </div>
          </div>

          <!-- Step 3: Column Mapping -->
          <div v-if="step === 'mapping' && preview" class="space-y-4">
            <div>
              <h3 class="text-sm font-semibold theme-text mb-2">{{ t('importer.preview') }}</h3>
              <div class="theme-bg-content rounded border theme-border overflow-x-auto">
                <table class="w-full text-xs">
                  <thead class="theme-bg-footer sticky top-0">
                    <tr>
                      <th
                        v-for="col in preview.columns"
                        :key="col"
                        class="px-3 py-2 text-left theme-text border-r theme-border"
                      >
                        {{ col }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="(row, idx) in preview.rows"
                      :key="idx"
                      class="border-b theme-border theme-bg-hover"
                    >
                      <td
                        v-for="col in preview.columns"
                        :key="col"
                        class="px-3 py-2 theme-text-muted border-r theme-border font-mono"
                      >
                        {{ row[col] }}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div>
              <h3 class="text-sm font-semibold theme-text mb-2">{{ t('importer.columnMapping') }}</h3>
              <div class="space-y-2">
                <div
                  v-for="fileCol in preview.columns"
                  :key="fileCol"
                  class="flex items-center gap-3 p-3 theme-bg-content rounded border theme-border"
                >
                  <span class="text-xs theme-text-muted w-32 truncate">{{ fileCol }}</span>
                  <span class="theme-text-muted opacity-80">→</span>
                  <input
                    v-model="columnMapping[fileCol]"
                    type="text"
                    :placeholder="fileCol"
                    class="flex-1 theme-input rounded px-3 py-1.5 text-xs"
                  />
                </div>
              </div>
            </div>

            <div class="flex items-center justify-end gap-3 pt-4">
              <button
                @click="step = 'select'"
                class="px-4 py-2 rounded text-xs font-semibold theme-bg-input theme-bg-input-hover theme-text transition-colors"
              >
                {{ t('importer.back') }}
              </button>
              <button
                @click="handleImport"
                class="px-6 py-2 rounded text-xs font-semibold bg-[#1677ff] hover:bg-[#4096ff] text-white transition-colors"
              >
                {{ t('importer.startImport') }}
              </button>
            </div>
          </div>

          <!-- Step 4: Importing -->
          <div v-if="step === 'importing'" class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="w-8 h-8 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
              <p class="text-sm theme-text-muted">{{ t('importer.importing') }}</p>
            </div>
          </div>

          <!-- Step 5: Result -->
          <div v-if="step === 'result' && importResult" class="space-y-4">
            <div
              :class="[
                'p-4 rounded-lg border',
                importResult.success
                  ? 'bg-green-500/10 border-green-500/50'
                  : 'bg-red-500/10 border-red-500/50'
              ]"
            >
              <div class="flex items-center gap-2 mb-2">
                <component
                  :is="importResult.success ? CheckCircle : AlertCircle"
                  :size="20"
                  :class="importResult.success ? 'text-green-500' : 'text-red-500'"
                />
                <span
                  :class="[
                    'text-sm font-semibold',
                    importResult.success ? 'text-green-400' : 'text-red-400'
                  ]"
                >
                  {{ importResult.success ? t('importer.importSuccess') : t('importer.importFailed') }}
                </span>
              </div>
              <div v-if="importResult.success" class="text-xs theme-text space-y-1">
                <p>{{ t('importer.imported', { inserted: importResult.inserted, total: importResult.totalRows }) }}</p>
              </div>
              <div v-else class="text-xs text-red-400">
                <p>{{ importResult.error }}</p>
              </div>
            </div>

            <div class="flex items-center justify-end gap-3 pt-4">
              <button
                @click="handleClose"
                class="px-6 py-2 rounded text-xs font-semibold bg-[#1677ff] hover:bg-[#4096ff] text-white transition-colors"
              >
                {{ t('importer.complete') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
