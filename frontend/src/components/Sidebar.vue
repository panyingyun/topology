<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search, Plus, FileDown, ChevronLeft, ChevronRight, MoreHorizontal } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import ConnectionTree from './ConnectionTree.vue'
import type { Connection } from '../types'

const { t } = useI18n()

const props = defineProps<{
  width: number
  collapsed: boolean
  connections: Connection[]
  connectionInvalidation?: { id: string; at: number } | null
}>()

const emit = defineEmits<{
  (e: 'update:width', width: number): void
  (e: 'update:collapsed', value: boolean): void
  (e: 'new-connection'): void
  (e: 'table-selected', connectionId: string, database: string, tableName: string): void
  (e: 'table-query', connectionId: string, database: string, tableName: string): void
  (e: 'edit-connection', connection: Connection): void
  (e: 'refresh-connection', connectionId: string): void
  (e: 'delete-connection', connectionId: string): void
  (e: 'new-table', connectionId: string, database: string): void
  (e: 'table-import', connectionId: string, database: string, tableName: string): void
  (e: 'table-export', connectionId: string, database: string, tableName: string): void
  (e: 'open-monitor', connection: import('../types').Connection): void
  (e: 'backup', connectionId: string): void
  (e: 'restore', connectionId: string): void
  (e: 'er-diagram', connectionId: string, database: string): void
  (e: 'import-navicat'): void
  (e: 'open-backup-manager'): void
  (e: 'open-data-compare'): void
  (e: 'open-schema-sync'): void
  (e: 'open-audit-log'): void
}>()

const searchQuery = ref('')
const isResizing = ref(false)
const moreOpen = ref(false)
const moreBtnRef = ref<HTMLElement | null>(null)

const isNarrow = computed(() => props.collapsed || props.width <= 56)

const moreDropdownStyle = computed(() => {
  if (!moreBtnRef.value) return {}
  const rect = moreBtnRef.value.getBoundingClientRect()
  return {
    left: `${rect.right + 4}px`,
    top: `${rect.top}px`,
  }
})

const startResize = (e: MouseEvent) => {
  if (props.collapsed) return
  isResizing.value = true
  const startX = e.clientX
  const startWidth = props.width

  const handleMouseMove = (e: MouseEvent) => {
    const newWidth = Math.max(240, Math.min(320, startWidth + (e.clientX - startX)))
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

const toggleCollapse = () => {
  emit('update:collapsed', !props.collapsed)
}

const moreOptions = [
  { key: 'backup', label: () => t('backup.manage'), action: () => emit('open-backup-manager') },
  { key: 'compare', label: () => t('dataCompare.title'), action: () => emit('open-data-compare') },
  { key: 'sync', label: () => t('schemaSync.title'), action: () => emit('open-schema-sync') },
  { key: 'audit', label: () => t('audit.title'), action: () => emit('open-audit-log') },
]

const handleMoreSelect = (opt: { action: () => void }) => {
  moreOpen.value = false
  opt.action()
}
</script>

<template>
  <aside
    :style="{ width: width + 'px' }"
    class="flex flex-col theme-bg-panel border-r theme-border transition-[width] duration-150 relative shrink-0"
  >
    <div class="p-2 flex flex-col gap-1.5">
      <!-- 新建连接 -->
      <button
        @click="emit('new-connection')"
        class="w-full bg-[#1677ff] hover:bg-[#4096ff] text-white text-xs py-2 rounded-md font-bold transition-all active:scale-[0.98] flex items-center justify-center gap-2"
        :title="t('sidebar.newConnection')"
      >
        <Plus :size="14" />
        <span v-if="!isNarrow">{{ t('sidebar.newConnection') }}</span>
      </button>

      <!-- 导入 Navicat -->
      <button
        @click="emit('import-navicat')"
        class="w-full border theme-border text-xs py-2 rounded-md font-medium transition-all active:scale-[0.98] flex items-center justify-center gap-2 theme-text theme-bg-hover"
        :title="t('sidebar.importNavicat')"
      >
        <FileDown :size="14" />
        <span v-if="!isNarrow">{{ t('sidebar.importNavicat') }}</span>
      </button>

      <!-- 更多（下拉） -->
      <div ref="moreBtnRef" class="relative">
        <button
          @click="moreOpen = !moreOpen"
          class="w-full border theme-border text-xs py-2 rounded-md font-medium transition-all active:scale-[0.98] flex items-center justify-center gap-2 theme-text theme-bg-hover"
          :title="t('sidebar.more')"
        >
          <MoreHorizontal :size="14" />
          <span v-if="!isNarrow">{{ t('sidebar.more') }}</span>
        </button>
        <Teleport to="body">
          <Transition name="fade">
            <div
              v-if="moreOpen && moreBtnRef"
              class="fixed z-[200] theme-bg-panel border theme-border rounded shadow-lg py-1 min-w-[140px]"
              :style="moreDropdownStyle"
            >
              <button
                v-for="opt in moreOptions"
                :key="opt.key"
                @click="handleMoreSelect(opt)"
                class="w-full px-4 py-2 text-left text-xs theme-text theme-bg-hover transition-colors"
              >
                {{ opt.label() }}
              </button>
            </div>
          </Transition>
          <div v-if="moreOpen" class="fixed inset-0 z-[199]" @click="moreOpen = false" />
        </Teleport>
      </div>

      <!-- 筛选 -->
      <div v-if="!isNarrow" class="relative">
        <Search :size="14" class="absolute left-2 top-1/2 -translate-y-1/2 theme-text-muted" />
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="t('sidebar.filter')"
          class="w-full theme-input text-xs pl-8 pr-3 py-1.5 rounded border transition-all"
        />
      </div>
      <button
        v-else
        @click="toggleCollapse"
        class="w-full border theme-border py-1.5 rounded flex items-center justify-center theme-text theme-bg-hover"
        :title="t('sidebar.filter')"
      >
        <Search :size="14" />
      </button>

      <!-- 折叠/展开 -->
      <button
        @click="toggleCollapse"
        class="w-full py-1.5 rounded flex items-center justify-center theme-text-muted hover:theme-text theme-bg-hover"
        :title="collapsed ? t('sidebar.expand') : t('sidebar.collapse')"
      >
        <ChevronLeft v-if="!collapsed" :size="16" />
        <ChevronRight v-else :size="16" />
      </button>
    </div>

    <div v-if="!isNarrow" class="flex-1 min-h-0 flex flex-col px-2">
      <ConnectionTree
        :search-query="searchQuery"
        :connections="connections"
        :connection-invalidation="connectionInvalidation"
        @table-selected="(connId, db, tableName) => emit('table-selected', connId, db, tableName)"
        @table-query="(connId, db, tableName) => emit('table-query', connId, db, tableName)"
        @edit-connection="(conn) => emit('edit-connection', conn)"
        @refresh-connection="(id) => emit('refresh-connection', id)"
        @delete-connection="(id) => emit('delete-connection', id)"
        @new-table="(connId, database) => emit('new-table', connId, database)"
        @table-import="(connId, db, tableName) => emit('table-import', connId, db, tableName)"
        @table-export="(connId, db, tableName) => emit('table-export', connId, db, tableName)"
        @open-monitor="(conn) => emit('open-monitor', conn)"
        @backup="(id) => emit('backup', id)"
        @restore="(id) => emit('restore', id)"
        @er-diagram="(connId, db) => emit('er-diagram', connId, db)"
      />
    </div>

    <!-- Resize handle -->
    <div
      v-if="!collapsed"
      @mousedown="startResize"
      class="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-[#1677ff]/50 transition-colors"
    />
  </aside>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
