<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import TitleBar from '../components/TitleBar.vue'
import Sidebar from '../components/Sidebar.vue'
import TabBar from '../components/TabBar.vue'
import StatusBar from '../components/StatusBar.vue'
import QueryConsole from './QueryConsole.vue'
import DataViewer from './DataViewer.vue'
import ConnectionManager from './ConnectionManager.vue'
import TableDesigner from '../components/TableDesigner.vue'
import { connectionService } from '../services/connectionService'
import { queryService } from '../services/queryService'
import type { TabItem, Connection, QueryResult } from '../types'

const { t } = useI18n()

const sidebarWidth = ref(260)
const showConnectionManager = ref(false)
const connectionManagerMode = ref<'create' | 'edit'>('create')
const editingConnection = ref<Connection | null>(null)
const showTableDesigner = ref(false)
const tableDesignerContext = ref<{ connectionId: string; database: string } | null>(null)
const connections = ref<Connection[]>([])
const tabs = ref<TabItem[]>([])
const activeTabId = ref('')
const currentConnection = ref<Connection | undefined>()
const queryResult = ref<QueryResult | undefined>()
const editorLine = ref(1)
const editorColumn = ref(1)

onMounted(async () => {
  await loadConnections()
})

const loadConnections = async () => {
  try {
    connections.value = await connectionService.getConnections()
    if (connections.value.length > 0) {
      currentConnection.value = connections.value[0]
    }
  } catch (error) {
    console.error('Failed to load connections:', error)
  }
}

const handleNewConnection = () => {
  connectionManagerMode.value = 'create'
  editingConnection.value = null
  showConnectionManager.value = true
}

const handleEditConnection = (connection: Connection) => {
  connectionManagerMode.value = 'edit'
  editingConnection.value = connection
  showConnectionManager.value = true
}

const handleConnectionUpdate = async (connection: Connection) => {
  try {
    await connectionService.updateConnection(connection)
    await loadConnections()
    showConnectionManager.value = false
    editingConnection.value = null
  } catch (error) {
    console.error('Failed to update connection:', error)
  }
}

const handleRefreshConnection = async (connectionId: string) => {
  try {
    await connectionService.reconnectConnection(connectionId)
    await loadConnections()
  } catch (error) {
    console.error('Failed to refresh connection:', error)
  }
}

const handleDeleteConnection = async (connectionId: string) => {
  try {
    await connectionService.deleteConnection(connectionId)
    await loadConnections()
    tabs.value = tabs.value.filter((t) => t.connectionId !== connectionId)
    if (activeTabId.value && tabs.value.find((t) => t.id === activeTabId.value) === undefined) {
      activeTabId.value = tabs.value.length > 0 ? tabs.value[0].id : ''
    }
  } catch (error) {
    console.error('Failed to delete connection:', error)
  }
}

const handleConnectionConnect = async (connection: Connection) => {
  connections.value.push(connection)
  currentConnection.value = connection
  await loadConnections()
  
  // Create a new query tab
  const tabId = `query-${connection.id}`
  const newTab: TabItem = {
    id: tabId,
    type: 'query',
    title: `Query - ${connection.name}`,
    connectionId: connection.id,
    sql: 'SELECT * FROM users LIMIT 50;',
  }
  tabs.value.push(newTab)
  activeTabId.value = tabId
}

const handleTableSelected = (connectionId: string, database: string, tableName: string) => {
  const conn = connections.value.find((c) => c.id === connectionId)
  if (!conn) return

  const tabId = `table-${connectionId}-${database}-${tableName}`
  const existingTab = tabs.value.find((t) => t.id === tabId)
  if (existingTab) {
    activeTabId.value = tabId
  } else {
    const newTab: TabItem = {
      id: tabId,
      type: 'table',
      title: `${conn.name} / ${database} / ${tableName}`,
      connectionId,
      database,
      tableName,
    }
    tabs.value.push(newTab)
    activeTabId.value = tabId
  }
}

const handleTabClick = (tabId: string) => {
  activeTabId.value = tabId
}

const handleTabClose = (tabId: string) => {
  const index = tabs.value.findIndex(t => t.id === tabId)
  if (index > -1) {
    tabs.value.splice(index, 1)
    if (activeTabId.value === tabId) {
      activeTabId.value = tabs.value.length > 0 ? tabs.value[0].id : ''
    }
  }
}

