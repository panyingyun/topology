<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useVirtualizer } from '@tanstack/vue-virtual'
import { ChevronRight, ChevronDown, Database, Table as TableIcon, Circle, FolderOpen } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { dataService } from '../services/dataService'
import type { Connection, Table } from '../types'

const { t } = useI18n()

const ROW_HEIGHT = 28

const props = defineProps<{
  searchQuery: string
  connections: Connection[]
  connectionInvalidation?: { id: string; at: number } | null
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
  (e: 'open-monitor', connection: Connection): void
  (e: 'backup', connectionId: string): void
  (e: 'restore', connectionId: string): void
  (e: 'er-diagram', connectionId: string, database: string): void
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

const contextMenuConnection = computed(() => {
  if (contextMenu.value.type !== 'table' || !contextMenu.value.connectionId) return null
  return props.connections.find((c) => c.id === contextMenu.value.connectionId) ?? null
})

watch(() => props.connections, (newConns) => {
  // 连接默认折叠，用户点击再展开（不再自动展开第一个连接）
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

watch(
  () => props.connectionInvalidation?.at,
  (at) => {
    const inv = props.connectionInvalidation
    if (!inv || at == null) return
    const id = inv.id
    delete databasesCache.value[id]
    Object.keys(tablesCache.value).forEach((key) => {
      if (key.startsWith(id + ':')) delete tablesCache.value[key]
    })
    expandedConnections.value = new Set([...expandedConnections.value, id])
    loadDatabases(id)
  }
)

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

const handleOpenMonitor = () => {
  if (contextMenu.value.connection) {
    emit('open-monitor', contextMenu.value.connection)
  }
  closeContextMenu()
}

const handleBackup = () => {
  if (contextMenu.value.connection) {
    emit('backup', contextMenu.value.connection.id)
  }
  closeContextMenu()
}

const handleRestore = () => {
  if (contextMenu.value.connection) {
    emit('restore', contextMenu.value.connection.id)
  }
  closeContextMenu()
}

const handleNewTable = () => {
  if (contextMenu.value.type === 'database' && contextMenu.value.connectionId && contextMenu.value.database) {
    emit('new-table', contextMenu.value.connectionId, contextMenu.value.database)
  }
  closeContextMenu()
}

function handleERDiagram() {
  if (contextMenu.value.type === 'database' && contextMenu.value.connectionId && contextMenu.value.database) {
    emit('er-diagram', contextMenu.value.connectionId, contextMenu.value.database)
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

// 扁平化树用于虚拟滚动
type FlatItem =
  | { type: 'connection'; conn: Connection }
  | { type: 'database'; conn: Connection; db: string }
  | { type: 'table'; conn: Connection; db: string; table: Table }

const flatTreeItems = computed<FlatItem[]>(() => {
  const out: FlatItem[] = []
  for (const conn of filteredConnections.value) {
    out.push({ type: 'connection', conn })
    if (expandedConnections.value.has(conn.id)) {
      const dbs = databasesCache.value[conn.id] || []
      for (const db of dbs) {
        out.push({ type: 'database', conn, db })
        if (expandedDatabases.value.has(dbKey(conn.id, db))) {
          const tables = tablesCache.value[dbKey(conn.id, db)] || []
          for (const table of tables) {
            out.push({ type: 'table', conn, db, table })
          }
        }
      }
    }
  }
  return out
})

const scrollRef = ref<HTMLElement | null>(null)
const virtualizer = useVirtualizer(
  computed(() => ({
    count: flatTreeItems.value.length,
    getScrollElement: () => scrollRef.value,
    estimateSize: () => ROW_HEIGHT,
    overscan: 5,
  }))
)

const virtualItemsWithData = computed(() => {
  const items = flatTreeItems.value
  return virtualizer.value.getVirtualItems().map((v) => ({
    vitem: v,
    item: items[v.index],
  }))
})

onMounted(() => {
  document.addEventListener('click', closeContextMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
})
</script>

<template>
  <div
    ref="scrollRef"
    class="flex-1 min-h-0 overflow-y-auto custom-scrollbar"
  >
    <div
      :style="{ height: virtualizer.getTotalSize() + 'px', position: 'relative' }"
      class="w-full"
    >
      <div
        v-for="({ vitem, item }) in virtualItemsWithData"
        :key="String(vitem.key)"
        :data-index="vitem.index"
        class="absolute left-0 right-0 px-2 select-none"
        :style="{ transform: 'translateY(' + vitem.start + 'px)' }"
      >
        <div
          v-if="item?.type === 'connection'"
          @click="toggleConnection(item.conn.id)"
          @contextmenu="handleConnectionContextMenu($event, item.conn)"
          class="flex items-center gap-2 px-2 py-1.5 rounded theme-bg-hover cursor-pointer group transition-colors"
        >
          <component
            :is="expandedConnections.has(item.conn.id) ? ChevronDown : ChevronRight"
            :size="14"
            class="theme-text-muted shrink-0"
          />
          <Database :size="14" class="theme-text-muted shrink-0" />
          <span class="text-xs theme-text flex-1 truncate">{{ item.conn.name }}</span>
          <Circle
            :size="6"
            :class="item.conn.status === 'connected' ? 'text-green-500' : 'theme-text-muted'"
            :fill="item.conn.status === 'connected' ? 'currentColor' : 'none'"
            class="shrink-0"
          />
        </div>
        <div
          v-else-if="item?.type === 'database'"
          @click="toggleDatabase(item.conn.id, item.db)"
          @contextmenu="handleDatabaseContextMenu($event, item.conn, item.db)"
          class="flex items-center gap-2 ml-4 px-2 py-1 rounded theme-bg-hover cursor-pointer group transition-colors"
        >
          <component
            :is="expandedDatabases.has(dbKey(item.conn.id, item.db)) ? ChevronDown : ChevronRight"
            :size="12"
            class="theme-text-muted shrink-0"
          />
          <FolderOpen :size="12" class="theme-text-muted group-hover:text-[#1677ff] shrink-0" />
          <span class="text-xs theme-text-muted group-hover:theme-text truncate">{{ item.db }}</span>
        </div>
        <div
          v-else-if="item?.type === 'table'"
          @click="handleTableClick(item.conn.id, item.db, item.table.name)"
          @contextmenu="handleTableContextMenu($event, item.conn, item.db, item.table.name)"
          class="flex items-center gap-2 ml-8 px-2 py-1 rounded theme-bg-hover cursor-pointer group transition-colors"
        >
          <TableIcon :size="12" class="theme-text-muted group-hover:text-[#1677ff] shrink-0" />
          <span class="text-xs theme-text-muted group-hover:theme-text truncate">{{ item.table.name }}</span>
          <span v-if="item.table.rowCount" class="text-[10px] theme-text-muted opacity-80 ml-auto">
            {{ item.table.rowCount.toLocaleString() }}
          </span>
        </div>
      </div>
    </div>

    <!-- Context Menu -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="contextMenu.show"
          class="fixed z-[100] theme-bg-panel border theme-border rounded shadow-lg py-1 min-w-[160px]"
          :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
          @click.stop
        >
          <!-- 连接右键菜单 -->
          <template v-if="contextMenu.type === 'connection'">
            <button
              @click="handleEditConnection"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors"
            >
              {{ t('connection.editConnection') }}
            </button>
            <button
              @click="handleRefreshConnection"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors"
            >
              {{ t('connection.refresh') }}
            </button>
            <button
              v-if="['mysql','postgresql','postgres','sqlite'].includes(contextMenu.connection?.type || '')"
              @click="handleBackup"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors"
            >
              {{ t('backup.backupNow') }}
            </button>
            <button
              v-if="['mysql','postgresql','postgres','sqlite'].includes(contextMenu.connection?.type || '')"
              @click="handleRestore"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors"
            >
              {{ t('backup.restoreBackup') }}
            </button>
            <button
              @click="handleOpenMonitor"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors"
            >
              {{ t('monitor.openMonitor') }}
            </button>
            <div class="h-px bg-[var(--border)] my-1"></div>
            <button
              @click="handleDeleteConnection"
              class="w-full px-4 py-2 text-left text-xs text-red-400 theme-bg-hover transition-colors"
            >
              {{ t('connection.delete') }}
            </button>
          </template>
          <!-- 数据库右键菜单 -->
          <template v-else-if="contextMenu.type === 'database'">
            <button
              @click="handleERDiagram"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors"
            >
              {{ t('erDiagram.title') }}
            </button>
            <button
              @click="handleNewTable"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors flex items-center gap-2"
            >
              <Database :size="12" />
              {{ t('sidebar.newTable') }}
            </button>
          </template>
          <!-- 表右键菜单：查询、导入、导出 -->
          <template v-else-if="contextMenu.type === 'table'">
            <button
              @click="handleTableQuery"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors flex items-center gap-2"
            >
              <TableIcon :size="12" />
              {{ t('tableContext.query') }}
            </button>
            <button
              v-if="!contextMenuConnection?.readOnly"
              @click="handleTableImport"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors flex items-center gap-2"
            >
              {{ t('tableContext.import') }}
            </button>
            <button
              @click="handleTableExport"
              class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors flex items-center gap-2"
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
