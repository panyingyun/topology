<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { GitCompare, X, Download } from 'lucide-vue-next'
import { dataService } from '../services/dataService'
import type { Connection } from '../types'

const { t } = useI18n()
const message = useMessage()

const props = defineProps<{
  show: boolean
  connections: Connection[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const connA = ref('')
const dbA = ref('')
const tableA = ref('')
const connB = ref('')
const dbB = ref('')
const tableB = ref('')
const dbsA = ref<string[]>([])
const dbsB = ref<string[]>([])
const tablesA = ref<string[]>([])
const tablesB = ref<string[]>([])
const loading = ref(false)
const hasRun = ref(false)
const diffRows = ref<{ status: 'same' | 'only_in_a' | 'only_in_b' | 'modified'; rowIndex: number; cells: Record<string, { a?: unknown; b?: unknown }> }[]>([])
const columns = ref<string[]>([])
const limit = 1000

watch(
  () => props.show,
  (v) => {
    if (v) {
      hasRun.value = false
      diffRows.value = []
      columns.value = []
      if (props.connections.length && !connA.value) connA.value = props.connections[0].id
      if (props.connections.length && !connB.value) {
        connB.value = props.connections.length > 1 ? props.connections[1].id : props.connections[0].id
      }
    }
  }
)

watch(connA, async (id) => {
  dbA.value = ''
  tableA.value = ''
  dbsA.value = []
  tablesA.value = []
  if (!id) return
  try {
    dbsA.value = await dataService.getDatabases(id, '')
  } catch {
    dbsA.value = []
  }
})
watch(dbA, async (db) => {
  tableA.value = ''
  tablesA.value = []
  if (!connA.value || !db) return
  try {
    const list = await dataService.getTables(connA.value, db, '')
    tablesA.value = list.map((x) => x.name)
  } catch {
    tablesA.value = []
  }
})
watch(connB, async (id) => {
  dbB.value = ''
  tableB.value = ''
  dbsB.value = []
  tablesB.value = []
  if (!id) return
  try {
    dbsB.value = await dataService.getDatabases(id, '')
  } catch {
    dbsB.value = []
  }
})
watch(dbB, async (db) => {
  tableB.value = ''
  tablesB.value = []
  if (!connB.value || !db) return
  try {
    const list = await dataService.getTables(connB.value, db, '')
    tablesB.value = list.map((x) => x.name)
  } catch {
    tablesB.value = []
  }
})

watch(
  () => [props.show, connA.value, connB.value],
  () => {
    if (props.show && connA.value) {
      dataService.getDatabases(connA.value, '').then((r) => { dbsA.value = r })
    }
    if (props.show && connB.value) {
      dataService.getDatabases(connB.value, '').then((r) => { dbsB.value = r })
    }
  }
)

function valStr(v: unknown): string {
  if (v == null) return ''
  const s = String(v)
  return s.length > 50 ? s.slice(0, 47) + '...' : s
}

function cellSame(a: unknown, b: unknown): boolean {
  const sa = valStr(a)
  const sb = valStr(b)
  return sa === sb
}

async function runCompare() {
  if (!connA.value || !dbA.value || !tableA.value || !connB.value || !dbB.value || !tableB.value) {
    message.warning(t('dataCompare.selectBoth'))
    return
  }
  loading.value = true
  hasRun.value = false
  diffRows.value = []
  columns.value = []
  try {
    const [dataA, dataB] = await Promise.all([
      dataService.getTableData(connA.value, dbA.value, tableA.value, limit, 0, ''),
      dataService.getTableData(connB.value, dbB.value, tableB.value, limit, 0, ''),
    ])
    const colsA = dataA?.columns ?? []
    const colsB = dataB?.columns ?? []
    const rowsA = dataA?.rows ?? []
    const rowsB = dataB?.rows ?? []
    const colSet = new Set<string>([...colsA, ...colsB])
    const colList = [...colsA.filter((c) => colSet.has(c)), ...colsB.filter((c) => !colsA.includes(c))]
    columns.value = colList

    const out: { status: 'same' | 'only_in_a' | 'only_in_b' | 'modified'; rowIndex: number; cells: Record<string, { a?: unknown; b?: unknown }> }[] = []
    const n = Math.max(rowsA.length, rowsB.length)
    for (let i = 0; i < n; i++) {
      const ra = rowsA[i] as Record<string, unknown> | undefined
      const rb = rowsB[i] as Record<string, unknown> | undefined
      const cells: Record<string, { a?: unknown; b?: unknown }> = {}
      for (const c of colList) {
        const va = ra?.[c]
        const vb = rb?.[c]
        cells[c] = { a: va, b: vb }
      }
      if (!ra && rb) {
        out.push({ status: 'only_in_b', rowIndex: i + 1, cells })
      } else if (ra && !rb) {
        out.push({ status: 'only_in_a', rowIndex: i + 1, cells })
      } else if (ra && rb) {
        let same = true
        for (const c of colList) {
          if (!cellSame(ra[c], rb[c])) {
            same = false
            break
          }
        }
        out.push({
          status: same ? 'same' : 'modified',
          rowIndex: i + 1,
          cells,
        })
      }
    }
    diffRows.value = out
    hasRun.value = true
  } catch (e) {
    message.error(e instanceof Error ? e.message : 'Compare failed')
  } finally {
    loading.value = false
  }
}

function exportReport() {
  if (!columns.value.length || !diffRows.value.length) return
  const lines: string[] = []
  const header = ['Status', 'Row', ...columns.value.flatMap((c) => [`${c} (A)`, `${c} (B)`])]
  lines.push(header.join('\t'))
  for (const r of diffRows.value) {
    const row: string[] = [r.status, String(r.rowIndex)]
    for (const c of columns.value) {
      const cell = r.cells[c]
      row.push(valStr(cell?.a), valStr(cell?.b))
    }
    lines.push(row.join('\t'))
  }
  const blob = new Blob([lines.join('\n')], { type: 'text/plain;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `data-compare-${tableA.value}-vs-${tableB.value}.tsv`
  a.click()
  URL.revokeObjectURL(a.href)
  message.success(t('dataCompare.reportExported'))
}
</script>

<template>
  <Transition name="fade">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <div class="theme-bg-panel rounded-lg border theme-border w-full max-w-6xl max-h-[90vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b theme-border flex items-center justify-between">
          <h2 class="text-lg font-semibold theme-text flex items-center gap-2">
            <GitCompare :size="20" class="text-[#1677ff]" />
            {{ t('dataCompare.title') }}
          </h2>
          <button class="p-1.5 theme-text-muted-hover rounded" @click="emit('close')">
            <X :size="20" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto p-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <h3 class="text-sm font-medium theme-text">{{ t('dataCompare.sourceA') }}</h3>
              <div class="flex flex-col gap-2">
                <select
                  v-model="connA"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('dataCompare.selectConnection') }}</option>
                  <option v-for="c in connections" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
                <select
                  v-model="dbA"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('dataCompare.selectDatabase') }}</option>
                  <option v-for="d in dbsA" :key="d" :value="d">{{ d }}</option>
                </select>
                <select
                  v-model="tableA"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('dataCompare.selectTable') }}</option>
                  <option v-for="tbl in tablesA" :key="tbl" :value="tbl">{{ tbl }}</option>
                </select>
              </div>
            </div>
            <div class="space-y-2">
              <h3 class="text-sm font-medium theme-text">{{ t('dataCompare.sourceB') }}</h3>
              <div class="flex flex-col gap-2">
                <select
                  v-model="connB"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('dataCompare.selectConnection') }}</option>
                  <option v-for="c in connections" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
                <select
                  v-model="dbB"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('dataCompare.selectDatabase') }}</option>
                  <option v-for="d in dbsB" :key="d" :value="d">{{ d }}</option>
                </select>
                <select
                  v-model="tableB"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('dataCompare.selectTable') }}</option>
                  <option v-for="tbl in tablesB" :key="tbl" :value="tbl">{{ tbl }}</option>
                </select>
              </div>
            </div>
          </div>

          <div class="flex flex-wrap gap-2">
            <button
              class="px-4 py-2 rounded text-sm font-medium bg-green-600 hover:bg-green-500 text-white disabled:opacity-50"
              :disabled="loading"
              @click="runCompare"
            >
              {{ t('dataCompare.compare') }}
            </button>
            <button
              v-if="diffRows.length"
              class="px-4 py-2 rounded text-sm font-medium theme-bg-input theme-bg-input-hover theme-text border theme-border"
              @click="exportReport"
            >
              <Download :size="14" class="inline mr-1" />
              {{ t('dataCompare.exportReport') }}
            </button>
          </div>

          <div v-if="loading" class="flex justify-center py-12">
            <div class="w-8 h-8 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin" />
          </div>

          <div v-else-if="diffRows.length" class="overflow-auto max-h-96 rounded border theme-border">
            <table class="w-full text-sm border-collapse">
              <thead class="theme-bg-input sticky top-0">
                <tr>
                  <th class="text-left px-2 py-1.5 border-b theme-border font-medium">{{ t('dataCompare.status') }}</th>
                  <th class="text-left px-2 py-1.5 border-b theme-border font-medium">#</th>
                  <th
                    v-for="col in columns"
                    :key="col"
                    class="text-left px-2 py-1.5 border-b theme-border font-medium"
                  >
                    {{ col }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(r, idx) in diffRows"
                  :key="idx"
                  :class="{
                    'bg-emerald-500/10': r.status === 'same',
                    'bg-amber-500/10': r.status === 'modified',
                    'bg-blue-500/10': r.status === 'only_in_a',
                    'bg-purple-500/10': r.status === 'only_in_b',
                  }"
                >
                  <td class="px-2 py-1 border-b theme-border">
                    <span
                      :class="{
                        'text-emerald-600 dark:text-emerald-400': r.status === 'same',
                        'text-amber-600 dark:text-amber-400': r.status === 'modified',
                        'text-blue-600 dark:text-blue-400': r.status === 'only_in_a',
                        'text-purple-600 dark:text-purple-400': r.status === 'only_in_b',
                      }"
                    >
                      {{ r.status === 'same' ? t('dataCompare.same') : r.status === 'modified' ? t('dataCompare.modified') : r.status === 'only_in_a' ? t('dataCompare.onlyInA') : t('dataCompare.onlyInB') }}
                    </span>
                  </td>
                  <td class="px-2 py-1 border-b theme-border theme-text-muted">{{ r.rowIndex }}</td>
                  <td
                    v-for="col in columns"
                    :key="col"
                    class="px-2 py-1 border-b theme-border"
                  >
                    <template v-if="r.status === 'same' || r.status === 'only_in_a'">
                      {{ valStr(r.cells[col]?.a) }}
                    </template>
                    <template v-else-if="r.status === 'only_in_b'">
                      {{ valStr(r.cells[col]?.b) }}
                    </template>
                    <template v-else>
                      <span v-if="!cellSame(r.cells[col]?.a, r.cells[col]?.b)" class="inline-block">
                        <span class="text-amber-600 dark:text-amber-400" :title="valStr(r.cells[col]?.b)">{{ valStr(r.cells[col]?.a) }}</span>
                        <span class="theme-text-muted mx-1">→</span>
                        <span class="text-amber-600 dark:text-amber-400">{{ valStr(r.cells[col]?.b) }}</span>
                      </span>
                      <span v-else>{{ valStr(r.cells[col]?.a) }}</span>
                    </template>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <p v-else-if="hasRun && !loading && diffRows.length === 0" class="text-sm theme-text-muted">
            {{ t('dataCompare.noDiff') }}
          </p>
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
