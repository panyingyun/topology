<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { backupService, type BackupRecord } from '../services/backupService'

const { t } = useI18n()
const message = useMessage()

const props = defineProps<{
  show: boolean
  connectionId: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const backups = ref<BackupRecord[]>([])
const selectedPath = ref('')
const step = ref<'select' | 'confirm'>('select')
const loading = ref(false)
const restoring = ref(false)

const load = async () => {
  if (!props.connectionId) return
  loading.value = true
  try {
    backups.value = await backupService.listBackups(props.connectionId)
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.connectionId],
  () => {
    if (props.show && props.connectionId) {
      load()
      selectedPath.value = ''
      step.value = 'select'
    }
  }
)

const pickFile = async () => {
  const path = await backupService.pickBackupFile()
  if (path) {
    selectedPath.value = path
    step.value = 'confirm'
  }
}

const selectBackup = (r: BackupRecord) => {
  selectedPath.value = r.path
  step.value = 'confirm'
}

const back = () => {
  step.value = 'select'
  selectedPath.value = ''
}

const confirmRestore = async () => {
  if (!selectedPath.value) return
  restoring.value = true
  try {
    const res = await backupService.restoreBackup(props.connectionId, selectedPath.value)
    if (res.success) {
      message.success(t('backup.restoreSuccess'))
      emit('close')
    } else {
      message.error(t('backup.restoreFailed') + ': ' + (res.error || ''))
    }
  } finally {
    restoring.value = false
  }
}

const close = () => {
  if (!restoring.value) emit('close')
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
          class="theme-bg-panel rounded-lg border theme-border w-full max-w-lg shadow-xl flex flex-col max-h-[80vh]"
          @click.stop
        >
          <div class="px-4 py-3 border-b theme-border flex items-center justify-between">
            <h3 class="text-sm font-semibold theme-text">{{ t('backup.restoreBackup') }}</h3>
            <button
              class="theme-text-muted hover:theme-text p-1 rounded"
              :disabled="restoring"
              @click="close"
            >
              ✕
            </button>
          </div>
          <div class="flex-1 overflow-y-auto p-4 space-y-4">
            <template v-if="step === 'select'">
              <div class="flex gap-2">
                <button
                  class="px-3 py-1.5 rounded text-xs font-medium bg-[#1677ff] hover:bg-[#4096ff] text-white"
                  @click="pickFile"
                >
                  {{ t('backup.pickFile') }}
                </button>
              </div>
              <p class="text-xs theme-text-muted">{{ t('backup.recentBackups') }}</p>
              <div v-if="loading" class="text-xs theme-text-muted">...</div>
              <ul v-else-if="backups.length" class="space-y-1">
                <li
                  v-for="r in backups"
                  :key="r.path + r.at"
                  class="flex flex-col gap-0.5 px-3 py-2 rounded theme-bg-hover cursor-pointer text-left"
                  @click="selectBackup(r)"
                >
                  <span class="text-xs font-mono truncate theme-text" :title="r.path">{{ r.path }}</span>
                  <span class="text-[10px] theme-text-muted">{{ r.at }}</span>
                </li>
              </ul>
              <p v-else class="text-xs theme-text-muted">{{ t('backup.noBackups') }}</p>
            </template>
            <template v-else>
              <p class="text-xs theme-text-muted">{{ t('backup.confirmRestore') }}</p>
              <p class="text-xs font-mono truncate theme-text" :title="selectedPath">{{ selectedPath }}</p>
              <div class="flex gap-2">
                <button
                  class="px-3 py-1.5 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
                  :disabled="restoring"
                  @click="back"
                >
                  {{ t('common.cancel') }}
                </button>
                <button
                  class="px-3 py-1.5 rounded text-xs font-medium bg-red-600 hover:bg-red-500 text-white disabled:opacity-50"
                  :disabled="restoring"
                  @click="confirmRestore"
                >
                  {{ restoring ? '...' : t('backup.confirmRestoreButton') }}
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