const handleTabReorder = (newTabs: TabItem[]) => {
  tabs.value = newTabs
}

const handleQueryResult = (result: QueryResult) => {
  queryResult.value = result
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab) {
    activeTab.queryResult = result
  }
}

const handleEditorPosition = (line: number, column: number) => {
  editorLine.value = line
  editorColumn.value = column
}

const handleNewTable = (connectionId: string, database: string) => {
  tableDesignerContext.value = { connectionId, database }
  showTableDesigner.value = true
}

const handleCreateTable = async (sql: string) => {
  const connId = tableDesignerContext.value?.connectionId ?? currentConnection.value?.id
  if (!connId) {
    alert(t('common.error') + ': ' + 'Please select a connection first')
    return
  }
  const database = tableDesignerContext.value?.database
  const driver = (tableDesignerContext.value ? connections.value.find(c => c.id === tableDesignerContext.value!.connectionId) : currentConnection.value)?.type
  const sqlToRun = database && driver === 'mysql' ? `USE \`${database}\`;\n${sql}` : sql
  try {
    await queryService.executeQuery(connId, sqlToRun)
    alert(t('common.success') + ': ' + 'Table created successfully!')
    showTableDesigner.value = false
    tableDesignerContext.value = null
    await loadConnections()
  } catch (error) {
    console.error('Failed to create table:', error)
    alert(t('common.error') + ': ' + (error instanceof Error ? error.message : 'Unknown error'))
  }
}

const activeTab = computed(() => {
  return tabs.value.find(t => t.id === activeTabId.value)
})
</script>

<template>
  <div class="flex h-screen w-screen flex-col bg-[#1e1e1e] text-gray-200 overflow-hidden select-none font-sans">
    <TitleBar />

    <div class="flex flex-1 overflow-hidden">
      <Sidebar
        :width="sidebarWidth"
        :connections="connections"
        @update:width="sidebarWidth = $event"
        @new-connection="handleNewConnection"
        @table-selected="handleTableSelected"
        @edit-connection="handleEditConnection"
        @refresh-connection="handleRefreshConnection"
        @delete-connection="handleDeleteConnection"
        @new-table="handleNewTable"
      />

      <div class="flex-1 flex flex-col overflow-hidden">
        <TabBar
          v-if="tabs.length > 0"
          :tabs="tabs"
          :active-tab-id="activeTabId"
          @tab-click="handleTabClick"
          @tab-close="handleTabClose"
          @tab-reorder="handleTabReorder"
        />

        <div class="flex-1 overflow-hidden">
          <QueryConsole
            v-if="activeTab?.type === 'query'"
            :key="activeTab.id"
            :connection-id="activeTab.connectionId"
            :connection="currentConnection"
            @query-result="handleQueryResult"
            @editor-position="handleEditorPosition"
          />
          <DataViewer
            v-else-if="activeTab?.type === 'table' && activeTab.connectionId && activeTab.database && activeTab.tableName"
            :key="activeTab.id"
            :connection-id="activeTab.connectionId"
            :database="activeTab.database"
            :table-name="activeTab.tableName"
            @update="(updates) => console.log('Table updates:', updates)"
          />
          <div v-else class="h-full flex flex-col items-center justify-center text-gray-500">
            <p class="mb-4">{{ $t('tabs.noTabs') }}</p>
            <p class="text-xs text-gray-600">{{ $t('tabs.selectTable') }}</p>
          </div>
        </div>
      </div>
    </div>

    <StatusBar
      :current-connection="currentConnection"
      :query-result="queryResult"
      :editor-line="editorLine"
      :editor-column="editorColumn"
    />

    <ConnectionManager
      :show="showConnectionManager"
      :mode="connectionManagerMode"
      :edit-connection="editingConnection"
      @close="showConnectionManager = false; editingConnection = null"
      @connect="handleConnectionConnect"
      @update="handleConnectionUpdate"
    />

    <TableDesigner
      :show="showTableDesigner"
      :connection-id="tableDesignerContext?.connectionId ?? currentConnection?.id"
      :database="tableDesignerContext?.database ?? currentConnection?.database"
      :driver="(tableDesignerContext ? connections.find(c => c.id === tableDesignerContext?.connectionId) : currentConnection)?.type"
      @close="showTableDesigner = false; tableDesignerContext = null"
      @create="handleCreateTable"
    />
  </div>
</template>
