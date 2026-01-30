<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { GitBranch, AlertTriangle, X, Search, Key, Copy, GitCompare } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { queryService } from '../services/queryService'
import type { ExecutionPlanResult, ExecutionPlanNode, IndexSuggestion } from '../types'

const { t } = useI18n()
const message = useMessage()

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
const planA = ref<ExecutionPlanResult | null>(null)
const planB = ref<ExecutionPlanResult | null>(null)
const compareMode = ref(false)
const isLoading = ref(false)
const isLoadingB = ref(false)
const suggestions = ref<IndexSuggestion[]>([])
const suggestionsError = ref<string | null>(null)

const totalRows = computed(() => {
  const p = plan.value
  if (!p?.nodes?.length) return 0
  return p.nodes.reduce((s, n) => s + (n.rows ?? 0), 0)
})

const nodeShare = (node: ExecutionPlanNode) => {
  const tot = totalRows.value
  if (tot <= 0) return 0
  return Math.round(((node.rows ?? 0) / tot) * 100)
}

async function load() {
  if (!props.sql.trim() || !props.connectionId) return
  isLoading.value = true
  plan.value = null
  try {
    plan.value = await queryService.getExecutionPlan(
      props.connectionId,
      props.tabId || '',
      props.sql
    )
    if (plan.value && !plan.value.error) {
      try {
        const res = await queryService.getIndexSuggestions(
          props.connectionId,
          props.tabId || '',
          props.sql
        )
        suggestions.value = res.suggestions ?? []
        suggestionsError.value = res.error ?? null
      } catch {
        suggestions.value = []
        suggestionsError.value = null
      }
    } else {
      suggestions.value = []
      suggestionsError.value = null
    }
  } catch (error) {
    plan.value = {
      nodes: [],
      summary: {},
      error: error instanceof Error ? error.message : 'Unknown error',
    }
    suggestions.value = []
    suggestionsError.value = null
  } finally {
    isLoading.value = false
  }
}

function saveAsA() {
  if (plan.value && !plan.value.error) {
    planA.value = JSON.parse(JSON.stringify(plan.value))
    compareMode.value = true
  }
}

async function runB() {
  if (!props.sql.trim() || !props.connectionId) return
  isLoadingB.value = true
  planB.value = null
  try {
    planB.value = await queryService.getExecutionPlan(
      props.connectionId,
      props.tabId || '',
      props.sql
    )
  } catch (error) {
    planB.value = {
      nodes: [],
      summary: {},
      error: error instanceof Error ? error.message : 'Unknown error',
    }
  } finally {
    isLoadingB.value = false
  }
}

function clearCompare() {
  planA.value = null
  planB.value = null
  compareMode.value = false
}

async function copySuggestionSql(sql: string) {
  try {
    await navigator.clipboard.writeText(sql)
    message.success(t('indexSuggestions.copied'))
  } catch {}
}

