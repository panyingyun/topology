<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { VxeGrid } from 'vxe-table'
import type { VxeGridProps } from 'vxe-table'
import { Download, ChevronDown, Filter } from 'lucide-vue-next'
import type { QueryResult, UpdateRecord, ExportFormat } from '../types'

const { t } = useI18n()

const props = defineProps<{
  data: QueryResult
  queryText?: string
}>()

const emit = defineEmits<{
  (e: 'update', updates: UpdateRecord[]): void
  (e: 'export', format: ExportFormat): void
}>()

const showExportMenu = ref(false)
const pendingChanges = ref(0)
const gridRef = ref<any>()
const exportMenuRef = ref<HTMLElement | null>(null)

const gridOptions = ref<VxeGridProps>({
  border: true,
  height: '100%',
  columnConfig: { resizable: true },
  rowConfig: { isCurrent: true, isHover: true },
  scrollY: { enabled: true, gt: 50 },
  editConfig: { trigger: 'dblclick', mode: 'cell' },
  filterConfig: { remote: false },
  columns: [],
  data: [],
})

watch(() => props.data, (newData) => {
  if (newData && newData.columns) {
    gridOptions.value.columns = newData.columns.map(col => ({
      field: col,
      title: col,
      editRender: { name: 'input' },
      width: 150,
      filters: [
        { label: '包含', value: 'contains' },
        { label: '等于', value: 'equals' },
        { label: '不为空', value: 'notEmpty' },
        { label: '为空', value: 'isEmpty' },
      ],
      filterRender: { name: 'input' },
    }))
    gridOptions.value.data = newData.rows
    pendingChanges.value = 0
  }
}, { immediate: true })

const handleEditClosed = () => {
  if (gridRef.value) {
    try {
      const recordset = gridRef.value.getRecordset()
      if (recordset && recordset.updateRecords) {
        pendingChanges.value = recordset.updateRecords.length
      }
    } catch (error) {
      console.error('Error getting recordset:', error)
    }
  }
}

const saveChanges = () => {
  if (gridRef.value) {
    try {
      const recordset = gridRef.value.getRecordset()
      if (recordset && recordset.updateRecords) {
        const updates: UpdateRecord[] = recordset.updateRecords.map((record: any) => ({
          rowIndex: record._X_ROW_KEY || 0,
          column: '',
          oldValue: record._X_ORIGIN_DATA || record,
          newValue: record,
        }))
        emit('update', updates)
        pendingChanges.value = 0
        gridRef.value.reloadRow(recordset.updateRecords, null)
      }
    } catch (error) {
      console.error('Error saving changes:', error)
    }
  }
}

const handleExport = (format: ExportFormat) => {
  emit('export', format)
  showExportMenu.value = false
}

const exportFormats = computed(() => [
  { label: t('dataGrid.csv'), value: 'csv' as ExportFormat },
  { label: t('dataGrid.json'), value: 'json' as ExportFormat },
  { label: t('dataGrid.sql'), value: 'sql' as ExportFormat },
])

// Close export menu when clicking outside
const handleClickOutside = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (exportMenuRef.value && !exportMenuRef.value.contains(target)) {
    showExportMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="flex flex-col h-full bg-[#1e1e1e]">
    <div class="h-10 flex items-center justify-between px-4 bg-[#252526] border-b border-[#333]">
      <div class="flex items-center gap-4 text-xs text-gray-400">
        <span>{{ t('dataGrid.rows') }}: {{ data.rowCount.toLocaleString() }}</span>
        <span v-if="pendingChanges > 0" class="text-yellow-400">
          {{ pendingChanges }} {{ t('dataGrid.pendingChanges') }}
        </span>
      </div>
      <div class="flex items-center gap-2">
        <button
          v-if="pendingChanges > 0"
          @click="saveChanges"
          class="px-3 py-1 bg-green-600 hover:bg-green-500 text-white text-xs rounded transition-colors"
        >
          {{ t('dataGrid.saveChanges') }}
        </button>
        <div class="relative" ref="exportMenuRef">
          <button
            @click.stop="showExportMenu = !showExportMenu"
            class="flex items-center gap-1 px-3 py-1 bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 text-xs rounded transition-colors"
          >
            <Download :size="12" />
            {{ t('dataGrid.export') }}
            <ChevronDown :size="12" />
          </button>
          <Transition name="fade">
            <div
              v-if="showExportMenu"
              class="absolute right-0 top-full mt-1 bg-[#252526] border border-[#333] rounded shadow-lg py-1 min-w-[140px] z-50"
              @click.stop
            >
              <button
                v-for="format in exportFormats"
                :key="format.value"
                @click="handleExport(format.value)"
                class="w-full px-4 py-2 text-left text-xs text-gray-300 hover:bg-[#37373d] transition-colors"
              >
                {{ format.label }}
              </button>
            </div>
          </Transition>
        </div>
      </div>
    </div>
    <div class="flex-1 overflow-hidden" style="height: calc(100% - 40px);">
      <VxeGrid
        ref="gridRef"
        v-bind="gridOptions"
        @edit-closed="handleEditClosed"
        class="custom-scrollbar"
        style="height: 100%;"
      />
    </div>
  </div>
</template>


<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>

<style>
/* vxe-table 样式覆盖 */
.vxe-table {
  background-color: #1e1e1e;
  color: #d4d4d4;
}

.vxe-header--column {
  background-color: #252526;
  color: #cccccc;
}

.vxe-body--row {
  background-color: #1e1e1e;
}

.vxe-body--row.row--hover {
  background-color: #2a2d2e;
}

.vxe-body--column {
  border-color: #333;
}

/* 编辑后的单元格黄色背景 */
.vxe-body--column.col--update,
.vxe-body--column.col--edit {
  background-color: #5c4a00 !important;
}

.vxe-body--column.col--update:hover,
.vxe-body--column.col--edit:hover {
  background-color: #6d5500 !important;
}

/* vxe-table 修改后的行标记 */
.vxe-body--row.row--update {
  background-color: #2a2a1e !important;
}

.vxe-body--row.row--update:hover {
  background-color: #3a3a2e !important;
}

/* 筛选器样式 */
.vxe-filter--wrapper {
  background-color: #252526;
  border-color: #333;
}

.vxe-filter--panel {
  background-color: #252526;
  border-color: #333;
}

.vxe-filter--panel .vxe-input {
  background-color: #3c3c3c;
  border-color: #444;
  color: #d4d4d4;
}
</style>
