<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Database, CheckCircle, XCircle, Loader } from 'lucide-vue-next'
import { connectionService } from '../services/connectionService'
import type { Connection, DatabaseType } from '../types'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'connect', connection: Connection): void
}>()

const activeDbType = ref<DatabaseType>('mysql')
const testStatus = ref<'idle' | 'loading' | 'success' | 'error'>('idle')
const errorMessage = ref('')

const form = reactive({
  name: '',
  host: '127.0.0.1',
  port: 3306,
  username: 'root',
  password: '',
  database: '',
  useSSL: false,
})

const handleTypeChange = (type: DatabaseType) => {
  activeDbType.value = type
  if (type === 'mysql') {
    form.port = 3306
  } else if (type === 'postgresql') {
    form.port = 5432
  } else if (type === 'sqlite') {
    form.port = 0
  }
}

const testConnection = async () => {
  testStatus.value = 'loading'
  errorMessage.value = ''
  
  try {
    const result = await connectionService.testConnection({
      ...form,
      type: activeDbType.value,
    } as any)
    
    if (result) {
      testStatus.value = 'success'
      setTimeout(() => {
        testStatus.value = 'idle'
      }, 2000)
    } else {
      testStatus.value = 'error'
      errorMessage.value = 'Connection failed'
    }
  } catch (error) {
    testStatus.value = 'error'
    errorMessage.value = error instanceof Error ? error.message : 'Connection failed'
  }
}

const handleConnect = async () => {
  try {
    const connection = await connectionService.createConnection({
      name: form.name,
      type: activeDbType.value,
      host: form.host,
      port: form.port,
      username: form.username,
      password: form.password,
      database: form.database || undefined,
      useSSL: form.useSSL,
    })
    emit('connect', connection)
    emit('close')
  } catch (error) {
    console.error('Failed to create connection:', error)
  }
}

const dbTypes: Array<{ type: DatabaseType; label: string; icon: string; color: string }> = [
  { type: 'mysql', label: 'MySQL', icon: '🐬', color: '#4479a1' },
  { type: 'postgresql', label: 'PostgreSQL', icon: '🐘', color: '#336791' },
  { type: 'sqlite', label: 'SQLite', icon: '🗄️', color: '#003b57' },
]
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
          <h2 class="text-lg font-semibold text-gray-200">New Connection</h2>
          <button
            @click="emit('close')"
            class="text-gray-400 hover:text-gray-200 transition-colors"
          >
            <XCircle :size="20" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <!-- Database Type Selection -->
          <div class="mb-6">
            <label class="block text-xs font-semibold text-gray-400 mb-3 uppercase tracking-wider">
              Database Type
            </label>
            <div class="grid grid-cols-3 gap-4">
              <div
                v-for="db in dbTypes"
                :key="db.type"
                @click="handleTypeChange(db.type)"
                :class="[
                  'p-4 rounded-lg border-2 cursor-pointer transition-all',
                  activeDbType === db.type
                    ? 'border-[#1677ff] bg-[#1677ff]/10'
                    : 'border-[#333] bg-[#2d2d30] hover:border-[#444]'
                ]"
              >
                <div class="text-3xl mb-2">{{ db.icon }}</div>
                <div class="text-sm font-semibold text-gray-200">{{ db.label }}</div>
              </div>
            </div>
          </div>

          <!-- Connection Form -->
          <div class="space-y-4">
            <div>
              <label class="block text-xs font-semibold text-gray-400 mb-2">Connection Name</label>
              <input
                v-model="form.name"
                type="text"
                placeholder="My Database"
                class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
              />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold text-gray-400 mb-2">Host</label>
                <input
                  v-model="form.host"
                  type="text"
                  placeholder="localhost"
                  class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-400 mb-2">Port</label>
                <input
                  v-model.number="form.port"
                  type="number"
                  class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold text-gray-400 mb-2">Username</label>
                <input
                  v-model="form.username"
                  type="text"
                  class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-400 mb-2">Password</label>
                <input
                  v-model="form.password"
                  type="password"
                  class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                />
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold text-gray-400 mb-2">Database (optional)</label>
              <input
                v-model="form.database"
                type="text"
                class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
              />
            </div>

            <div class="flex items-center gap-2">
              <input
                v-model="form.useSSL"
                type="checkbox"
                id="useSSL"
                class="w-4 h-4 rounded border-[#444] bg-[#3c3c3c] text-[#1677ff] focus:ring-[#1677ff]"
              />
              <label for="useSSL" class="text-xs text-gray-400">Use SSL/TLS</label>
            </div>
          </div>
        </div>

        <div class="px-6 py-4 border-t border-[#333] flex items-center justify-between bg-[#2d2d30]">
          <div class="flex items-center gap-3">
            <button
              @click="testConnection"
              :disabled="testStatus === 'loading'"
              :class="[
                'flex items-center gap-2 px-4 py-2 rounded text-xs font-semibold transition-all',
                testStatus === 'loading'
                  ? 'bg-gray-600 cursor-not-allowed'
                  : 'bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300'
              ]"
            >
              <Loader v-if="testStatus === 'loading'" :size="14" class="animate-spin" />
              <CheckCircle v-else-if="testStatus === 'success'" :size="14" class="text-green-500" />
              <XCircle v-else-if="testStatus === 'error'" :size="14" class="text-red-500" />
              <Database v-else :size="14" />
              {{ testStatus === 'loading' ? 'Testing...' : testStatus === 'success' ? 'Success' : testStatus === 'error' ? 'Failed' : 'Test Connection' }}
            </button>
            <span v-if="errorMessage" class="text-xs text-red-400">{{ errorMessage }}</span>
          </div>

          <div class="flex items-center gap-3">
            <button
              @click="emit('close')"
              class="px-4 py-2 rounded text-xs font-semibold bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="handleConnect"
              class="px-6 py-2 rounded text-xs font-semibold bg-[#1677ff] hover:bg-[#4096ff] text-white transition-colors"
            >
              Connect
            </button>
          </div>
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
