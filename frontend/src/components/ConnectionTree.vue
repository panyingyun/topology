<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ChevronRight, ChevronDown, Database, Table as TableIcon, Circle, FolderOpen } from 'lucide-vue-next'
import { connectionService } from '../services/connectionService'
import { dataService } from '../services/dataService'
import type { Connection, Table } from '../types'

const props = defineProps<{
  searchQuery: string
}>()

const emit = defineEmits<{
  (e: 'table-selected', connectionId: string, database: string, tableName: string): void
}>()

const connections = ref<Connection[]>([])
const expandedConnections = ref<Set<string>>(new Set())
const expandedDatabases = ref<Set<string>>(new Set())
const databasesCache = ref<Record<string, string[]>>({})
const tablesCache = ref<Record<string, Table[]>>({})

function dbKey(connId: string, db: string) {
  return `${connId}:${db}`
}

const loadConnections = async () => {
  try {
    const conns = await connectionService.getConnections()
    connections.value = conns
    if (connections.value.length > 0) {
      expandedConnections.value.add(connections.value[0].id)
      await loadDatabases(connections.value[0].id)
    }
  } catch (error) {
    console.error('Failed to load connections:', error)
  }
}

const loadDatabases = async (connectionId: string) => {
  if (databasesCache.value[connectionId]) return
  try {
    const dbs = await dataService.getDatabases(connectionId)
    databasesCache.value[connectionId] = dbs
  } catch (error) {
    console.error('Failed to load databases:', error)
  }
}

const loadTables = async (connectionId: string, database: string) => {
  const key = dbKey(connectionId, database)
  if (tablesCache.value[key]) return
  try {
    const tables = await dataService.getTables(connectionId, database)
    tablesCache.value[key] = tables
  } catch (error) {
    console.error('Failed to load tables:', error)
  }
}

const toggleConnection = async (connectionId: string) => {
  if (expandedConnections.value.has(connectionId)) {
    expandedConnections.value.delete(connectionId)
  } else {
    expandedConnections.value.add(connectionId)
    await loadDatabases(connectionId)
  }
}

const toggleDatabase = async (connectionId: string, database: string) => {
  const key = dbKey(connectionId, database)
  if (expandedDatabases.value.has(key)) {
    expandedDatabases.value.delete(key)
  } else {
    expandedDatabases.value.add(key)
    await loadTables(connectionId, database)
  }
}

const handleTableClick = (connectionId: string, database: string, tableName: string) => {
  emit('table-selected', connectionId, database, tableName)
}

const filteredConnections = computed(() => {
  if (!props.searchQuery) return connections.value
  const q = props.searchQuery.toLowerCase()
  return connections.value.filter(
    (c) => c.name.toLowerCase().includes(q) || c.host.toLowerCase().includes(q)
  )
})

onMounted(() => {
  loadConnections()
})
</script>

<template>
  <div class="space-y-1">
    <div v-for="conn in filteredConnections" :key="conn.id" class="select-none">
      <!-- Level 0: Connection -->
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

      <!-- Level 1: Databases -->
      <div v-if="expandedConnections.has(conn.id)" class="ml-4 space-y-0.5 mt-1">
        <div v-for="dbName in databasesCache[conn.id] || []" :key="dbKey(conn.id, dbName)" class="select-none">
          <div
            @click="toggleDatabase(conn.id, dbName)"
            class="flex items-center gap-2 px-2 py-1 rounded hover:bg-[#37373d] cursor-pointer group transition-colors"
          >
            <component
              :is="expandedDatabases.has(dbKey(conn.id, dbName)) ? ChevronDown : ChevronRight"
              :size="12"
              class="text-gray-400 shrink-0"
            />
            <FolderOpen :size="12" class="text-gray-500 group-hover:text-[#1677ff] shrink-0" />
            <span class="text-xs text-gray-400 group-hover:text-gray-200 truncate">{{ dbName }}</span>
          </div>

          <!-- Level 2: Tables -->
          <div
            v-if="expandedDatabases.has(dbKey(conn.id, dbName))"
            class="ml-4 space-y-0.5 mt-0.5"
          >
            <div
              v-for="table in tablesCache[dbKey(conn.id, dbName)] || []"
              :key="table.name"
              @click="handleTableClick(conn.id, dbName, table.name)"
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
    </div>
  </div>
</template>
