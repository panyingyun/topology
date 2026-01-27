<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Upload } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import DataGrid from '../components/DataGrid.vue'
import DataImporter from '../components/DataImporter.vue'
import { dataService } from '../services/dataService'
import type { TableData, UpdateRecord, ExportFormat, QueryResult, ImportResult } from '../types'

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
      offset
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
  } catch (error) {
    console.error('Failed to load table data:', error)
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

const handleUpdate = (updates: UpdateRecord[]) => {
  emit('update', updates)
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

      <div class="flex items-center gap-2">
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
        @update="handleUpdate"
        @export="handleExport"
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
