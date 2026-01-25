<script setup lang="ts">
import { ref } from 'vue'
import { Search, Plus } from 'lucide-vue-next'
import ConnectionTree from './ConnectionTree.vue'
import type { Connection } from '../types'

const props = defineProps<{
  width: number
  connections: Connection[]
}>()

const emit = defineEmits<{
  (e: 'update:width', width: number): void
  (e: 'new-connection'): void
  (e: 'table-selected', connectionId: string, database: string, tableName: string): void
  (e: 'edit-connection', connection: Connection): void
  (e: 'refresh-connection', connectionId: string): void
  (e: 'delete-connection', connectionId: string): void
}>()

const searchQuery = ref('')
const isResizing = ref(false)

const startResize = (e: MouseEvent) => {
  isResizing.value = true
  const startX = e.clientX
  const startWidth = props.width

  const handleMouseMove = (e: MouseEvent) => {
    const newWidth = Math.max(240, Math.min(400, startWidth + (e.clientX - startX)))
    emit('update:width', newWidth)
  }

  const handleMouseUp = () => {
    isResizing.value = false
    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseup', handleMouseUp)
  }

  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
}
</script>

<template>
  <aside
    :style="{ width: width + 'px' }"
    class="flex flex-col bg-[#252526] border-r border-[#333] transition-[width] duration-75 relative"
  >
    <div class="p-3 space-y-3">
      <button
        @click="emit('new-connection')"
        class="w-full bg-[#1677ff] hover:bg-[#4096ff] text-white text-xs py-2 rounded-md font-bold transition-all active:scale-[0.98] flex items-center justify-center gap-2"
      >
        <Plus :size="14" />
        NEW CONNECTION
      </button>
      <div class="relative">
        <Search :size="14" class="absolute left-2 top-1/2 -translate-y-1/2 text-gray-400" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Filter..."
          class="w-full bg-[#3c3c3c] text-xs pl-8 pr-3 py-1.5 rounded border border-transparent focus:border-[#1677ff] outline-none transition-all text-gray-200"
        />
      </div>
    </div>

    <div class="flex-1 overflow-y-auto custom-scrollbar px-2">
      <ConnectionTree
        :search-query="searchQuery"
        :connections="connections"
        @table-selected="(connId, db, tableName) => emit('table-selected', connId, db, tableName)"
        @edit-connection="(conn) => emit('edit-connection', conn)"
        @refresh-connection="(id) => emit('refresh-connection', id)"
        @delete-connection="(id) => emit('delete-connection', id)"
      />
    </div>

    <!-- Resize handle -->
    <div
      @mousedown="startResize"
      class="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-[#1677ff]/50 transition-colors"
    ></div>
  </aside>
</template>
