<script setup lang="ts">
import { ref, onMounted, onUnmounted, shallowRef, watch } from 'vue'
import { Play, Square, FileCode, Save, History, Sparkles, Bookmark, GitBranch } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import * as monaco from 'monaco-editor'
import { queryService } from '../services/queryService'
import { snippetService } from '../services/snippetService'
import { useSchemaMetadata } from '../composables/useSchemaMetadata'
import DataGrid from '../components/DataGrid.vue'
import QueryHistory from '../components/QueryHistory.vue'
import Snippets from '../components/Snippets.vue'
import SQLAnalyzer from '../components/SQLAnalyzer.vue'
import ExecutionPlanViewer from '../components/ExecutionPlanViewer.vue'
import type { QueryResult, Connection } from '../types'

const { t } = useI18n()

const props = defineProps<{
  /** Tab id for per-tab DB session isolation */
  tabId?: string
  connectionId?: string
  connection?: Connection
  /** One-shot SQL to inject into editor (e.g. from table right-click "Query") */
  initialSql?: string
  /** Restore editor content when switching back to this tab */
  restoreSql?: string
  /** Restore query result when switching back to this tab */
  savedQueryResult?: QueryResult
  /** Current context: database and table (e.g. when opened from table right-click "Query") */
  database?: string
  tableName?: string
}>()

const emit = defineEmits<{
  (e: 'query-result', result: QueryResult): void
  (e: 'editor-position', line: number, column: number): void
  (e: 'update-sql', sql: string): void
  (e: 'initial-sql-applied'): void
}>()

const DEFAULT_SQL = 'SELECT * FROM users LIMIT 50;'

const editorContainer = ref<HTMLElement | null>(null)
const editor = shallowRef<any>(null)
const sqlQuery = ref(DEFAULT_SQL)
const isRunning = ref(false)
const queryResult = ref<QueryResult | null>(null)
const editorLine = ref(1)
const editorColumn = ref(1)
const showHistory = ref(false)
const showSnippets = ref(false)
const snippetRefreshKey = ref(0)
const showAnalyzer = ref(false)
const showExplainPlan = ref(false)

const {
  load: loadSchemaMetadata,
  getAllTableNames,
  getColumnsForTable,
  getAllColumns,
} = useSchemaMetadata()
const connectionIdForCompletion = ref('')
let completionProviderDisposable: monaco.IDisposable | null = null

// SQL vs Results split: default 1/3 SQL, 2/3 results; user can drag to resize
const SPLIT_STORAGE_KEY = 'query-console-split'
const DEFAULT_SQL_PERCENT = 100 / 3
const sqlHeightPercent = ref(
  Math.min(85, Math.max(15, Number(localStorage.getItem(SPLIT_STORAGE_KEY)) || DEFAULT_SQL_PERCENT))
)
const splitContainerRef = ref<HTMLElement | null>(null)
const isResizing = ref(false)

