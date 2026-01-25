<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Upload } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import DataGrid from '../components/DataGrid.vue'
import DataImporter from '../components/DataImporter.vue'
import { dataService } from '../services/dataService'
import type { TableData, UpdateRecord, ExportFormat, QueryResult, ImportResult } from '../types'

const { t } = useI18n()

const props = defineProps<{
  connectionId: string
  database: string
  tableName: string
}>()

const emit = defineEmits<{
  (e: 'update', updates: UpdateRecord[]): void
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

const loadTableData = async (page: number = 1) => {
  isLoading.value = true
  try {
    const offset = (page - 1) * pageSize.value
    const data = await dataService.getTableData(
      props.connectionId,
      props.database,
      props.tableName,
      pageSize.value,
      offset
    )
    tableData.value = data
    currentPage.value = page
  } catch (error) {
    console.error('Failed to load table data:', error)
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

const handleUpdate = (updates: UpdateRecord[]) => {
  emit('update', updates)
}

const handleExport = async (format: ExportFormat) => {
  try {
    const result = await dataService.exportData(
      props.connectionId,
      props.database,
      props.tableName,
      format
    )
    if (result.success) {
      console.log('Export successful:', result.filename)
      // Note: In production, show a notification to user
    } else {
      console.error('Export failed:', result.error)
    }
  } catch (error) {
    console.error('Export error:', error)
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
</script>

<template>
  <div class="flex flex-col h-full bg-[#1e1e1e] overflow-hidden">
    <!-- Header with table info and pagination -->
    <div class="h-12 flex items-center justify-between px-4 bg-[#252526] border-b border-[#333]">
      <div class="flex items-center gap-4 text-xs text-gray-400">
        <span class="font-semibold text-gray-300">{{ tableName }}</span>
        <span>{{ t('table.totalRows') }}: {{ tableData.totalRows.toLocaleString() }}</span>
        <span>{{ t('table.page') }}: {{ currentPage }} / {{ Math.ceil(tableData.totalRows / pageSize) || 1 }}</span>
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
          class="px-3 py-1 bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 text-xs rounded transition-colors"
        >
          {{ t('table.loadMore') }}
        </button>
        <button
          v-if="currentPage > 1"
          @click="handlePageChange(currentPage - 1)"
          class="px-3 py-1 bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 text-xs rounded transition-colors"
        >
          {{ t('table.previous') }}
        </button>
        <button
          v-if="currentPage * pageSize < tableData.totalRows"
          @click="handlePageChange(currentPage + 1)"
          class="px-3 py-1 bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 text-xs rounded transition-colors"
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
      <div v-else-if="isLoading" class="h-full flex items-center justify-center text-gray-500">
        <div class="text-center">
          <div class="w-8 h-8 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
          <p class="text-xs">{{ t('table.loading') }}</p>
        </div>
      </div>
      <div v-else class="h-full flex items-center justify-center text-gray-500">
        <p class="text-sm">{{ t('table.noData') }}</p>
      </div>
    </div>

    <!-- Data Importer -->
    <DataImporter
      :show="showImporter"
      :connection-id="connectionId"
      :database="database"
      :table-name="tableName"
      @close="showImporter = false"
      @success="handleImportSuccess"
    />
  </div>
</template>
