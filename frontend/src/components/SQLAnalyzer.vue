<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { Sparkles, AlertTriangle, Lightbulb, TrendingUp, X } from 'lucide-vue-next'
import { schemaService } from '../services/schemaService'
import type { SQLAnalysis, DatabaseType } from '../types'

const { t } = useI18n()
const message = useMessage()

const props = defineProps<{
  show: boolean
  sql: string
  driver?: DatabaseType
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const analysis = ref<SQLAnalysis | null>(null)
const isLoading = ref(false)

const analyze = async () => {
  if (!props.sql.trim()) {
    message.warning(t('analyzer.enterSQL'))
    return
  }
  isLoading.value = true
  try {
    analysis.value = await schemaService.analyzeSQL(props.sql, props.driver || 'mysql')
  } catch (error) {
    console.error('Failed to analyze SQL:', error)
    message.error(t('analyzer.analyzeFailed') + ': ' + (error instanceof Error ? error.message : 'Unknown error'))
  } finally {
    isLoading.value = false
  }
}

const complexityColor = computed(() => {
  if (!analysis.value) return 'theme-text-muted'
  const complexity = analysis.value.performance?.estimatedComplexity
  if (complexity === 'low') return 'text-green-400'
  if (complexity === 'medium') return 'text-yellow-400'
  return 'text-red-400'
})

watch(() => props.show, (newVal) => {
  if (newVal && props.sql) {
    analyze()
  }
})
</script>

<template>
  <Transition name="fade">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <div class="theme-bg-panel rounded-lg border theme-border w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
        <!-- Header -->
        <div class="px-6 py-4 border-b theme-border flex items-center justify-between">
          <h2 class="text-lg font-semibold theme-text flex items-center gap-2">
            <Sparkles :size="20" class="text-[#1677ff]" />
            {{ t('analyzer.title') }}
          </h2>
          <button
            @click="emit('close')"
            class="theme-text-muted-hover transition-colors"
          >
            <X :size="20" />
          </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <div v-if="isLoading" class="flex items-center justify-center h-64">
            <div class="text-center">
              <div class="w-8 h-8 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
              <p class="text-sm theme-text-muted">{{ t('analyzer.analyzing') }}</p>
            </div>
          </div>

          <div v-else-if="analysis" class="space-y-4">
            <!-- Query Type -->
            <div class="p-4 theme-bg-content rounded border theme-border">
              <div class="flex items-center gap-2 mb-2">
                <TrendingUp :size="16" class="text-[#1677ff]" />
                <span class="text-sm font-semibold theme-text">{{ t('analyzer.queryType') }}</span>
              </div>
              <span class="text-xs theme-text-muted uppercase">{{ analysis.queryType }}</span>
            </div>

            <!-- Performance -->
            <div class="p-4 theme-bg-content rounded border theme-border">
              <div class="flex items-center gap-2 mb-3">
                <TrendingUp :size="16" class="text-[#1677ff]" />
                <span class="text-sm font-semibold theme-text">{{ t('analyzer.performance') }}</span>
              </div>
              <div class="space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-xs theme-text-muted">{{ t('analyzer.complexity') }}</span>
                  <span :class="['text-xs font-semibold uppercase', complexityColor]">
                    {{ analysis.performance?.estimatedComplexity || 'unknown' }}
                  </span>
                </div>
                <div v-if="analysis.performance?.indexUsage" class="text-xs theme-text-muted">
                  {{ analysis.performance.indexUsage }}
                </div>
              </div>
            </div>

            <!-- Warnings -->
            <div v-if="analysis.warnings && analysis.warnings.length > 0" class="p-4 bg-red-500/10 rounded border border-red-500/50">
              <div class="flex items-center gap-2 mb-3">
                <AlertTriangle :size="16" class="text-red-400" />
                <span class="text-sm font-semibold text-red-400">{{ t('analyzer.warnings') }}</span>
              </div>
              <ul class="space-y-1">
                <li
                  v-for="(warning, idx) in analysis.warnings"
                  :key="idx"
                  class="text-xs text-red-300 flex items-start gap-2"
                >
                  <span class="mt-0.5">•</span>
                  <span>{{ warning }}</span>
                </li>
              </ul>
            </div>

            <!-- Suggestions -->
            <div v-if="analysis.suggestions && analysis.suggestions.length > 0" class="p-4 bg-yellow-500/10 rounded border border-yellow-500/50">
              <div class="flex items-center gap-2 mb-3">
                <Lightbulb :size="16" class="text-yellow-400" />
                <span class="text-sm font-semibold text-yellow-400">{{ t('analyzer.suggestions') }}</span>
              </div>
              <ul class="space-y-1">
                <li
                  v-for="(suggestion, idx) in analysis.suggestions"
                  :key="idx"
                  class="text-xs text-yellow-300 flex items-start gap-2"
                >
                  <span class="mt-0.5">•</span>
                  <span>{{ suggestion }}</span>
                </li>
              </ul>
            </div>

            <div v-if="analysis.warnings?.length === 0 && analysis.suggestions?.length === 0" class="p-4 bg-green-500/10 rounded border border-green-500/50 text-center">
              <p class="text-xs text-green-400">{{ t('analyzer.noIssues') }}</p>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t theme-border flex items-center justify-end theme-bg-footer">
          <button
            @click="emit('close')"
            class="px-6 py-2 rounded text-xs font-semibold bg-[#1677ff] hover:bg-[#4096ff] text-white transition-colors"
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
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
