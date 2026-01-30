<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { GitBranch, X, Download, Image as ImageIcon } from 'lucide-vue-next'
import { dataService } from '../services/dataService'
import type { TableSchema } from '../types'

const { t } = useI18n()
const message = useMessage()

const props = defineProps<{
  show: boolean
  connectionId: string
  database: string
  sessionId?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const tables = ref<TableSchema[]>([])
const error = ref('')
const isLoading = ref(false)
const svgRef = ref<SVGSVGElement | null>(null)

const BOX_WIDTH = 200
const BOX_HEIGHT = 160
const COLS = 4
const PAD = 60
const GAP = 80

const layout = computed(() => {
  const list = tables.value
  const positions: { [k: string]: { x: number; y: number } } = {}
  for (let i = 0; i < list.length; i++) {
    const c = i % COLS
    const r = Math.floor(i / COLS)
    const x = PAD + c * (BOX_WIDTH + GAP)
    const y = PAD + r * (BOX_HEIGHT + GAP)
    positions[list[i].name] = { x, y }
  }
  return positions
})

const relations = computed(() => {
  const out: { from: string; to: string; cols: string[] }[] = []
  for (const t of tables.value) {
    for (const fk of t.foreignKeys || []) {
      if (fk.referencedTable && tables.value.some((x) => x.name === fk.referencedTable)) {
        out.push({ from: t.name, to: fk.referencedTable, cols: fk.columns || [] })
      }
    }
  }
  return out
})

const svgSize = computed(() => {
  const list = tables.value
  if (!list.length) return { w: 800, h: 400 }
  const rows = Math.ceil(list.length / COLS)
  const w = PAD * 2 + COLS * BOX_WIDTH + (COLS - 1) * GAP
  const h = PAD * 2 + rows * BOX_HEIGHT + (rows - 1) * GAP
  return { w: Math.max(800, w), h: Math.max(400, h) }
})

async function load() {
  if (!props.connectionId || !props.database) return
  isLoading.value = true
  error.value = ''
  tables.value = []
  try {
    const res = await dataService.getERMetadata(
      props.connectionId,
      props.database,
      props.sessionId || '',
      '[]'
    )
    if (res.error) {
      error.value = res.error
      return
    }
    tables.value = res.tables || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load ER metadata'
  } finally {
    isLoading.value = false
  }
}

watch(
  () => [props.show, props.connectionId, props.database] as const,
  ([show]) => {
    if (show) load()
  }
)

function center(tableName: string) {
  const pos = layout.value[tableName]
  if (!pos) return { x: 0, y: 0 }
  return {
    x: pos.x + BOX_WIDTH / 2,
    y: pos.y + BOX_HEIGHT / 2,
  }
}

function exportSVG() {
  const el = svgRef.value
  if (!el) return
  const s = new XMLSerializer().serializeToString(el)
  const blob = new Blob([s], { type: 'image/svg+xml;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `er-${props.database || 'schema'}.svg`
  a.click()
  URL.revokeObjectURL(a.href)
  message.success(t('erDiagram.exported') + ' SVG')
}

function exportPNG() {
  const el = svgRef.value
  if (!el) return
  const { w, h } = svgSize.value
  const s = new XMLSerializer().serializeToString(el)
  const img = new window.Image()
  const blob = new Blob([s], { type: 'image/svg+xml;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  img.onload = () => {
    const canvas = document.createElement('canvas')
    canvas.width = w
    canvas.height = h
    const ctx = canvas.getContext('2d')
    if (!ctx) {
      URL.revokeObjectURL(url)
      return
    }
    ctx.fillStyle = '#fff'
    ctx.fillRect(0, 0, w, h)
    ctx.drawImage(img as CanvasImageSource, 0, 0)
    canvas.toBlob(
      (b) => {
        URL.revokeObjectURL(url)
        if (!b) return
        const a = document.createElement('a')
        a.href = URL.createObjectURL(b)
        a.download = `er-${props.database || 'schema'}.png`
        a.click()
        URL.revokeObjectURL(a.href)
        message.success(t('erDiagram.exported') + ' PNG')
      },
      'image/png'
    )
  }
  img.onerror = () => URL.revokeObjectURL(url)
  img.src = url
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
            <GitBranch :size="20" class="text-[#1677ff]" />
            {{ t('erDiagram.title') }} — {{ database || '-' }}
          </h2>
          <div class="flex items-center gap-2">
            <button
              v-if="tables.length > 0"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
              @click="exportSVG"
            >
              <Download :size="14" />
              {{ t('erDiagram.exportSvg') }}
            </button>
            <button
              v-if="tables.length > 0"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
              @click="exportPNG"
            >
              <ImageIcon :size="14" />
              {{ t('erDiagram.exportPng') }}
            </button>
            <button class="p-1.5 theme-text-muted-hover rounded" @click="emit('close')">
              <X :size="20" />
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-auto custom-scrollbar p-4">
          <div v-if="isLoading" class="flex justify-center items-center h-64">
            <div class="w-8 h-8 border-2 border-[#1677ff] border-t-transparent rounded-full animate-spin" />
          </div>
          <div v-else-if="error" class="p-4 bg-red-500/10 rounded border border-red-500/50">
            <p class="text-sm text-red-400">{{ error }}</p>
          </div>
          <div v-else-if="tables.length === 0" class="flex justify-center items-center h-64 theme-text-muted">
            {{ t('erDiagram.noTables') }}
          </div>
          <div v-else class="inline-block">
            <svg
              ref="svgRef"
              :width="svgSize.w"
              :height="svgSize.h"
              xmlns="http://www.w3.org/2000/svg"
              class="border theme-border rounded bg-[#fafafa] dark:bg-[#111]"
            >
              <!-- FK lines -->
              <g v-for="(rel, i) in relations" :key="i">
                <line
                  :x1="center(rel.from).x"
                  :y1="center(rel.from).y"
                  :x2="center(rel.to).x"
                  :y2="center(rel.to).y"
                  stroke="var(--border-strong)"
                  stroke-width="1.5"
                  marker-end="url(#arrow)"
                />
              </g>
              <defs>
                <marker id="arrow" markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto">
                  <path d="M0,0 L8,4 L0,8 Z" fill="var(--border-strong)" />
                </marker>
              </defs>
              <!-- Table boxes -->
              <g v-for="tbl in tables" :key="tbl.name">
                <rect
                  :x="layout[tbl.name]?.x ?? 0"
                  :y="layout[tbl.name]?.y ?? 0"
                  :width="BOX_WIDTH"
                  :height="BOX_HEIGHT"
                  rx="6"
                  fill="var(--panel)"
                  stroke="var(--border-strong)"
                  stroke-width="1.5"
                />
                <text
                  :x="(layout[tbl.name]?.x ?? 0) + BOX_WIDTH / 2"
                  :y="(layout[tbl.name]?.y ?? 0) + 24"
                  text-anchor="middle"
                  class="fill-current"
                  style="font-size: 13px; font-weight: 600;"
                >
                  {{ tbl.name }}
                </text>
                <line
                  :x1="layout[tbl.name]?.x ?? 0"
                  :y1="(layout[tbl.name]?.y ?? 0) + 32"
                  :x2="(layout[tbl.name]?.x ?? 0) + BOX_WIDTH"
                  :y2="(layout[tbl.name]?.y ?? 0) + 32"
                  stroke="var(--border-strong)"
                  stroke-width="1"
                />
                <g v-for="(col, j) in (tbl.columns || []).slice(0, 6)" :key="col.name">
                  <text
                    :x="(layout[tbl.name]?.x ?? 0) + 10"
                    :y="(layout[tbl.name]?.y ?? 0) + 50 + j * 16"
                    class="fill-current theme-text-muted"
                    style="font-size: 11px; font-family: ui-monospace, monospace;"
                  >
                    {{ col.isPrimaryKey ? '▪ ' : '' }}{{ col.name }}
                  </text>
                </g>
                <text
                  v-if="(tbl.columns?.length ?? 0) > 6"
                  :x="(layout[tbl.name]?.x ?? 0) + 10"
                  :y="(layout[tbl.name]?.y ?? 0) + 50 + 6 * 16"
                  class="fill-current theme-text-muted"
                  style="font-size: 10px;"
                >
                  …
                </text>
              </g>
            </svg>
          </div>
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