watch(
  () => [props.show, props.sql, props.connectionId],
  () => {
    if (props.show && props.sql?.trim() && props.connectionId) {
      clearCompare()
      load()
    }
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
      <div class="theme-bg-panel rounded-lg border theme-border w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b theme-border flex items-center justify-between flex-wrap gap-2">
          <h2 class="text-lg font-semibold theme-text flex items-center gap-2">
            <GitBranch :size="20" class="text-[#1677ff]" />
            {{ t('explainPlan.title') }}
          </h2>
          <div class="flex items-center gap-2">
            <template v-if="plan && !plan.error">
              <button
                v-if="!compareMode"
                class="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
                @click="saveAsA"
              >
                <GitCompare :size="12" />
                {{ t('explainPlan.saveAsA') }}
              </button>
              <template v-else>
                <button
                  class="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
                  :disabled="isLoadingB"
                  @click="runB"
                >
                  {{ t('explainPlan.runB') }}
                </button>
                <button
                  class="px-2.5 py-1 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
                  @click="clearCompare"
                >
                  {{ t('explainPlan.clearCompare') }}
                </button>
              </template>
            </template>
            <button
              @click="emit('close')"
              class="theme-text-muted-hover transition-colors p-1"
            >
              <X :size="20" />
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <div v-if="isLoading" class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="w-8 h-8 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
              <p class="text-sm theme-text-muted">{{ t('explainPlan.loading') }}</p>
            </div>
          </div>

          <div v-else-if="plan?.error" class="p-4 bg-red-500/10 rounded border border-red-500/50">
            <p class="text-sm text-red-400">{{ plan.error }}</p>
            <p v-if="plan.error?.includes('MySQL and PostgreSQL only')" class="text-xs theme-text-muted mt-2">{{ t('explainPlan.unsupportedDriver') }}</p>
            <p v-else-if="plan.error?.includes('MySQL') && !plan.error?.includes('PostgreSQL')" class="text-xs theme-text-muted mt-2">{{ t('explainPlan.mysqlOnly') }}</p>
          </div>

          <div v-else-if="plan" class="space-y-6">
            <!-- Compare: Plan A vs Plan B -->
            <div v-if="compareMode && planA && planB" class="grid grid-cols-2 gap-4">
              <div>
                <h3 class="text-xs font-semibold theme-text-muted mb-2">{{ t('explainPlan.planA') }}</h3>
                <div class="space-y-0">
                  <div
                    v-for="(node, index) in planA.nodes"
                    :key="'a-' + node.id"
                    class="flex flex-col items-center"
                  >
                    <div v-if="index > 0" class="w-0.5 h-4 bg-[var(--border-strong)] flex-shrink-0" />
                    <div :class="['w-full rounded-lg border-2 px-3 py-2 text-xs transition-colors', nodeTypeColor(node)]">
                      <span class="font-mono theme-text">{{ node.label }}</span>
                      <span v-if="(node.rows ?? 0) > 0" class="theme-text-muted"> ~{{ (node.rows ?? 0).toLocaleString() }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div>
                <h3 class="text-xs font-semibold theme-text-muted mb-2">{{ t('explainPlan.planB') }}</h3>
                <div v-if="isLoadingB" class="flex justify-center py-8">
                  <div class="w-6 h-6 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin" />
                </div>
                <div v-else-if="planB?.error" class="p-3 rounded bg-red-500/10 border border-red-500/50 text-xs text-red-400">
                  {{ planB.error }}
                </div>
                <div v-else-if="planB?.nodes?.length" class="space-y-0">
                  <div
                    v-for="(node, index) in planB.nodes"
                    :key="'b-' + node.id"
                    class="flex flex-col items-center"
                  >
                    <div v-if="index > 0" class="w-0.5 h-4 bg-[var(--border-strong)] flex-shrink-0" />
                    <div :class="['w-full rounded-lg border-2 px-3 py-2 text-xs transition-colors', nodeTypeColor(node)]">
                      <span class="font-mono theme-text">{{ node.label }}</span>
                      <span v-if="(node.rows ?? 0) > 0" class="theme-text-muted"> ~{{ (node.rows ?? 0).toLocaleString() }}</span>
                    </div>
                  </div>
                </div>
                <div v-else class="py-4 text-xs theme-text-muted text-center">
                  —
                </div>
              </div>
            </div>

            <template v-else>
              <!-- Summary: total duration, total rows -->
              <div class="flex flex-wrap gap-4 text-xs theme-text-muted">
                <span v-if="plan.summary?.totalDurationMs != null">
                  {{ t('explainPlan.totalDuration') }}: {{ plan.summary.totalDurationMs }} ms
                </span>
                <span v-if="totalRows > 0">
                  {{ t('explainPlan.estRows') }}: {{ totalRows.toLocaleString() }}
                </span>
              </div>

              <!-- Flowchart: vertical chain of nodes -->
              <div class="space-y-0">
                <div
                  v-for="(node, index) in plan.nodes"
                  :key="node.id"
                  class="flex flex-col items-center"
                >
                  <div v-if="index > 0" class="w-0.5 h-6 bg-[var(--border-strong)] flex-shrink-0" />
                  <div
                    :class="[
                      'w-full max-w-md rounded-lg border-2 px-4 py-3 transition-colors',
                      nodeTypeColor(node)
                    ]"
                  >
                    <div class="flex items-center justify-between gap-4 flex-wrap">
                      <div class="flex items-center gap-2 flex-wrap">
                        <Search v-if="node.type === 'Scan'" :size="16" class="text-[#1677ff] flex-shrink-0" />
                        <span class="text-sm font-semibold theme-text">{{ node.type }}</span>
                        <span class="text-sm theme-text-muted">—</span>
                        <span class="text-sm font-mono theme-text">{{ node.label }}</span>
                        <span v-if="node.detail" class="text-xs theme-text-muted opacity-80">({{ node.detail }})</span>
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
                    <div v-if="node.rows !== undefined && node.rows > 0" class="mt-1 flex items-center gap-2">
                      <span class="text-xs theme-text-muted">
                        ~{{ node.rows.toLocaleString() }} {{ t('explainPlan.rows') }}
                      </span>
                      <span v-if="totalRows > 0" class="text-xs theme-text-muted">
                        ({{ nodeShare(node) }}%)
                      </span>
                      <div
                        v-if="totalRows > 0"
                        class="flex-1 h-1 max-w-24 rounded bg-[var(--border)] overflow-hidden"
                      >
                        <div
                          class="h-full bg-[#1677ff]/50 rounded"
                          :style="{ width: nodeShare(node) + '%' }"
                        />
                      </div>
                    </div>
                    <div v-if="node.extra" class="mt-1 text-xs theme-text-muted truncate" :title="node.extra">
                      {{ node.extra }}
                    </div>
                  </div>
                </div>
              </div>

              <!-- Warnings -->
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

              <!-- Index suggestions -->
              <div
                v-if="suggestions.length > 0"
                class="p-4 bg-emerald-500/10 rounded border border-emerald-500/50"
              >
                <div class="flex items-center gap-2 mb-2">
                  <Key :size="16" class="text-emerald-500" />
                  <span class="text-sm font-semibold text-emerald-600 dark:text-emerald-400">{{ t('explainPlan.suggestions') }}</span>
                </div>
                <ul class="space-y-2">
                  <li
                    v-for="(s, i) in suggestions"
                    :key="i"
                    class="text-xs theme-text flex items-start gap-2"
                  >
                    <span class="mt-0.5">•</span>
                    <div class="flex-1 min-w-0">
                      <p class="theme-text-muted">{{ s.reason }}</p>
                      <div class="flex items-center gap-2 mt-1">
                        <code class="px-1.5 py-0.5 rounded bg-black/10 dark:bg-white/10 font-mono text-[11px] break-all">{{ s.createIndex }}</code>
                        <button
                          type="button"
                          class="p-1 rounded theme-bg-hover"
                          @click="copySuggestionSql(s.createIndex)"
                        >
                          <Copy :size="12" class="theme-text-muted" />
                        </button>
                      </div>
                    </div>
                  </li>
                </ul>
              </div>
            </template>
          </div>
        </div>

        <div class="px-6 py-4 border-t theme-border flex justify-end theme-bg-footer">
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
