<script setup lang="ts">
import type { Connection, QueryResult } from '../types'

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
      <span v-else class="text-gray-600">Not connected</span>
    </div>

    <div class="flex items-center gap-4">
      <span v-if="queryResult && queryResult.executionTime">
        Query time: {{ queryResult.executionTime }}ms
      </span>
      <span v-if="queryResult && queryResult.affectedRows !== undefined">
        Rows: {{ queryResult.affectedRows }}
      </span>
      <span v-if="editorLine !== undefined && editorColumn !== undefined">
        Ln {{ editorLine }}, Col {{ editorColumn }}
      </span>
    </div>
  </div>
</template>
