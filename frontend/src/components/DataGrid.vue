<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { VxeGrid } from 'vxe-table'
import type { VxeGridProps } from 'vxe-table'
import { Download, ChevronDown, Filter } from 'lucide-vue-next'
import type { QueryResult, UpdateRecord, ExportFormat } from '../types'

const { t } = useI18n()
const message = useMessage()

const props = withDefaults(
  defineProps<{
    data: QueryResult
    queryText?: string
    /** When true, disable cell editing (e.g. for query results). Table DataViewer keeps editable. */
    readonly?: boolean
    /** When true, use light table theme + border-only selection (for query results). */
    useLightTable?: boolean
  }>(),
  { readonly: false, useLightTable: false }
)

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
  rowConfig: { isCurrent: true, isHover: false },
  scrollY: { enabled: true, gt: 20 },
  editConfig: { trigger: 'dblclick', mode: 'cell' },
  filterConfig: { remote: false },
  columns: [],
  data: [],
})

watch(
  () => [props.data, props.readonly] as const,
  (tuple) => {
    const [data, ro] = tuple
    if (!data?.columns) return
    const isReadonly = !!ro
    gridOptions.value.editConfig = isReadonly
      ? { enabled: false }
      : { trigger: 'dblclick', mode: 'cell' }
    gridOptions.value.columns = data.columns.map((col: string) => {
      const colDef: Record<string, unknown> = {
        field: col,
        title: col,
        width: 150,
        filters: [
          { label: '包含', value: 'contains' },
          { label: '等于', value: 'equals' },
          { label: '不为空', value: 'notEmpty' },
          { label: '为空', value: 'isEmpty' },
        ],
        filterRender: { name: 'input' },
        formatter: ({ cellValue }: { cellValue: unknown }) => formatCellDisplay(cellValue),
      }
      if (!isReadonly) (colDef as any).editRender = { name: 'input' }
      return colDef
    })
    gridOptions.value.data = data.rows
    pendingChanges.value = 0
  },
  { immediate: true }
)

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

function cellValueToString(val: unknown): string {
  if (val == null) return ''
  if (typeof val === 'object') return JSON.stringify(val)
  return String(val)
}

const CELL_TRUNCATE_LEN = 30
const CELL_TRUNCATE_SHOW = 30

function formatCellDisplay(val: unknown): string {
  const s = cellValueToString(val)
  if (s.length > CELL_TRUNCATE_LEN) return s.slice(0, CELL_TRUNCATE_SHOW) + '...'
  return s
}

async function handleCellDblclick({ row, column }: { row?: Record<string, unknown>; column?: { field?: string } }) {
  if (!props.useLightTable || !row || !column?.field) return
  const val = row[column.field]
  const str = cellValueToString(val)
  try {
    await navigator.clipboard.writeText(str)
    message.success(t('dataGrid.copiedToClipboard'))
  } catch (e) {
    message.error(t('common.error') + ': ' + (e instanceof Error ? e.message : 'Copy failed'))
  }
}

function buildUpdatesFromRecordset(recordset: { updateRecords?: any[] }, columns: string[]): UpdateRecord[] {
  const updates: UpdateRecord[] = []
  const list = recordset.updateRecords || []
  for (const record of list) {
    const origin = record._X_ORIGIN_DATA as Record<string, unknown> | undefined
    if (!origin) continue
    for (const col of columns) {
      const oldVal = origin[col]
      const newVal = record[col]
      if (oldVal !== newVal && JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
        updates.push({
          rowIndex: record._X_ROW_KEY ?? 0,
          column: col,
          oldValue: oldVal,
          newValue: newVal,
        })
      }
    }
  }
  return updates
}

const commitChanges = () => {
  if (!gridRef.value || !props.data?.columns) return
  try {
    const recordset = gridRef.value.getRecordset()
    if (recordset?.updateRecords?.length) {
      const updates = buildUpdatesFromRecordset(recordset, props.data.columns)
      if (updates.length) {
        emit('update', updates)
      }
      pendingChanges.value = 0
      gridRef.value.reloadRow(recordset.updateRecords, null)
    }
  } catch (error) {
    console.error('Error committing changes:', error)
  }
}

