<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Connection, QueryResult } from '../types'

const { t } = useI18n()

const props = defineProps<{
  currentConnection?: Connection
  queryResult?: QueryResult
  editorLine?: number
  editorColumn?: number
}>()
</script>

<template>
  <div class="h-6 flex items-center justify-between px-4 theme-bg-panel border-t theme-border text-[10px] theme-text-muted font-mono">
    <div class="flex items-center gap-4">
      <span v-if="currentConnection">
        {{ currentConnection.host }}:{{ currentConnection.port }} / {{ currentConnection.username }}
      </span>
      <span v-else class="opacity-70">{{ t('statusBar.notConnected') }}</span>
    </div>

    <div class="flex items-center gap-4">
      <span v-if="queryResult && queryResult.executionTime">
        {{ t('statusBar.queryTime') }}: {{ queryResult.executionTime }}ms
      </span>
      <span v-if="queryResult && queryResult.affectedRows !== undefined">
        {{ t('statusBar.rows') }}: {{ queryResult.affectedRows }}
      </span>
      <span v-if="queryResult?.cached" class="text-emerald-500">{{ t('statusBar.cacheHit') }}</span>
      <span v-if="editorLine !== undefined && editorColumn !== undefined">
        {{ t('statusBar.line') }} {{ editorLine }}, {{ t('statusBar.column') }} {{ editorColumn }}
      </span>
    </div>
  </div>
</template>
