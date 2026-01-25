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
  <div class="h-6 flex items-center justify-between px-4 bg-[#252526] border-t border-[#333] text-[10px] text-gray-400 font-mono">
    <div class="flex items-center gap-4">
      <span v-if="currentConnection">
        {{ currentConnection.host }}:{{ currentConnection.port }} / {{ currentConnection.username }}
      </span>
      <span v-else class="text-gray-600">{{ t('statusBar.notConnected') }}</span>
    </div>

    <div class="flex items-center gap-4">
      <span v-if="queryResult && queryResult.executionTime">
        {{ t('statusBar.queryTime') }}: {{ queryResult.executionTime }}ms
      </span>
      <span v-if="queryResult && queryResult.affectedRows !== undefined">
        {{ t('statusBar.rows') }}: {{ queryResult.affectedRows }}
      </span>
      <span v-if="editorLine !== undefined && editorColumn !== undefined">
        {{ t('statusBar.line') }} {{ editorLine }}, {{ t('statusBar.column') }} {{ editorColumn }}
      </span>
    </div>
  </div>
</template>