const rollbackChanges = () => {
  if (!gridRef.value || !props.data?.rows) return
  try {
    const recordset = gridRef.value.getRecordset()
    if (recordset?.updateRecords?.length) {
      const rows = props.data.rows.map((r: Record<string, unknown>) => ({ ...r }))
      gridOptions.value.data = rows
      if (typeof gridRef.value.reloadData === 'function') {
        gridRef.value.reloadData(rows)
      }
      pendingChanges.value = 0
    }
  } catch (error) {
    console.error('Error rolling back:', error)
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
  <div class="flex flex-col h-full theme-bg-content">
    <div class="h-10 flex items-center justify-between px-4 theme-bg-panel border-b theme-border">
      <div class="flex items-center gap-4 text-xs theme-text-muted">
        <span>{{ t('dataGrid.rows') }}: {{ data.rowCount.toLocaleString() }}</span>
        <span v-if="!props.readonly && pendingChanges > 0" class="text-yellow-400">
          {{ pendingChanges }} {{ t('dataGrid.pendingChanges') }}
        </span>
      </div>
      <div class="flex items-center gap-2">
        <button
          v-if="!props.readonly && pendingChanges > 0"
          @click="commitChanges"
          class="px-3 py-1 bg-green-600 hover:bg-green-500 text-white text-xs rounded transition-colors"
        >
          {{ t('dataGrid.commit') }}
        </button>
        <button
          v-if="!props.readonly && pendingChanges > 0"
          @click="rollbackChanges"
          class="px-3 py-1 theme-bg-input theme-bg-input-hover theme-text text-xs rounded transition-colors"
        >
          {{ t('dataGrid.rollback') }}
        </button>
        <div class="relative" ref="exportMenuRef">
          <button
            @click.stop="showExportMenu = !showExportMenu"
            class="flex items-center gap-1 px-3 py-1 theme-bg-input theme-bg-input-hover theme-text text-xs rounded transition-colors"
          >
            <Download :size="12" />
            {{ t('dataGrid.export') }}
            <ChevronDown :size="12" />
          </button>
          <Transition name="fade">
            <div
              v-if="showExportMenu"
              class="absolute right-0 top-full mt-1 theme-bg-panel border theme-border rounded shadow-lg py-1 min-w-[140px] z-50"
              @click.stop
            >
              <button
                v-for="format in exportFormats"
                :key="format.value"
                @click="handleExport(format.value)"
                class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors"
              >
                {{ format.label }}
              </button>
            </div>
          </Transition>
        </div>
      </div>
    </div>
    <div
      class="flex-1 overflow-hidden"
      :class="{ 'data-grid-light': props.useLightTable }"
      style="height: calc(100% - 40px);"
    >
      <VxeGrid
        ref="gridRef"
        v-bind="gridOptions"
        @edit-closed="handleEditClosed"
        @cell-dblclick="handleCellDblclick"
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
/* 表格：浅色主题；仅常态与选中态，选中用左边框 */
.data-grid-light {
  --table-bg: #fafafa;
  --table-header-bg: #f0f0f0;
  --table-border: #e0e0e0;
  --table-text: #1a1a1a;
  --table-border-current: #1677ff;
}

.data-grid-light .vxe-table {
  background-color: var(--table-bg);
  color: var(--table-text);
}

.data-grid-light .vxe-header--column {
  background-color: var(--table-header-bg);
  color: var(--table-text);
  border-color: var(--table-border);
}

.data-grid-light .vxe-body--row {
  background-color: var(--table-bg);
}

.data-grid-light .vxe-body--row.row--current {
  background-color: var(--table-bg) !important;
  box-shadow: inset 3px 0 0 var(--table-border-current);
}
.data-grid-light .vxe-body--row.row--current > .vxe-body--column {
  background-color: transparent !important;
}

.data-grid-light .vxe-body--column {
  background-color: var(--table-bg);
  border-color: var(--table-border);
  color: var(--table-text);
}

.data-grid-light .vxe-body--column.col--update,
.data-grid-light .vxe-body--column.col--edit {
  background-color: #fef9e7 !important;
  border: 1px solid #e6a23c;
}

.data-grid-light .vxe-body--column.col--update:hover,
.data-grid-light .vxe-body--column.col--edit:hover {
  background-color: #fdf6dc !important;
}

.data-grid-light .vxe-body--row.row--update {
  background-color: #fef9e7 !important;
}

.data-grid-light .vxe-body--row.row--update:hover {
  background-color: #fdf6dc !important;
}

.data-grid-light .vxe-filter--wrapper,
.data-grid-light .vxe-filter--panel {
  background-color: var(--table-header-bg);
  border-color: var(--table-border);
}

.data-grid-light .vxe-filter--panel .vxe-input {
  background-color: #fff;
  border-color: var(--table-border);
  color: var(--table-text);
}
</style>
