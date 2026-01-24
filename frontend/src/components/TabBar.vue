<script setup lang="ts">
import { computed } from 'vue'
import { X } from 'lucide-vue-next'
import type { TabItem } from '../types'

const props = defineProps<{
  tabs: TabItem[]
  activeTabId: string
}>()

const emit = defineEmits<{
  (e: 'tab-click', tabId: string): void
  (e: 'tab-close', tabId: string): void
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
  <div class="flex items-center bg-[#2d2d30] border-b border-[#333] overflow-x-auto custom-scrollbar">
    <div
      v-for="tab in tabs"
      :key="tab.id"
      @click="handleTabClick(tab.id)"
      :class="[
        'flex items-center gap-2 px-4 py-2 text-xs cursor-pointer border-r border-[#333] transition-colors min-w-[120px]',
        activeTabId === tab.id
          ? 'bg-[#1e1e1e] text-gray-200 border-b-2 border-[#1677ff]'
          : 'bg-[#2d2d30] text-gray-400 hover:bg-[#37373d] hover:text-gray-300'
      ]"
    >
      <span class="truncate">{{ tab.title }}</span>
      <button
        @click="handleTabClose($event, tab.id)"
        class="ml-auto hover:bg-[#424242] rounded p-0.5 transition-colors"
      >
        <X :size="12" />
      </button>
    </div>
  </div>
</template>
