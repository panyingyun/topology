<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import {
  backupService,
  type BackupRecord,
  type BackupSchedule,
} from '../services/backupService'
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

const tab = ref<'list' | 'schedules'>('list')
const backups = ref<BackupRecord[]>([])
const schedules = ref<BackupSchedule[]>([])
const loading = ref(false)
const verifyCache = ref<Record<string, { exists: boolean; size: number }>>({})
const connMap = computed(() => {
  const m: Record<string, string> = {}
  for (const c of props.connections) m[c.id] = c.name
  return m
})

const dayNames = [
  t('backup.never'),
  'Sun',
  'Mon',
  'Tue',
  'Wed',
  'Thu',
  'Fri',
  'Sat',
]
const zhDays = ['日', '一', '二', '三', '四', '五', '六']

function loadBackups() {
  loading.value = true
  backupService
    .listBackups('')
    .then((r) => { backups.value = r })
    .finally(() => { loading.value = false })
}

function loadSchedules() {
  loading.value = true
  backupService
    .getSchedules()
    .then((r) => { schedules.value = r })
    .finally(() => { loading.value = false })
}

watch(
  () => props.show,
  (v) => {
    if (v) {
      loadBackups()
      loadSchedules()
      verifyCache.value = {}
    }
  }
)

async function verify(path: string) {
  const v = await backupService.verifyBackup(path)
  verifyCache.value[path] = v
}

function verifiedInfo(path: string) {
  const v = verifyCache.value[path]
  if (!v) return null
  if (!v.exists) return 'missing'
  return `${(v.size / 1024).toFixed(1)} KB`
}

async function removeBackup(r: BackupRecord) {
  if (!confirm(t('backup.delete') + '?\n' + r.path)) return
  const res = await backupService.deleteBackup(r.path)
  if (res.success) {
    message.success(t('common.success'))
    loadBackups()
    delete verifyCache.value[r.path]
  } else {
    message.error((res.error || '') + '')
  }
}

function addSchedule() {
  const conn = props.connections.find(
    (c) => ['mysql', 'postgresql', 'postgres', 'sqlite'].includes(c.type)
  )
  if (!conn) {
    message.warning('No MySQL/PostgreSQL/SQLite connection')
    return
  }
  const s: BackupSchedule = {
    connectionId: conn.id,
    enabled: true,
    schedule: 'daily',
    time: '02:00',
    day: 0,
    outputDir: '',
  }
  schedules.value = [...schedules.value, s]
}

function removeSchedule(i: number) {
  schedules.value = schedules.value.filter((_, j) => j !== i)
}

async function saveSchedules() {
  try {
    await backupService.setSchedules(schedules.value)
    message.success(t('common.success'))
  } catch (e) {
    message.error(t('common.error') + ': ' + (e instanceof Error ? e.message : ''))
  }
}

function scheduleDayLabel(d: number) {
  const locale = (typeof navigator !== 'undefined' && navigator.language) || 'en'
  return locale.startsWith('zh') ? `周${zhDays[d]}` : dayNames[d + 1]
}

