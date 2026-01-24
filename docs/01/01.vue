<script setup lang="ts">
import { ref } from 'vue'

// 状态管理
const activeTab = ref('query-1')
const sidebarWidth = ref(260)
const isConnected = ref(true)

// 模拟数据
const databases = [
  { name: 'Production_DB', open: true, tables: ['users', 'orders', 'products'] },
  { name: 'Staging_Test', open: false, tables: ['logs', 'temp_data'] }
]

const closeTab = (id: string) => {
  // 关闭逻辑
}
</script>

<template>
  <div class="flex h-screen w-screen flex-col bg-[#1e1e1e] text-gray-200 overflow-hidden select-none font-sans">
    
    <header 
      class="flex h-10 items-center justify-between bg-[#252526] px-3 border-b border-[#333] z-50"
      style="--wails-draggable: drag"
    >
      <div class="flex items-center gap-2 text-sm font-semibold">
        <div class="w-5 h-5 bg-blue-600 rounded flex items-center justify-center text-[10px]">Go</div>
        <span class="tracking-tight text-gray-300">GopherDB Client</span>
      </div>
      
      <div class="flex no-drag" style="--wails-draggable: no-drag">
        <button class="hover:bg-[#37373d] w-10 h-10 transition-colors text-xs">－</button>
        <button class="hover:bg-red-600 w-10 h-10 transition-colors text-xs">✕</button>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden">
      
      <aside 
        :style="{ width: sidebarWidth + 'px' }"
        class="flex flex-col bg-[#252526] border-r border-[#333] transition-[width] duration-75"
      >
        <div class="p-3 space-y-3">
          <button class="w-full bg-blue-600 hover:bg-blue-500 text-white text-xs py-2 rounded-md font-bold transition-all active:scale-[0.98]">
            + NEW CONNECTION
          </button>
          <div class="relative group">
            <input 
              type="text" 
              placeholder="Filter..." 
              class="w-full bg-[#3c3c3c] text-xs px-3 py-1.5 rounded border border-transparent focus:border-blue-500 outline-none transition-all"
            />
          </div>
        </div>
        
        <div class="flex-1 overflow-y-auto custom-scrollbar px-2">
          <div v-for="db in databases" :key="db.<script setup lang="ts">
import { ref } from 'vue'

// 状态管理
const activeTab = ref('query-1')
const sidebarWidth = ref(260)
const isConnected = ref(true)

// 模拟数据
const databases = [
  { name: 'Production_DB', open: true, tables: ['users', 'orders', 'products'] },
  { name: 'Staging_Test', open: false, tables: ['logs', 'temp_data'] }
]

const closeTab = (id: string) => {
  // 关闭逻辑
}
</script>

<template>
  <div class="flex h-screen w-screen flex-col bg-[#1e1e1e] text-gray-200 overflow-hidden select-none font-sans">
    
    <header 
      class="flex h-10 items-center justify-between bg-[#252526] px-3 border-b border-[#333] z-50"
      style="--wails-draggable: drag"
    >
      <div class="flex items-center gap-2 text-sm font-semibold">
        <div class="w-5 h-5 bg-blue-600 rounded flex items-center justify-center text-[10px]">Go</div>
        <span class="tracking-tight text-gray-300">GopherDB Client</span>
      </div>
      
      <div class="flex no-drag" style="--wails-draggable: no-drag">
        <button class="hover:bg-[#37373d] w-10 h-10 transition-colors text-xs">－</button>
        <button class="hover:bg-red-600 w-10 h-10 transition-colors text-xs">✕</button>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden">
      
      <aside 
        :style="{ width: sidebarWidth + 'px' }"
        class="flex flex-col bg-[#252526] border-r border-[#333] transition-[width] duration-75"
      >
        <div class="p-3 space-y-3">
          <button class="w-full bg-blue-600 hover:bg-blue-500 text-white text-xs py-2 rounded-md font-bold transition-all active:scale-[0.98]">
            + NEW CONNECTION
          </button>
          <div class="relative group">
            <input 
              type="text" 
              placeholder="Filter..." 
              class="w-full bg-[#3c3c3c] text-xs px-3 py-1.5 rounded border border-transparent focus:border-blue-500 outline-none transition-all"
            />
          </div>
        </div>
        
        <div class="flex-1 overflow-y-auto custom-scrollbar px-2">
          <div v-for="db in databases" :key="db.