const startResize = (e: MouseEvent) => {
  if (!splitContainerRef.value) return
  isResizing.value = true
  const startY = e.clientY
  const startPercent = sqlHeightPercent.value
  const containerHeight = splitContainerRef.value.clientHeight

  const onMove = (e: MouseEvent) => {
    const delta = e.clientY - startY
    const deltaPercent = (delta / containerHeight) * 100
    let next = startPercent + deltaPercent
    next = Math.min(85, Math.max(15, next))
    sqlHeightPercent.value = next
    editor.value?.layout?.()
  }
  const onUp = () => {
    isResizing.value = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
    localStorage.setItem(SPLIT_STORAGE_KEY, String(sqlHeightPercent.value))
    editor.value?.layout?.()
  }

  document.body.style.cursor = 'ns-resize'
  document.body.style.userSelect = 'none'
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

onMounted(async () => {
  if (editorContainer.value) {
    const initialValue = props.initialSql ?? props.restoreSql ?? DEFAULT_SQL
    sqlQuery.value = initialValue
    editor.value = monaco.editor.create(editorContainer.value, {
      value: initialValue,
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
      emit('update-sql', sqlQuery.value)
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

    completionProviderDisposable = registerSQLCompletionProvider(editor.value)
    applyInitialSql()
  }
})

watch(() => props.connectionId, (id) => {
  connectionIdForCompletion.value = id || ''
  if (id) loadSchemaMetadata(id)
}, { immediate: true })

function registerSQLCompletionProvider(_editorInstance: any): monaco.IDisposable {
  const kindTable = monaco.languages.CompletionItemKind.Class
  const kindColumn = monaco.languages.CompletionItemKind.Field
  const kindKeyword = monaco.languages.CompletionItemKind.Keyword

  return monaco.languages.registerCompletionItemProvider('sql', {
    triggerCharacters: ['.', ' ', ',', '\n'],
    provideCompletionItems(model, position) {
      const connId = connectionIdForCompletion.value
      const word = model.getWordUntilPosition(position)
      const lineText = model.getLineContent(position.lineNumber).slice(0, position.column - 1)
      const trimmedRight = lineText.replace(/\s+$/, '')
      const suggestions: monaco.languages.CompletionItem[] = []

      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn,
      }

      // After "xxx." -> suggest columns for table xxx
      const dotMatch = trimmedRight.match(/(\w+)\.\s*$/)
      if (dotMatch && connId) {
        const tableOrAlias = dotMatch[1]
        const columns = getColumnsForTable(connId, tableOrAlias)
        for (const col of columns) {
          suggestions.push({
            label: col.label,
            kind: kindColumn,
            detail: col.detail,
            insertText: col.label,
            range,
          })
        }
        if (suggestions.length) return { suggestions }
      }

      // After FROM, JOIN, INTO -> prefer table names
      const afterFrom = /\b(?:FROM|JOIN|INTO|UPDATE)\s+$/i.test(trimmedRight)
      if (connId) {
        const tables = getAllTableNames(connId)
        for (const name of tables) {
          suggestions.push({
            label: name,
            kind: kindTable,
            detail: 'table',
            insertText: name,
            range,
          })
        }
        if (!afterFrom) {
          const columns = getAllColumns(connId)
          for (const col of columns) {
            suggestions.push({
              label: col.label,
              kind: kindColumn,
              detail: col.detail,
              insertText: col.label,
              range,
            })
          }
        }
      }

      const sqlKeywords = ['SELECT', 'FROM', 'WHERE', 'JOIN', 'LEFT', 'RIGHT', 'INNER', 'ON', 'GROUP BY', 'ORDER BY', 'LIMIT', 'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE', 'AS', 'AND', 'OR', 'ASC', 'DESC']
      for (const kw of sqlKeywords) {
        if (!word.word || kw.toLowerCase().startsWith(word.word.toLowerCase())) {
          suggestions.push({
            label: kw,
            kind: kindKeyword,
            insertText: kw,
            range,
          })
        }
      }
      return { suggestions }
    },
  })
}

watch(() => props.initialSql, () => applyInitialSql())
// Restore saved result when switching back to this tab (immediate so we sync on mount too)
watch(
  () => props.savedQueryResult,
  (next) => {
    queryResult.value = next !== undefined && next !== null ? next : null
  },
  { immediate: true }
)

function applyInitialSql() {
  if (!props.initialSql || !editor.value) return
  editor.value.setValue(props.initialSql)
  sqlQuery.value = props.initialSql
  emit('update-sql', props.initialSql)
  emit('initial-sql-applied')
}

onUnmounted(() => {
  completionProviderDisposable?.dispose()
  completionProviderDisposable = null
  if (editor.value) {
    editor.value.dispose()
  }
})

const runExecute = async () => {
  if (!sqlQuery.value.trim() || isRunning.value) return

  const connectionId = props.connectionId
  if (!connectionId) {
    queryResult.value = {
      columns: [],
      rows: [],
      rowCount: 0,
      error: 'No connection selected',
    }
    return
  }

  let queryToExecute = sqlQuery.value
  if (editor.value) {
    const selection = editor.value.getSelection()
    if (selection && !selection.isEmpty()) {
      queryToExecute = editor.value.getModel()?.getValueInRange(selection) || sqlQuery.value
    }
  }

  isRunning.value = true
  try {
    const result = await queryService.executeQuery(connectionId, props.tabId ?? '', queryToExecute)
    queryResult.value = result
    emit('query-result', result)
  } catch (error) {
    console.error('Query execution error:', error)
    const errMsg = error instanceof Error ? error.message : 'Unknown error'
    queryResult.value = {
      columns: [],
      rows: [],
      rowCount: 0,
      error: errMsg,
    }
    emit('query-result', queryResult.value)
  } finally {
    isRunning.value = false
  }
}

const stopQuery = () => {
  isRunning.value = false
  // Note: backend query cannot be cancelled; next run will work. User sees "Running" clear.
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

const handleSnippetSelect = (sql: string) => {
  if (editor.value) {
    const model = editor.value.getModel()
    const selection = editor.value.getSelection()
    if (model && selection) {
      editor.value.executeEdits('snippet', [{ range: selection, text: sql }])
    }
    editor.value.focus()
  }
  showSnippets.value = false
}

const handleSaveSnippet = async (alias: string) => {
  try {
    await snippetService.saveSnippet(alias, sqlQuery.value)
    snippetRefreshKey.value += 1
  } catch (error) {
    console.error('Failed to save snippet:', error)
    throw error
  }
}
</script>

<template>
  <div class="flex flex-col h-full theme-bg-content overflow-hidden">
    <!-- Toolbar -->
    <div class="h-10 flex items-center justify-between px-4 theme-bg-panel border-b theme-border">
      <div class="flex items-center gap-3">
        <span
          v-if="database && tableName"
          class="flex items-center gap-1.5 px-3 py-1 rounded text-xs bg-[#3c3c3c] text-gray-400 font-mono border border-[#404040]"
          :title="`库: ${database} / 表: ${tableName}`"
        >
          <span class="text-gray-500">库</span>
          <span class="text-gray-300">{{ database }}</span>
          <span class="text-gray-600">/</span>
          <span class="text-gray-500">表</span>
          <span class="text-gray-300">{{ tableName }}</span>
        </span>
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
          @click="showSnippets = !showSnippets"
          :class="[
            'px-3 py-1 rounded text-xs transition-colors',
            showSnippets
              ? 'bg-[#1677ff] text-white'
              : 'bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300'
          ]"
        >
          <Bookmark :size="14" class="inline mr-1" />
          {{ t('snippets.title') }}
        </button>

        <button
          @click="showAnalyzer = true"
          class="px-3 py-1 rounded text-xs bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 transition-colors"
        >
          <Sparkles :size="14" class="inline mr-1" />
          {{ t('query.analyzeSQL') }}
        </button>

        <button
          @click="showExplainPlan = true"
          class="px-3 py-1 rounded text-xs bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 transition-colors"
          :title="t('explainPlan.title')"
        >
          <GitBranch :size="14" class="inline mr-1" />
          {{ t('explainPlan.viewPlan') }}
        </button>
      </div>
    </div>

    <!-- Editor and Results (resizable split: default SQL 1/3, Results 2/3) -->
    <div ref="splitContainerRef" class="flex-1 flex flex-col min-h-0">
      <!-- SQL Editor -->
      <div
        class="flex-shrink-0 relative min-h-0 theme-border"
        :style="{ height: sqlHeightPercent + '%' }"
      >
        <div ref="editorContainer" class="absolute inset-0"></div>
      </div>

      <!-- Resize handle -->
      <div
        class="flex-shrink-0 h-1.5 cursor-ns-resize theme-bg-panel border-y theme-border flex items-center justify-center hover:bg-[#1677ff]/20 transition-colors"
        @mousedown="startResize"
      >
        <div class="w-12 h-0.5 rounded theme-bg-hover" />
      </div>

      <!-- Results -->
      <div class="flex-1 min-h-0 overflow-hidden">
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

    <!-- Snippets Panel -->
    <Snippets
      :show="showSnippets"
      :current-sql="sqlQuery"
      :refresh-trigger="snippetRefreshKey"
      @select="handleSnippetSelect"
      @save="handleSaveSnippet"
      @close="showSnippets = false"
    />

    <!-- SQL Analyzer -->
    <SQLAnalyzer
      :show="showAnalyzer"
      :sql="sqlQuery"
      :driver="connection?.type"
      @close="showAnalyzer = false"
    />

    <!-- Execution Plan -->
    <ExecutionPlanViewer
      :show="showExplainPlan"
      :connection-id="connectionId ?? ''"
      :tab-id="tabId ?? ''"
      :sql="sqlQuery"
      :driver="connection?.type"
      @close="showExplainPlan = false"
    />
  </div>
</template>