function close() {
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="show"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
        @click.self="close"
      >
        <div
          class="theme-bg-panel rounded-lg border theme-border w-full max-w-2xl shadow-xl flex flex-col max-h-[85vh]"
          @click.stop
        >
          <div class="px-4 py-3 border-b theme-border flex items-center justify-between">
            <h3 class="text-sm font-semibold theme-text">{{ t('backup.manage') }}</h3>
            <button class="theme-text-muted hover:theme-text p-1 rounded" @click="close">
              ✕
            </button>
          </div>
          <div class="flex border-b theme-border">
            <button
              :class="['px-4 py-2 text-xs font-medium', tab === 'list' ? 'theme-text border-b-2 border-[#1677ff]' : 'theme-text-muted theme-bg-hover']"
              @click="tab = 'list'"
            >
              {{ t('backup.list') }}
            </button>
            <button
              :class="['px-4 py-2 text-xs font-medium', tab === 'schedules' ? 'theme-text border-b-2 border-[#1677ff]' : 'theme-text-muted theme-bg-hover']"
              @click="tab = 'schedules'"
            >
              {{ t('backup.schedules') }}
            </button>
          </div>
          <div class="flex-1 overflow-y-auto p-4">
            <template v-if="tab === 'list'">
              <div v-if="loading" class="text-xs theme-text-muted">...</div>
              <ul v-else-if="backups.length" class="space-y-2">
                <li
                  v-for="r in backups"
                  :key="r.path + r.at"
                  class="flex items-center gap-2 flex-wrap rounded border theme-border p-2 text-xs"
                >
                  <span class="font-mono truncate flex-1 min-w-0" :title="r.path">{{ r.path }}</span>
                  <span class="theme-text-muted shrink-0">{{ connMap[r.connectionId] || r.connectionId }}</span>
                  <span class="theme-text-muted shrink-0">{{ r.at }}</span>
                  <span v-if="verifiedInfo(r.path)" class="shrink-0" :class="verifiedInfo(r.path) === 'missing' ? 'text-red-400' : 'theme-text-muted'">
                    {{ verifiedInfo(r.path) === 'missing' ? 'missing' : verifiedInfo(r.path) }}
                  </span>
                  <button
                    class="px-2 py-0.5 rounded theme-bg-input theme-bg-input-hover theme-text"
                    @click="verify(r.path)"
                  >
                    {{ t('backup.verify') }}
                  </button>
                  <button
                    class="px-2 py-0.5 rounded bg-red-600/80 hover:bg-red-500 text-white"
                    @click="removeBackup(r)"
                  >
                    {{ t('backup.delete') }}
                  </button>
                </li>
              </ul>
              <p v-else class="text-xs theme-text-muted">{{ t('backup.noBackups') }}</p>
            </template>
            <template v-else>
              <div class="flex justify-end mb-2">
                <button
                  class="px-3 py-1 rounded text-xs bg-[#1677ff] hover:bg-[#4096ff] text-white"
                  @click="addSchedule"
                >
                  {{ t('backup.addSchedule') }}
                </button>
              </div>
              <div v-if="loading" class="text-xs theme-text-muted">...</div>
              <ul v-else class="space-y-3">
                <li
                  v-for="(s, i) in schedules"
                  :key="i"
                  class="rounded border theme-border p-3 text-xs space-y-2"
                >
                  <div class="flex items-center gap-2 flex-wrap">
                    <select
                      v-model="s.connectionId"
                      class="theme-bg-input theme-text rounded px-2 py-1 border theme-border"
                    >
                      <option
                        v-for="c in connections.filter((x) => ['mysql','postgresql','postgres','sqlite'].includes(x.type))"
                        :key="c.id"
                        :value="c.id"
                      >
                        {{ c.name }}
                      </option>
                    </select>
                    <label class="flex items-center gap-1">
                      <input v-model="s.enabled" type="checkbox" />
                      {{ s.enabled ? 'ON' : 'OFF' }}
                    </label>
                    <select
                      v-model="s.schedule"
                      class="theme-bg-input theme-text rounded px-2 py-1 border theme-border"
                    >
                      <option value="daily">{{ t('backup.daily') }}</option>
                      <option value="weekly">{{ t('backup.weekly') }}</option>
                    </select>
                    <input
                      v-model="s.time"
                      type="text"
                      placeholder="02:00"
                      class="theme-bg-input theme-text rounded px-2 py-1 border theme-border w-16"
                    />
                    <template v-if="s.schedule === 'weekly'">
                      <select
                        v-model.number="s.day"
                        class="theme-bg-input theme-text rounded px-2 py-1 border theme-border"
                      >
                        <option v-for="d in 7" :key="d - 1" :value="d - 1">{{ scheduleDayLabel(d - 1) }}</option>
                      </select>
                    </template>
                    <input
                      v-model="s.outputDir"
                      type="text"
                      :placeholder="t('backup.outputDir')"
                      class="theme-bg-input theme-text rounded px-2 py-1 border theme-border flex-1 min-w-[120px]"
                    />
                    <span class="theme-text-muted">{{ t('backup.lastRun') }}: {{ s.lastRun || t('backup.never') }}</span>
                    <button
                      class="px-2 py-0.5 rounded bg-red-600/80 hover:bg-red-500 text-white"
                      @click="removeSchedule(i)"
                    >
                      {{ t('backup.delete') }}
                    </button>
                  </div>
                </li>
              </ul>
              <div v-if="schedules.length" class="mt-3">
                <button
                  class="px-3 py-1.5 rounded text-xs bg-green-600 hover:bg-green-500 text-white"
                  @click="saveSchedules"
                >
                  {{ t('common.save') }}
                </button>
              </div>
            </template>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
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
