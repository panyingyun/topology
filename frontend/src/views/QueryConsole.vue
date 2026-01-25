<script setup lang="ts">
import { ref, onMounted, onUnmounted, shallowRef } from 'vue'
import { Play, Square, FileCode, Save, History, Sparkles } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import * as monaco from 'monaco-editor'
import { queryService } from '../services/queryService'
import DataGrid from '../components/DataGrid.vue'
import QueryHistory from '../components/QueryHistory.vue'
import SQLAnalyzer from '../components/SQLAnalyzer.vue'
import type { QueryResult, Connection } from '../types'

const { t } = useI18n()

const props = defineProps<{
  connectionId?: string
  connection?: Connection
}>()

const emit = defineEmits<{
  (e: 'query-result', result: QueryResult): void
  (e: 'editor-position', line: number, column: number): void
}>()

const editorContainer = ref<HTMLElement | null>(null)
const editor = shallowRef<any>(null)
const sqlQuery = ref("SELECT * FROM users LIMIT 50;")
const isRunning = ref(false)
const queryResult = ref<QueryResult | null>(null)
const editorLine = ref(1)
const editorColumn = ref(1)
const showHistory = ref(false)
const showAnalyzer = ref(false)

onMounted(async () => {
  if (editorContainer.value) {
    editor.value = monaco.editor.create(editorContainer.value, {
      value: sqlQuery.value,
      language: 'sql',
      theme: 'vs-dark',
      automaticLayout: true,
      minimap: { enabled: true },
      fontSize: 14,
      fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
      padding: { top: 12 },
      scrollBeyondLastLine: false,
      wordWrap: 'on',
    })

    editor.value.onDidChangeModelContent(() => {
      sqlQuery.value = editor.value.getValue()
    })

    editor.value.onDidChangeCursorPosition((e: any) => {
      editorLine.value = e.position.lineNumber
      editorColumn.value = e.position.column
      emit('editor-position', editorLine.value, editorColumn.value)
    })

    // Ctrl+Enter to execute
    editor.value.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
      runExecute()
    })
  }
})

onUnmounted(() => {
  if (editor.value) {
    editor.value.dispose()
  }
})

const runExecute = async () => {
  if (!sqlQuery.value.trim() || isRunning.value) return

  // Get selected text if any, otherwise use full query
  let queryToExecute = sqlQuery.value
  if (editor.value) {
    const selection = editor.value.getSelection()
    if (selection && !selection.isEmpty()) {
      queryToExecute = editor.value.getModel()?.getValueInRange(selection) || sqlQuery.value
    }
  }

  isRunning.value = true
  try {
    const result = await queryService.executeQuery(props.connectionId || '1', queryToExecute)
    queryResult.value = result
    emit('query-result', result)
  } catch (error) {
    console.error('Query execution error:', error)
  } finally {
    isRunning.value = false
  }
}

const stopQuery = () => {
  // In real implementation, this would cancel the query
  isRunning.value = false
}

const formatSQL = async () => {
  if (editor.value) {
    try {
      const formatted = await queryService.formatSQL(sqlQuery.value)
      editor.value.setValue(formatted)
    } catch (error) {
      console.error('Format error:', error)
    }
  }
}

const saveScript = () => {
  // In real implementation, save to file or local storage
  console.log('Save script:', sqlQuery.value)
}

const handleHistorySelect = (sql: string) => {
  if (editor.value) {
    editor.value.setValue(sql)
    sqlQuery.value = sql
  }
}
</script>

<template>
  <div class="flex flex-col h-full bg-[#1e1e1e] overflow-hidden">
    <!-- Toolbar -->
    <div class="h-10 flex items-center justify-between px-4 bg-[#252526] border-b border-[#333]">
      <div class="flex items-center gap-3">
        <button
          @click="runExecute"
          :disabled="isRunning"
          :class="[
            'flex items-center gap-2 px-4 py-1 rounded text-xs font-bold transition-all',
            isRunning
              ? 'bg-gray-600 cursor-not-allowed'
              : 'bg-green-600 hover:bg-green-500 active:scale-95'
          ]"
        >
          <Play v-if="!isRunning" :size="14" />
          <Square v-else :size="14" class="animate-pulse" />
          {{ isRunning ? t('query.running') : t('query.execute') }}
        </button>

        <button
          v-if="isRunning"
          @click="stopQuery"
          class="flex items-center gap-2 px-3 py-1 rounded text-xs font-bold bg-red-600 hover:bg-red-500 transition-all"
        >
          <Square :size="14" />
          {{ t('query.stop') }}
        </button>

        <button
          @click="formatSQL"
          class="px-3 py-1 rounded text-xs bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 transition-colors"
        >
          <FileCode :size="14" class="inline mr-1" />
          {{ t('query.formatSQL') }}
        </button>

        <button
          @click="saveScript"
          class="px-3 py-1 rounded text-xs bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 transition-colors"
        >
          <Save :size="14" class="inline mr-1" />
          {{ t('query.save') }}
        </button>

        <button
          @click="showHistory = !showHistory"
          :class="[
            'px-3 py-1 rounded text-xs transition-colors',
            showHistory
              ? 'bg-[#1677ff] text-white'
              : 'bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300'
          ]"
        >
          <History :size="14" class="inline mr-1" />
          {{ t('query.history') }}
        </button>

        <button
          @click="showAnalyzer = true"
          class="px-3 py-1 rounded text-xs bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 transition-colors"
        >
          <Sparkles :size="14" class="inline mr-1" />
          {{ t('query.analyzeSQL') }}
        </button>
      </div>

      <div class="flex items-center gap-4 text-[10px] text-gray-500 font-mono italic">
        <span>Dialect: <b class="text-[#1677ff]">PostgreSQL</b></span>
        <span>Schema: public</span>
      </div>
    </div>

    <!-- Editor and Results -->
    <div class="flex-1 flex flex-col min-h-0">
      <!-- SQL Editor (60%) -->
      <div class="flex-[6] relative border-b border-[#333]">
        <div ref="editorContainer" class="absolute inset-0"></div>
      </div>

      <!-- Results (40%) -->
      <div class="flex-[4] overflow-hidden min-h-0">
        <DataGrid
          v-if="queryResult && queryResult.rows.length > 0"
          :data="queryResult"
          :query-text="sqlQuery"
          @update="(updates) => console.log('Updates:', updates)"
          @export="(format) => console.log('Export:', format)"
        />
        <div v-else-if="queryResult && queryResult.error" class="h-full flex items-center justify-center">
          <div class="text-red-400 text-sm">
            <p class="font-semibold">Query Error:</p>
            <p class="text-xs mt-1">{{ queryResult.error }}</p>
          </div>
        </div>
        <div v-else class="h-full flex items-center justify-center text-gray-500 text-sm">
          {{ t('query.noResults') }}. {{ t('query.executeQuery') }}
        </div>
      </div>
    </div>

    <!-- Query History Panel -->
    <QueryHistory
      :connection-id="connectionId"
      :show="showHistory"
      @select="handleHistorySelect"
      @close="showHistory = false"
    />

    <!-- SQL Analyzer -->
    <SQLAnalyzer
      :show="showAnalyzer"
      :sql="sqlQuery"
      :driver="connection?.type"
      @close="showAnalyzer = false"
    />
  </div>
</template>
