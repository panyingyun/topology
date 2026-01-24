<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ChevronRight, ChevronDown, Database, Table as TableIcon, Circle } from 'lucide-vue-next'
import { connectionService } from '../services/connectionService'
import { dataService } from '../services/dataService'
import type { Connection, Table } from '../types'

const props = defineProps<{
  searchQuery: string
}>()

const emit = defineEmits<{
  (e: 'table-selected', connectionId: string, tableName: string): void
}>()

const connections = ref<Connection[]>([])
const expandedConnections = ref<Set<string>>(new Set())
const tablesCache = ref<Record<string, Table[]>>({})

const loadConnections = async () => {
  try {
    const conns = await connectionService.getConnections()
    connections.value = conns
    // Auto expand first connection
    if (connections.value.length > 0) {
      expandedConnections.value.add(connections.value[0].id)
      await loadTables(connections.value[0].id)
    }
  } catch (error) {
    console.error('Failed to load connections:', error)
  }
}

const loadTables = async (connectionId: string) => {
  if (tablesCache.value[connectionId]) {
    return
  }
  try {
    const tables = await dataService.getTables(connectionId)
    tablesCache.value[connectionId] = tables
  } catch (error) {
    console.error('Failed to load tables:', error)
  }
}

const toggleConnection = (connectionId: string) => {
  if (expandedConnections.value.has(connectionId)) {
    expandedConnections.value.delete(connectionId)
  } else {
    expandedConnections.value.add(connectionId)
    loadTables(connectionId)
  }
}

const handleTableClick = (connectionId: string, tableName: string) => {
  emit('table-selected', connectionId, tableName)
}

const filteredConnections = computed(() => {
  if (!props.searchQuery) {
    return connections.value
  }
  const query = props.searchQuery.toLowerCase()
  return connections.value.filter(conn => 
    conn.name.toLowerCase().includes(query) ||
    conn.host.toLowerCase().includes(query)
  )
})

onMounted(() => {
  loadConnections()
})
</script>

<template>
  <div class="space-y-1">
    <div
      v-for="conn in filteredConnections"
      :key="conn.id"
      class="select-none"
    >
      <div
        @click="toggleConnection(conn.id)"
        class="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-[#37373d] cursor-pointer group transition-colors"
      >
        <component
          :is="expandedConnections.has(conn.id) ? ChevronDown : ChevronRight"
          :size="14"
          class="text-gray-400 shrink-0"
        />
        <Database :size="14" class="text-gray-400 shrink-0" />
        <span class="text-xs text-gray-300 flex-1 truncate">{{ conn.name }}</span>
        <Circle
          :size="6"
          :class="conn.status === 'connected' ? 'text-green-500' : 'text-gray-500'"
          :fill="conn.status === 'connected' ? 'currentColor' : 'none'"
          class="shrink-0"
        />
      </div>

      <div
        v-if="expandedConnections.has(conn.id)"
        class="ml-4 space-y-0.5 mt-1"
      >
        <div
          v-for="table in tablesCache[conn.id] || []"
          :key="table.name"
          @click="handleTableClick(conn.id, table.name)"
          class="flex items-center gap-2 px-2 py-1 rounded hover:bg-[#37373d] cursor-pointer group transition-colors"
        >
          <TableIcon :size="12" class="text-gray-500 group-hover:text-[#1677ff] shrink-0" />
          <span class="text-xs text-gray-400 group-hover:text-gray-200 truncate">{{ table.name }}</span>
          <span v-if="table.rowCount" class="text-[10px] text-gray-600 ml-auto">
            {{ table.rowCount.toLocaleString() }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
