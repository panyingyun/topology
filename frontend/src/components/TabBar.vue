<script setup lang="ts">
import { X } from 'lucide-vue-next'
import type { TabItem } from '../types'

const props = defineProps<{
  tabs: TabItem[]
  activeTabId: string
}>()

const emit = defineEmits<{
  (e: 'tab-click', tabId: string): void
  (e: 'tab-close', tabId: string): void
  (e: 'tab-reorder', tabs: TabItem[]): void
}>()

const handleTabClick = (tabId: string) => {
  emit('tab-click', tabId)
}

const handleTabClose = (e: MouseEvent, tabId: string) => {
  e.stopPropagation()
  emit('tab-close', tabId)
}
</script>

<template>
  <div class="flex items-center theme-bg-panel border-b theme-border overflow-x-auto custom-scrollbar">
    <div
      v-for="tab in tabs"
      :key="tab.id"
      @click="handleTabClick(tab.id)"
      :class="[
        'flex items-center gap-2 px-4 py-2 text-xs cursor-pointer border-r theme-border transition-colors min-w-[120px] select-none',
        activeTabId === tab.id
          ? 'theme-bg-content theme-text border-b-2 border-[#1677ff]'
          : 'theme-bg-panel theme-text-muted theme-bg-hover'
      ]"
      draggable="true"
      @dragstart="(e: DragEvent) => {
        if (e.dataTransfer) {
          e.dataTransfer.effectAllowed = 'move'
          e.dataTransfer.setData('text/plain', tab.id)
        }
      }"
      @dragover.prevent
      @drop="(e: DragEvent) => {
        const draggedTabId = e.dataTransfer?.getData('text/plain')
        if (draggedTabId && draggedTabId !== tab.id) {
          const draggedIndex = tabs.findIndex(t => t.id === draggedTabId)
          const targetIndex = tabs.findIndex(t => t.id === tab.id)
          if (draggedIndex !== -1 && targetIndex !== -1) {
            const newTabs = [...tabs]
            const [removed] = newTabs.splice(draggedIndex, 1)
            newTabs.splice(targetIndex, 0, removed)
            emit('tab-reorder', newTabs)
          }
        }
      }"
    >
      <span class="truncate">{{ tab.title }}</span>
      <button
        @click="handleTabClose($event, tab.id)"
        class="ml-auto theme-bg-hover rounded p-0.5 transition-colors"
      >
        <X :size="12" />
      </button>
    </div>
  </div>
</template>
