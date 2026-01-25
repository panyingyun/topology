<script setup lang="ts">
import { ref, watch } from 'vue'
import { GitBranch, AlertTriangle, X, Search, Key } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { queryService } from '../services/queryService'
import type { ExecutionPlanResult, ExecutionPlanNode } from '../types'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  connectionId: string
  tabId: string
  sql: string
  driver?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const plan = ref<ExecutionPlanResult | null>(null)
const isLoading = ref(false)

const load = async () => {
  if (!props.sql.trim() || !props.connectionId) return
  isLoading.value = true
  plan.value = null
  try {
    plan.value = await queryService.getExecutionPlan(
      props.connectionId,
      props.tabId || '',
      props.sql
    )
  } catch (error) {
    plan.value = {
      nodes: [],
      summary: {},
      error: error instanceof Error ? error.message : 'Unknown error',
    }
  } finally {
    isLoading.value = false
  }
}

watch(
  () => [props.show, props.sql, props.connectionId],
  () => {
    if (props.show && props.sql?.trim() && props.connectionId) load()
  }
)

function nodeTypeColor(node: ExecutionPlanNode): string {
  if (node.fullTableScan) return 'border-red-500/80 bg-red-500/10'
  if (node.type === 'Filter') return 'border-amber-500/60 bg-amber-500/10'
  if (node.type === 'Join') return 'border-green-500/60 bg-green-500/10'
  return 'border-[#1677ff]/60 bg-[#1677ff]/10'
}
</script>

<template>
  <Transition name="fade">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <div class="bg-[#252526] rounded-lg border border-[#333] w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b border-[#333] flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-200 flex items-center gap-2">
            <GitBranch :size="20" class="text-[#1677ff]" />
            {{ t('explainPlan.title') }}
          </h2>
          <button
            @click="emit('close')"
            class="text-gray-400 hover:text-gray-200 transition-colors p-1"
          >
            <X :size="20" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <div v-if="isLoading" class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="w-8 h-8 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
              <p class="text-sm text-gray-400">{{ t('explainPlan.loading') }}</p>
            </div>
          </div>

          <div v-else-if="plan?.error" class="p-4 bg-red-500/10 rounded border border-red-500/50">
            <p class="text-sm text-red-400">{{ plan.error }}</p>
            <p v-if="plan.error?.includes('MySQL')" class="text-xs text-gray-500 mt-2">{{ t('explainPlan.mysqlOnly') }}</p>
          </div>

          <div v-else-if="plan" class="space-y-6">
            <!-- Flowchart: vertical chain of nodes -->
            <div class="space-y-0">
              <div
                v-for="(node, index) in plan.nodes"
                :key="node.id"
                class="flex flex-col items-center"
              >
                <!-- Connector line from previous node -->
                <div
                  v-if="index > 0"
                  class="w-0.5 h-6 bg-[#444] flex-shrink-0"
                />
                <div
                  :class="[
                    'w-full max-w-md rounded-lg border-2 px-4 py-3 transition-colors',
                    nodeTypeColor(node)
                  ]"
                >
                  <div class="flex items-center justify-between gap-4 flex-wrap">
                    <div class="flex items-center gap-2 flex-wrap">
                      <Search v-if="node.type === 'Scan'" :size="16" class="text-[#1677ff] flex-shrink-0" />
                      <span class="text-sm font-semibold text-gray-200">{{ node.type }}</span>
                      <span class="text-sm text-gray-400">—</span>
                      <span class="text-sm font-mono text-gray-300">{{ node.label }}</span>
                      <span v-if="node.detail" class="text-xs text-gray-500">({{ node.detail }})</span>
                    </div>
                    <div class="flex items-center gap-2">
                      <span
                        v-if="node.fullTableScan"
                        class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-red-500/30 text-red-400 border border-red-500/50"
                      >
                        <AlertTriangle :size="12" />
                        {{ t('explainPlan.fullTableScan') }}
                      </span>
                      <span
                        v-if="node.indexUsed"
                        class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-green-500/30 text-green-400 border border-green-500/50"
                      >
                        <Key :size="12" />
                        {{ t('explainPlan.indexUsed') }}
                      </span>
                    </div>
                  </div>
                  <div v-if="node.rows !== undefined && node.rows > 0" class="mt-1 text-xs text-gray-500">
                    ~{{ node.rows.toLocaleString() }} {{ t('explainPlan.rows') }}
                  </div>
                  <div v-if="node.extra" class="mt-1 text-xs text-gray-500 truncate" :title="node.extra">
                    {{ node.extra }}
                  </div>
                </div>
              </div>
            </div>

            <!-- Summary & Warnings -->
            <div
              v-if="plan.summary?.warnings?.length"
              class="p-4 bg-amber-500/10 rounded border border-amber-500/50"
            >
              <div class="flex items-center gap-2 mb-2">
                <AlertTriangle :size="16" class="text-amber-400" />
                <span class="text-sm font-semibold text-amber-400">{{ t('explainPlan.warnings') }}</span>
              </div>
              <ul class="space-y-1">
                <li
                  v-for="(w, i) in plan.summary.warnings"
                  :key="i"
                  class="text-xs text-amber-200 flex items-start gap-2"
                >
                  <span class="mt-0.5">•</span>
                  <span>{{ w }}</span>
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div class="px-6 py-4 border-t border-[#333] flex justify-end bg-[#2d2d30]">
          <button
            @click="emit('close')"
            class="px-6 py-2 rounded text-xs font-semibold bg-[#1677ff] hover:bg-[#4096ff] text-white"
          >
            {{ t('common.close') }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
