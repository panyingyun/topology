<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Upload } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import DataGrid from '../components/DataGrid.vue'
import DataImporter from '../components/DataImporter.vue'
import { dataService } from '../services/dataService'
import type { TableData, UpdateRecord, ExportFormat, QueryResult, ImportResult, TableSchema } from '../types'

const { t } = useI18n()
const message = useMessage()

const props = defineProps<{
  /** Tab id for per-tab DB session isolation */
  tabId?: string
  connectionId: string
  database: string
  tableName: string
  /** 由父组件设置以打开导入弹窗（如从表名右键菜单选择「导入」） */
  importTrigger?: { connectionId: string; database: string; tableName: string } | null
}>()

const emit = defineEmits<{
  (e: 'update', updates: UpdateRecord[]): void
  (e: 'clear-import-trigger'): void
}>()

const showImporter = ref(false)

const tableData = ref<TableData>({
  columns: [],
  rows: [],
  totalRows: 0,
  page: 1,
  pageSize: 100,
})

const schema = ref<TableSchema | null>(null)
const isLoading = ref(false)
const currentPage = ref(1)
const pageSize = ref(100)
const pageSizeOptions = [50, 100, 200, 500]
const totalPages = computed(() => Math.ceil(tableData.value.totalRows / pageSize.value) || 1)

const LOAD_TIMEOUT_MS = 15000

const loadTableData = async (page: number = 1) => {
  isLoading.value = true
  try {
    const offset = (page - 1) * pageSize.value
    const dataPromise = dataService.getTableData(
      props.connectionId,
      props.database,
      props.tableName,
      pageSize.value,
      offset,
      props.tabId ?? ''
    )
    const timeoutPromise = new Promise<TableData>((_, reject) =>
      setTimeout(() => reject(new Error('timeout')), LOAD_TIMEOUT_MS)
    )
    const data = await Promise.race([dataPromise, timeoutPromise])
    tableData.value = {
      columns: data?.columns ?? [],
      rows: data?.rows ?? [],
      totalRows: data?.totalRows ?? 0,
      page: data?.page ?? page,
      pageSize: data?.pageSize ?? pageSize.value,
    }
    currentPage.value = page
    try {
      const sch = await dataService.getTableSchema(
        props.connectionId,
        props.database,
        props.tableName,
        props.tabId ?? ''
      )
      schema.value = sch?.columns?.length ? sch : null
    } catch {
      schema.value = null
    }
    try {
      await fetchTxStatus()
    } catch {
      txActive.value = false
    }
  } catch (error) {
    console.error('Failed to load table data:', error)
    schema.value = null
    txActive.value = false
    tableData.value = {
      columns: [],
      rows: [],
      totalRows: 0,
      page: 1,
      pageSize: pageSize.value,
    }
    currentPage.value = 1
  } finally {
    isLoading.value = false
  }
}

const handlePageChange = (page: number) => {
  loadTableData(page)
}

const handleLoadMore = () => {
  if (currentPage.value * pageSize.value < tableData.value.totalRows) {
    loadTableData(currentPage.value + 1)
  }
}

const handlePageSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadTableData(1)
}

function validateUpdates(updates: UpdateRecord[]): boolean {
  const cols = schema.value?.columns ?? []
  const byName = Object.fromEntries(cols.map((c) => [c.name, c]))
  for (const u of updates) {
    const col = byName[u.column]
    if (col && !col.nullable && (u.newValue == null || u.newValue === '')) {
      message.error(t('table.validationNonNull', { column: u.column }))
      return false
    }
  }
  return true
}

const handleUpdate = async (updates: UpdateRecord[]) => {
  if (!updates.length) return
  if (!validateUpdates(updates)) return
  try {
    await dataService.updateTableData(
      props.connectionId,
      props.database,
      props.tableName,
      updates,
      props.tabId ?? ''
    )
    await loadTableData(currentPage.value)
    message.success(t('common.success'))
    emit('update', updates)
  } catch (error) {
    message.error(t('common.error') + ': ' + (error instanceof Error ? error.message : 'Update failed'))
  }
}

const handleExport = async (format: ExportFormat) => {
  try {
    const result = await dataService.exportData(
      props.connectionId,
      props.database,
      props.tableName,
      format,
      props.tabId ?? ''
    )
    if (result.success) {
      message.success(t('common.success') + ': ' + (result.filename ?? 'Export completed'))
    } else {
      message.error(t('common.error') + ': ' + (result.error ?? 'Export failed'))
    }
  } catch (error) {
    console.error('Export error:', error)
    message.error(t('common.error') + ': ' + (error instanceof Error ? error.message : 'Export failed'))
  }
}

