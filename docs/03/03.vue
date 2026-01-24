<script setup lang="ts">
import { ref, onMounted, shallowRef } from 'vue';
import loader from '@monaco-editor/loader';

// 1. 状态管理
const editorContainer = ref<HTMLElement | null>(null);
const editorInstance = shallowRef<any>(null); // 使用 shallowRef 避免深层响应式破坏 Monaco 性能
const sqlQuery = ref("SELECT * FROM orders WHERE status = 'shipped' LIMIT 50;");
const isRunning = ref(false);
const results = ref<any[]>([]);

// 2. 初始化 Monaco
onMounted(async () => {
  const monaco = await loader.init();
  if (editorContainer.value) {
    editorInstance.value = monaco.editor.create(editorContainer.value, {
      value: sqlQuery.value,
      language: 'sql',
      theme: 'vs-dark',
      automaticLayout: true,
      minimap: { enabled: false },
      fontSize: 14,
      fontFamily: "'JetBrains Mono', monospace",
      padding: { top: 12 }
    });

    // 监听内容变化
    editorInstance.value.onDidChangeModelContent(() => {
      sqlQuery.value = editorInstance.value.getValue();
    });
  }
});

// 3. 执行 SQL 逻辑
const runExecute = async () => {
  isRunning.value = true;
  
  // 模拟 Wails 后端调用: 
  // const res = await window.go.main.App.ExecuteQuery(sqlQuery.value);
  
  setTimeout(() => {
    results.value = [
      { id: 'ORD-001', customer: 'TechCorp', amount: 1250.50, date: '2026-01-20' },
      { id: 'ORD-002', customer: 'GlobalSoft', amount: 890.00, date: '2026-01-21' },
      { id: 'ORD-003', customer: 'EcoSolutions', amount: 3400.25, date: '2026-01-24' },
    ];
    isRunning.value = false;
  }, 800);
};
</script>

<template>
  <div class="flex flex-col h-full bg-[#1e1e1e] overflow-hidden">
    
    <div class="h-10 flex items-center justify-between px-4 bg-[#252526] border-b border-[#333]">
      <div class="flex items-center gap-3">
        <button 
          @click="runExecute"
          :disabled="isRunning"
          class="flex items-center gap-2 px-4 py-1 rounded text-xs font-bold transition-all"
          :class="isRunning ? 'bg-gray-600' : 'bg-green-600 hover:bg-green-500 active:scale-95'"
        >
          <span v-if="isRunning" class="w-3 h-3 border-2 border-white/30 border-t-white animate-spin rounded-full"></span>
          <span v-else>▶</span>
          {{ isRunning ? 'RUNNING' : 'EXECUTE' }}
        </button>
        
        <button class="px-3 py-1 rounded text-xs bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300">
          Format SQL
        </button>
      </div>
      
      <div class="flex items-center gap-4 text-[10px] text-gray-500 font-mono italic">
        <span>Dialect: <b class="text-blue-400">PostgreSQL</b></span>
        <span>Schema: public</span>
      </div>
    </div>

    <div class="flex-1 relative border-b border-[#333]">
      <div ref="editorContainer" class="absolute inset-0"></div>
    </div>

    <div class="h-1/3 flex flex-col bg-[#1e1e1e]">
      <div class="h-8 flex items-center justify-between px-4 bg-[#252526] border-b border-[#333] text-[10px] text-gray-400">
        <div class="flex gap-4">
          <span>RESULTS: <b class="text-white">{{ results.length }} rows</b></span>
          <span v-if="!isRunning">TIME: 42ms</span>
        </div>
        <div class="flex gap-3">
          <button class="hover:text-white transition-colors">CLEAR</button>
          <button class="hover:text-white transition-colors">EXPORT</button>
        </div>
      </div>

      <div class="flex-1 overflow-auto custom-scrollbar">
        <table v-if="results.length > 0" class="w-full text-left text-xs border-collapse">
          <thead class="sticky top-0 bg-[#2d2d2d] text-gray-300 border-b border-[#333] z-10">
            <tr>
              <th v-for="key in Object.keys(results[0])" :key="key" class="px-4 py-2 border-r border-[#333] uppercase">
                {{ key }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, idx) in results" :key="idx" class="border-b border-[#2d2d2d] hover:bg-[#2d2d2d] group">
              <td v-for="val in row" :key="val" class="px-4 py-2 border-r border-[#2d2d2d] font-mono text-gray-400 group-hover:text-gray-200">
                {{ val }}
              </td>
            </tr>
          </tbody>
        </table>
        
        <div v-else class="h-full flex flex-col items-center justify-center text-gray-600 opacity-50 space-y-2">
          <div class="text-3xl">⌨</div>
          <p class="text-[10px] uppercase tracking-widest">Execute query to see results</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #3e3e42;
  border-radius: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
</style>