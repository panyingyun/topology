<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ChevronRight, ChevronDown, Database, Table as TableIcon, Circle, FolderOpen } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { dataService } from '../services/dataService'
import type { Connection, Table } from '../types'

const { t } = useI18n()

const props = defineProps<{
  searchQuery: string
  connections: Connection[]
}>()

const emit = defineEmits<{
  (e: 'table-selected', connectionId: string, database: string, tableName: string): void
  (e: 'table-query', connectionId: string, database: string, tableName: string): void
  (e: 'edit-connection', connection: Connection): void
  (e: 'refresh-connection', connectionId: string): void
  (e: 'delete-connection', connectionId: string): void
  (e: 'new-table', connectionId: string, database: string): void
  (e: 'table-import', connectionId: string, database: string, tableName: string): void
  (e: 'table-export', connectionId: string, database: string, tableName: string): void
}>()

const connections = ref<Connection[]>([])
const expandedConnections = ref<Set<string>>(new Set())
const expandedDatabases = ref<Set<string>>(new Set())
const databasesCache = ref<Record<string, string[]>>({})
const tablesCache = ref<Record<string, Table[]>>({})

const contextMenu = ref<{
  show: boolean
  x: number
  y: number
  type: 'connection' | 'database' | 'table'
  connection: Connection | null
  connectionId?: string
  database?: string
  tableName?: string
}>({
  show: false,
  x: 0,
  y: 0,
  type: 'connection',
  connection: null,
})

function dbKey(connId: string, db: string) {
  return `${connId}:${db}`
}

watch(() => props.connections, (newConns) => {
  if (newConns && newConns.length > 0 && expandedConnections.value.size === 0) {
    expandedConnections.value.add(newConns[0].id)
    loadDatabases(newConns[0].id)
  }
  // Clear caches for deleted connections
  const connIds = new Set(newConns?.map(c => c.id) || [])
  Object.keys(databasesCache.value).forEach(id => {
    if (!connIds.has(id)) delete databasesCache.value[id]
  })
  Object.keys(tablesCache.value).forEach(key => {
    const [connId] = key.split(':')
    if (!connIds.has(connId)) delete tablesCache.value[key]
  })
}, { immediate: true })

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

const handleConnectionContextMenu = (e: MouseEvent, conn: Connection) => {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = {
    show: true,
    x: e.clientX,
    y: e.clientY,
    type: 'connection',
    connection: conn,
  }
}

const handleDatabaseContextMenu = (e: MouseEvent, conn: Connection, database: string) => {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = {
    show: true,
    x: e.clientX,
    y: e.clientY,
    type: 'database',
    connection: conn,
    connectionId: conn.id,
    database,
  }
}

const handleTableContextMenu = (e: MouseEvent, conn: Connection, database: string, tableName: string) => {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = {
    show: true,
    x: e.clientX,
    y: e.clientY,
    type: 'table',
    connection: conn,
    connectionId: conn.id,
    database,
    tableName,
  }
}

const closeContextMenu = () => {
  contextMenu.value.show = false
}

const handleEditConnection = () => {
  if (contextMenu.value.connection) {
    emit('edit-connection', contextMenu.value.connection)
  }
  closeContextMenu()
}

const handleRefreshConnection = () => {
  if (contextMenu.value.connection) {
    emit('refresh-connection', contextMenu.value.connection.id)
  }
  closeContextMenu()
}

const handleDeleteConnection = () => {
  if (contextMenu.value.connection) {
    emit('delete-connection', contextMenu.value.connection.id)
  }
  closeContextMenu()
}

const handleNewTable = () => {
  if (contextMenu.value.type === 'database' && contextMenu.value.connectionId && contextMenu.value.database) {
    emit('new-table', contextMenu.value.connectionId, contextMenu.value.database)
  }
  closeContextMenu()
}

const handleTableQuery = () => {
  if (contextMenu.value.type === 'table' && contextMenu.value.connectionId && contextMenu.value.database && contextMenu.value.tableName) {
    emit('table-query', contextMenu.value.connectionId, contextMenu.value.database, contextMenu.value.tableName)
  }
  closeContextMenu()
}

