<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { VxeTableInstance, VxeGridProps } from 'vxe-table'

// 1. 模拟数据生成
const generateMockData = (count: number) => {
  return Array.from({ length: count }).map((_, i) => ({
    id: i + 1,
    name: `User_${i}`,
    role: i % 2 === 0 ? 'Admin' : 'Developer',
    email: `user${i}@example.com`,
    status: i % 3 === 0 ? 'Active' : 'Idle',
    last_login: '2026-01-24 12:00:00'
  }))
}

// 2. 表格配置
const tableRef = ref<VxeTableInstance>()
const gridOptions = reactive<VxeGridProps>({
  border: true,
  height: 'auto',
  columnConfig: { resizable: true },
  rowConfig: { isCurrent: true, isHover: true },
  // 开启虚拟滚动
  scrollY: { enabled: true, gt: 50 },
  editConfig: { trigger: 'dblclick', mode: 'cell' },
  columns: [
    { type: 'seq', width: 60, title: 'No.' },
    { field: 'id', title: 'ID', width: 80 },
    { field: 'name', title: 'Username', editRender: { name: 'input' } },
    { field: 'role', title: 'Role', editRender: { name: 'input' } },
    { field: 'email', title: 'Email', editRender: { name: 'input' } },
    { field: 'status', title: 'Status', width: 100 },
    { field: 'last_login', title: 'Last Login', width: 160 }
  ],
  data: generateMockData(1000) // 初始化1000条，虚拟滚动可支持10万+
})

// 3. 提交修改逻辑
const pendingChanges = ref(0)
const handleEditClosed = () => {
  const { updateRecords } = tableRef.value!.getRecordset()
  pendingChanges.value = updateRecords.length
}

const saveChanges = async () => {
  const { updateRecords } = tableRef.value!.getRecordset()
  console.log('提交到 Go 后端:', updateRecords)
  // 调用 Wails: await window.go.main.App.BatchUpdate(updateRecords)
  pendingChanges.value = 0
  tableRef.value?.reloadRow(updateRecords, null) // 清除修改状态
}
</script>

<template>
  <div class="flex flex-col h-full bg-[#1e1e1e]">
    
    <div class="h-10 flex items-center justify-between px-4 bg-[#252526] border-b border-[#333]">
      <div class="flex items-center gap-4">