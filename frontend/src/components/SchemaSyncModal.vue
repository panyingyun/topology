<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { GitMerge, X, Copy } from 'lucide-vue-next'
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
const script = ref('')

watch(
  () => props.show,
  (v) => {
    if (v) {
      script.value = ''
      if (props.connections.length && !connA.value) connA.value = props.connections[0].id
      if (props.connections.length && !connB.value && props.connections.length > 1) {
        connB.value = props.connections[1].id
      } else if (props.connections.length && !connB.value) {
        connB.value = props.connections[0].id
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

async function generate(direction: 'a_to_b' | 'b_to_a') {
  if (!connA.value || !dbA.value || !tableA.value || !connB.value || !dbB.value || !tableB.value) {
    message.warning(t('schemaSync.selectBoth'))
    return
  }
  loading.value = true
  script.value = ''
  try {
    script.value = await dataService.getSchemaSyncScript(
      connA.value,
      dbA.value,
      tableA.value,
      connB.value,
      dbB.value,
      tableB.value,
      direction
    )
  } catch (e) {
    script.value = '-- Error: ' + (e instanceof Error ? e.message : 'Unknown error')
  } finally {
    loading.value = false
  }
}

async function copyScript() {
  if (!script.value) return
  try {
    await navigator.clipboard.writeText(script.value)
    message.success(t('schemaSync.copied'))
  } catch {
    message.error(t('common.error'))
  }
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
        <div class="px-6 py-4 border-b theme-border flex items-center justify-between">
          <h2 class="text-lg font-semibold theme-text flex items-center gap-2">
            <GitMerge :size="20" class="text-[#1677ff]" />
            {{ t('schemaSync.title') }}
          </h2>
          <button class="p-1.5 theme-text-muted-hover rounded" @click="emit('close')">
            <X :size="20" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto p-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-2">
              <h3 class="text-sm font-medium theme-text">{{ t('schemaSync.sourceA') }}</h3>
              <div class="flex flex-col gap-2">
                <select
                  v-model="connA"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('schemaSync.selectConnection') }}</option>
                  <option v-for="c in connections" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
                <select
                  v-model="dbA"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('schemaSync.selectDatabase') }}</option>
                  <option v-for="d in dbsA" :key="d" :value="d">{{ d }}</option>
                </select>
                <select
                  v-model="tableA"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('schemaSync.selectTable') }}</option>
                  <option v-for="tbl in tablesA" :key="tbl" :value="tbl">{{ tbl }}</option>
                </select>
              </div>
            </div>
            <div class="space-y-2">
              <h3 class="text-sm font-medium theme-text">{{ t('schemaSync.sourceB') }}</h3>
              <div class="flex flex-col gap-2">
                <select
                  v-model="connB"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('schemaSync.selectConnection') }}</option>
                  <option v-for="c in connections" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
                <select
                  v-model="dbB"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('schemaSync.selectDatabase') }}</option>
                  <option v-for="d in dbsB" :key="d" :value="d">{{ d }}</option>
                </select>
                <select
                  v-model="tableB"
                  class="theme-bg-input theme-text rounded border theme-border px-3 py-2 text-sm"
                >
                  <option value="">{{ t('schemaSync.selectTable') }}</option>
                  <option v-for="tbl in tablesB" :key="tbl" :value="tbl">{{ tbl }}</option>
                </select>
              </div>
            </div>
          </div>

          <div class="flex flex-wrap gap-2">
            <button
              class="px-4 py-2 rounded text-sm font-medium bg-[#1677ff] hover:bg-[#4096ff] text-white disabled:opacity-50"
              :disabled="loading"
              @click="generate('a_to_b')"
            >
              {{ t('schemaSync.generateAtoB') }}
            </button>
            <button
              class="px-4 py-2 rounded text-sm font-medium bg-[#1677ff] hover:bg-[#4096ff] text-white disabled:opacity-50"
              :disabled="loading"
              @click="generate('b_to_a')"
            >
              {{ t('schemaSync.generateBtoA') }}
            </button>
          </div>

          <div v-if="script" class="mt-4">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium theme-text">{{ t('schemaSync.script') }}</span>
              <button
                class="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs theme-bg-input theme-bg-input-hover theme-text"
                @click="copyScript"
              >
                <Copy :size="14" />
                {{ t('schemaSync.copy') }}
              </button>
            </div>
            <pre
              class="p-4 rounded border theme-border theme-bg-input theme-text text-xs font-mono overflow-auto max-h-80 custom-scrollbar whitespace-pre-wrap"
            >{{ script }}</pre>
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