const handleImportSuccess = (result: ImportResult) => {
  // Reload table data after successful import
  loadTableData(currentPage.value)
}

// Convert TableData to QueryResult format for DataGrid
const queryResult = computed<QueryResult>(() => ({
  columns: tableData.value.columns,
  rows: tableData.value.rows,
  rowCount: tableData.value.rows.length,
  totalRows: tableData.value.totalRows,
}))

const tableContext = computed(() =>
  props.connectionId && props.database && props.tableName
    ? {
        connectionId: props.connectionId,
        database: props.database,
        tableName: props.tableName,
        sessionId: props.tabId ?? '',
      }
    : null
)

const handleBatchDelete = async (rows: Record<string, unknown>[]) => {
  if (!rows.length) return
  try {
    await dataService.deleteTableRows(
      props.connectionId,
      props.database,
      props.tableName,
      rows,
      props.tabId ?? ''
    )
    await loadTableData(currentPage.value)
    message.success(t('common.success'))
  } catch (error) {
    message.error(t('common.error') + ': ' + (error instanceof Error ? error.message : 'Delete failed'))
  }
}

const handleBatchEdit = async (payload: {
  column: string
  value: unknown
  selectedRows: Record<string, unknown>[]
}) => {
  const { column, value, selectedRows } = payload
  const updates: UpdateRecord[] = []
  for (const row of selectedRows) {
    const oldVal = row[column]
    if (oldVal === value && JSON.stringify(oldVal) === JSON.stringify(value)) continue
    updates.push({
      rowIndex: 0,
      column,
      oldValue: oldVal,
      newValue: value,
    })
  }
  if (!updates.length) return
  await handleUpdate(updates)
}

function validateInsertRows(rows: Record<string, unknown>[]): boolean {
  const cols = schema.value?.columns ?? []
  const nonNullable = new Set(cols.filter((c) => !c.nullable).map((c) => c.name))
  for (const row of rows) {
    for (const col of nonNullable) {
      const v = row[col]
      if (v == null || v === '') {
        message.error(t('table.validationNonNull', { column: col }))
        return false
      }
    }
  }
  return true
}

const handlePasteInsert = async (rows: Record<string, unknown>[]) => {
  if (!rows.length) return
  if (!validateInsertRows(rows)) return
  try {
    await dataService.insertTableRows(
      props.connectionId,
      props.database,
      props.tableName,
      rows,
      props.tabId ?? ''
    )
    await loadTableData(currentPage.value)
    message.success(t('common.success'))
  } catch (error) {
    message.error(t('common.error') + ': ' + (error instanceof Error ? error.message : 'Insert failed'))
  }
}

const txActive = ref(false)
const fetchTxStatus = async () => {
  try {
    const s = await dataService.getTransactionStatus(props.connectionId, props.tabId ?? '')
    txActive.value = s.active
  } catch {
    txActive.value = false
  }
}

const handleBeginTx = async () => {
  try {
    await dataService.beginTx(props.connectionId, props.tabId ?? '')
    await fetchTxStatus()
    message.success(t('table.txBegun'))
  } catch (error) {
    message.error(t('common.error') + ': ' + (error instanceof Error ? error.message : 'Begin failed'))
  }
}

const handleCommitTx = async () => {
  try {
    await dataService.commitTx(props.connectionId, props.tabId ?? '')
    await fetchTxStatus()
    await loadTableData(currentPage.value)
    message.success(t('table.txCommitted'))
  } catch (error) {
    message.error(t('common.error') + ': ' + (error instanceof Error ? error.message : 'Commit failed'))
  }
}

const handleRollbackTx = async () => {
  try {
    await dataService.rollbackTx(props.connectionId, props.tabId ?? '')
    await fetchTxStatus()
    await loadTableData(currentPage.value)
    message.success(t('table.txRolledBack'))
  } catch (error) {
    message.error(t('common.error') + ': ' + (error instanceof Error ? error.message : 'Rollback failed'))
  }
}

onMounted(() => {
  loadTableData(1)
})

watch(
  () => [props.connectionId, props.database, props.tableName],
  () => {
    loadTableData(1)
  }
)

