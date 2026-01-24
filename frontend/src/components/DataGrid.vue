<script setup lang="ts">
import { ref, watch } from 'vue'
import { VxeGrid } from 'vxe-table'
import type { VxeGridProps } from 'vxe-table'
import type { QueryResult, UpdateRecord, ExportFormat } from '../types'

const props = defineProps<{
  data: QueryResult
  queryText?: string
}>()

const emit = defineEmits<{
  (e: 'update', updates: UpdateRecord[]): void
  (e: 'export', format: ExportFormat): void
}>()

const gridOptions = ref<VxeGridProps>({
  border: true,
  height: '100%',
  columnConfig: { resizable: true },
  rowConfig: { isCurrent: true, isHover: true },
  scrollY: { enabled: true, gt: 50 },
  editConfig: { trigger: 'dblclick', mode: 'cell' },
  columns: [],
  data: [],
})

const pendingChanges = ref(0)
const gridRef = ref<any>()

watch(() => props.data, (newData) => {
  if (newData && newData.columns) {
    gridOptions.value.columns = newData.columns.map(col => ({
      field: col,
      title: col,
      editRender: { name: 'input' },
      width: 150,
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
}
</script>

<template>
  <div class="flex flex-col h-full bg-[#1e1e1e]">
    <div class="h-10 flex items-center justify-between px-4 bg-[#252526] border-b border-[#333]">
      <div class="flex items-center gap-4 text-xs text-gray-400">
        <span>Rows: {{ data.rowCount.toLocaleString() }}</span>
        <span v-if="pendingChanges > 0" class="text-yellow-400">
          {{ pendingChanges }} pending changes
        </span>
      </div>
      <div class="flex items-center gap-2">
        <button
          v-if="pendingChanges > 0"
          @click="saveChanges"
          class="px-3 py-1 bg-green-600 hover:bg-green-500 text-white text-xs rounded transition-colors"
        >
          Save Changes
        </button>
        <div class="relative">
          <button class="px-3 py-1 bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 text-xs rounded transition-colors">
            Export
          </button>
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
</style>
