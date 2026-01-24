<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import TitleBar from '../components/TitleBar.vue'
import Sidebar from '../components/Sidebar.vue'
import TabBar from '../components/TabBar.vue'
import StatusBar from '../components/StatusBar.vue'
import QueryConsole from './QueryConsole.vue'
import ConnectionManager from './ConnectionManager.vue'
import { connectionService } from '../services/connectionService'
import type { TabItem, Connection, QueryResult } from '../types'

const sidebarWidth = ref(260)
const showConnectionManager = ref(false)
const connections = ref<Connection[]>([])
const tabs = ref<TabItem[]>([])
const activeTabId = ref('')
const currentConnection = ref<Connection | undefined>()
const queryResult = ref<QueryResult | undefined>()

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
  showConnectionManager.value = true
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

const handleTableSelected = (connectionId: string, tableName: string) => {
  const conn = connections.value.find(c => c.id === connectionId)
  if (!conn) return

  const tabId = `table-${connectionId}-${tableName}`
  const existingTab = tabs.value.find(t => t.id === tabId)
  
  if (existingTab) {
    activeTabId.value = tabId
  } else {
    const newTab: TabItem = {
      id: tabId,
      type: 'table',
      title: `${conn.name}.${tableName}`,
      connectionId,
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

const handleQueryResult = (result: QueryResult) => {
  queryResult.value = result
  const activeTab = tabs.value.find(t => t.id === activeTabId.value)
  if (activeTab) {
    activeTab.queryResult = result
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
        @update:width="sidebarWidth = $event"
        @new-connection="handleNewConnection"
        @table-selected="handleTableSelected"
      />

      <div class="flex-1 flex flex-col overflow-hidden">
        <TabBar
          v-if="tabs.length > 0"
          :tabs="tabs"
          :active-tab-id="activeTabId"
          @tab-click="handleTabClick"
          @tab-close="handleTabClose"
        />

        <div class="flex-1 overflow-hidden">
          <QueryConsole
            v-if="activeTab?.type === 'query'"
            :key="activeTab.id"
            :connection-id="activeTab.connectionId"
            @query-result="handleQueryResult"
          />
          <div v-else-if="activeTab?.type === 'table'" class="h-full flex items-center justify-center text-gray-500">
            Table view for {{ activeTab.tableName }} (to be implemented)
          </div>
          <div v-else class="h-full flex flex-col items-center justify-center text-gray-500">
            <p class="mb-4">No tabs open</p>
            <p class="text-xs text-gray-600">Select a table from the sidebar or create a new connection</p>
          </div>
        </div>
      </div>
    </div>

    <StatusBar
      :current-connection="currentConnection"
      :query-result="queryResult"
    />

    <ConnectionManager
      :show="showConnectionManager"
      @close="showConnectionManager = false"
      @connect="handleConnectionConnect"
    />
  </div>
</template>