watch(
  () => props.importTrigger,
  (trigger) => {
    if (
      trigger &&
      trigger.connectionId === props.connectionId &&
      trigger.database === props.database &&
      trigger.tableName === props.tableName
    ) {
      showImporter.value = true
      emit('clear-import-trigger')
    }
  }
)
</script>

<template>
  <div class="flex flex-col h-full theme-bg-content overflow-hidden">
    <!-- Header with table info and pagination -->
    <div class="h-12 flex items-center justify-between px-4 theme-bg-panel border-b theme-border">
      <div class="flex items-center gap-4 text-xs theme-text-muted flex-wrap">
        <span class="font-semibold theme-text">{{ tableName }}</span>
        <span>{{ t('table.totalRows') }}: {{ tableData.totalRows.toLocaleString() }}</span>
        <span>{{ t('table.page') }}: {{ currentPage }} / {{ totalPages }}</span>
        <span class="flex items-center gap-1">
          {{ t('table.rowsPerPage') }}
          <select
            :value="pageSize"
            @change="(e) => handlePageSizeChange(Number((e.target as HTMLSelectElement).value))"
            class="theme-bg-input theme-text rounded px-1.5 py-0.5 text-xs border theme-border-strong"
          >
            <option v-for="n in pageSizeOptions" :key="n" :value="n">{{ n }}</option>
          </select>
        </span>
      </div>

      <div class="flex items-center gap-2 flex-wrap">
        <template v-if="txActive">
          <span class="text-xs text-amber-400">{{ t('table.inTransaction') }}</span>
          <button
            @click="handleCommitTx"
            class="px-2 py-1 bg-green-600 hover:bg-green-500 text-white text-xs rounded"
          >
            {{ t('table.commitTx') }}
          </button>
          <button
            @click="handleRollbackTx"
            class="px-2 py-1 theme-bg-input theme-bg-input-hover theme-text text-xs rounded"
          >
            {{ t('table.rollbackTx') }}
          </button>
        </template>
        <button
          v-else
          @click="handleBeginTx"
          class="px-2 py-1 theme-bg-input theme-bg-input-hover theme-text text-xs rounded"
        >
          {{ t('table.beginTx') }}
        </button>
        <button
          @click="showImporter = true"
          class="flex items-center gap-1.5 px-3 py-1 bg-[#1677ff] hover:bg-[#4096ff] text-white text-xs rounded transition-colors font-semibold"
        >
          <Upload :size="12" />
          {{ t('table.importData') }}
        </button>
        <button
          v-if="currentPage * pageSize < tableData.totalRows"
          @click="handleLoadMore"
          class="px-3 py-1 theme-bg-input theme-bg-input-hover theme-text text-xs rounded transition-colors"
        >
          {{ t('table.loadMore') }}
        </button>
        <button
          v-if="currentPage > 1"
          @click="handlePageChange(currentPage - 1)"
          class="px-3 py-1 theme-bg-input theme-bg-input-hover theme-text text-xs rounded transition-colors"
        >
          {{ t('table.previous') }}
        </button>
        <button
          v-if="currentPage * pageSize < tableData.totalRows"
          @click="handlePageChange(currentPage + 1)"
          class="px-3 py-1 theme-bg-input theme-bg-input-hover theme-text text-xs rounded transition-colors"
        >
          {{ t('table.next') }}
        </button>
      </div>
    </div>

    <!-- Data Grid -->
    <div class="flex-1 overflow-hidden min-h-0">
      <DataGrid
        v-if="tableData.rows.length > 0"
        :data="queryResult"
        :table-context="tableContext"
        :schema="schema ?? undefined"
        @update="handleUpdate"
        @export="handleExport"
        @batch-delete="handleBatchDelete"
        @batch-edit="handleBatchEdit"
        @paste-insert="handlePasteInsert"
      />
      <div v-else-if="isLoading" class="h-full flex items-center justify-center theme-text-muted">
        <n-spin :show="true" size="medium">
          <div class="h-12 w-12" />
          <template #description>{{ t('table.loading') }}</template>
        </n-spin>
      </div>
      <div v-else class="h-full flex items-center justify-center theme-text-muted">
        <p class="text-sm">{{ t('table.noDataInTable') }}</p>
      </div>
    </div>

    <!-- Data Importer -->
    <DataImporter
      :show="showImporter"
      :connection-id="connectionId"
      :database="database"
      :table-name="tableName"
      :session-id="tabId ?? ''"
      @close="showImporter = false"
      @success="handleImportSuccess"
    />
  </div>
</template>