const handleTableImport = () => {
  if (contextMenu.value.type === 'table' && contextMenu.value.connectionId && contextMenu.value.database && contextMenu.value.tableName) {
    emit('table-import', contextMenu.value.connectionId, contextMenu.value.database, contextMenu.value.tableName)
  }
  closeContextMenu()
}

const handleTableExport = () => {
  if (contextMenu.value.type === 'table' && contextMenu.value.connectionId && contextMenu.value.database && contextMenu.value.tableName) {
    emit('table-export', contextMenu.value.connectionId, contextMenu.value.database, contextMenu.value.tableName)
  }
  closeContextMenu()
}

const filteredConnections = computed(() => {
  const conns = props.connections || []
  if (!props.searchQuery) return conns
  const q = props.searchQuery.toLowerCase()
  return conns.filter(
    (c) => c.name.toLowerCase().includes(q) || c.host.toLowerCase().includes(q)
  )
})

onMounted(() => {
  document.addEventListener('click', closeContextMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
})
</script>

<template>
  <div class="space-y-1">
    <div v-for="conn in filteredConnections" :key="conn.id" class="select-none">
      <!-- Level 0: Connection -->
      <div
        @click="toggleConnection(conn.id)"
        @contextmenu="handleConnectionContextMenu($event, conn)"
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
            @contextmenu="handleDatabaseContextMenu($event, conn, dbName)"
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
              @contextmenu="handleTableContextMenu($event, conn, dbName, table.name)"
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

    <!-- Context Menu -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="contextMenu.show"
          class="fixed z-[100] bg-[#252526] border border-[#333] rounded shadow-lg py-1 min-w-[160px]"
          :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
          @click.stop
        >
          <!-- 连接右键菜单 -->
          <template v-if="contextMenu.type === 'connection'">
            <button
              @click="handleEditConnection"
              class="w-full px-4 py-2 text-left text-xs text-gray-300 hover:bg-[#37373d] transition-colors"
            >
              {{ t('connection.editConnection') }}
            </button>
            <button
              @click="handleRefreshConnection"
              class="w-full px-4 py-2 text-left text-xs text-gray-300 hover:bg-[#37373d] transition-colors"
            >
              {{ t('connection.refresh') }}
            </button>
            <div class="h-px bg-[#333] my-1"></div>
            <button
              @click="handleDeleteConnection"
              class="w-full px-4 py-2 text-left text-xs text-red-400 hover:bg-[#37373d] transition-colors"
            >
              {{ t('connection.delete') }}
            </button>
          </template>
          <!-- 数据库右键菜单 -->
          <template v-else-if="contextMenu.type === 'database'">
            <button
              @click="handleNewTable"
              class="w-full px-4 py-2 text-left text-xs text-gray-300 hover:bg-[#37373d] transition-colors flex items-center gap-2"
            >
              <Database :size="12" />
              {{ t('sidebar.newTable') }}
            </button>
          </template>
          <!-- 表右键菜单：查询、导入、导出 -->
          <template v-else-if="contextMenu.type === 'table'">
            <button
              @click="handleTableQuery"
              class="w-full px-4 py-2 text-left text-xs text-gray-300 hover:bg-[#37373d] transition-colors flex items-center gap-2"
            >
              <TableIcon :size="12" />
              {{ t('tableContext.query') }}
            </button>
            <button
              @click="handleTableImport"
              class="w-full px-4 py-2 text-left text-xs text-gray-300 hover:bg-[#37373d] transition-colors flex items-center gap-2"
            >
              {{ t('tableContext.import') }}
            </button>
            <button
              @click="handleTableExport"
              class="w-full px-4 py-2 text-left text-xs text-gray-300 hover:bg-[#37373d] transition-colors flex items-center gap-2"
            >
              {{ t('tableContext.export') }}
            </button>
          </template>
        </div>
      </Transition>
      <div
        v-if="contextMenu.show"
        class="fixed inset-0 z-[99]"
        @click="closeContextMenu"
      ></div>
    </Teleport>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
