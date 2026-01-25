<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { Database, CheckCircle, XCircle, Loader, Lock, ChevronDown, ChevronRight } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { connectionService } from '../services/connectionService'
import type { Connection, DatabaseType } from '../types'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  mode?: 'create' | 'edit'
  editConnection?: Connection | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'connect', connection: Connection): void
  (e: 'update', connection: Connection): void
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
  sshTunnel: {
    enabled: false,
    host: '',
    port: 22,
    username: '',
    password: '',
    privateKey: '',
  } as {
    enabled: boolean
    host: string
    port: number
    username: string
    password: string
    privateKey: string
  },
})

const showSshSection = ref(true)

const isEditMode = computed(() => props.mode === 'edit' && props.editConnection)

watch([() => props.show, () => props.editConnection], () => {
  if (props.show) {
    if (isEditMode.value && props.editConnection) {
      const conn = props.editConnection
      activeDbType.value = (conn.type as DatabaseType) || 'mysql'
      form.name = conn.name || ''
      form.host = conn.host || '127.0.0.1'
      form.port = conn.port || (activeDbType.value === 'mysql' ? 3306 : 5432)
      form.username = conn.username || 'root'
      form.password = conn.password || ''
      form.database = conn.database || ''
      form.useSSL = conn.useSSL || false
      const st = conn.sshTunnel
      form.sshTunnel = {
        enabled: st?.enabled ?? false,
        host: st?.host ?? '',
        port: st?.port ?? 22,
        username: st?.username ?? '',
        password: st?.password ?? '',
        privateKey: st?.privateKey ?? '',
      }
    } else {
      activeDbType.value = 'mysql'
      form.name = ''
      form.host = '127.0.0.1'
      form.port = 3306
      form.username = 'root'
      form.password = ''
      form.database = ''
      form.useSSL = false
      form.sshTunnel = { enabled: false, host: '', port: 22, username: '', password: '', privateKey: '' }
    }
    testStatus.value = 'idle'
    errorMessage.value = ''
  }
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

const buildConnectionPayload = () => ({
  name: form.name,
  type: activeDbType.value,
  host: form.host,
  port: form.port,
  username: form.username,
  password: form.password,
  database: form.database || undefined,
  useSSL: form.useSSL,
  sshTunnel:
    form.sshTunnel.enabled &&
    (activeDbType.value === 'mysql' || activeDbType.value === 'postgresql')
      ? {
          enabled: true,
          host: form.sshTunnel.host || undefined,
          port: form.sshTunnel.port || 22,
          username: form.sshTunnel.username || undefined,
          password: form.sshTunnel.password || undefined,
          privateKey: form.sshTunnel.privateKey || undefined,
        }
      : undefined,
})

const testConnection = async () => {
  testStatus.value = 'loading'
  errorMessage.value = ''
  try {
    const result = await connectionService.testConnection(buildConnectionPayload() as any)
    
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
    const connection = await connectionService.createConnection(buildConnectionPayload() as any)
    emit('connect', connection)
    emit('close')
  } catch (error) {
    console.error('Failed to create connection:', error)
  }
}

const handleUpdate = async () => {
  if (!isEditMode.value || !props.editConnection) return
  try {
    const payload = buildConnectionPayload()
    const updated: Connection = {
      ...props.editConnection,
      name: payload.name,
      type: payload.type,
      host: payload.host,
      port: payload.port,
      username: payload.username,
      password: payload.password,
      database: payload.database,
      useSSL: payload.useSSL,
      sshTunnel: payload.sshTunnel,
    }
    await connectionService.updateConnection(updated)
    emit('update', updated)
    emit('close')
  } catch (error) {
    console.error('Failed to update connection:', error)
    errorMessage.value = error instanceof Error ? error.message : 'Failed to update connection'
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
          <h2 class="text-lg font-semibold text-gray-200">
            {{ isEditMode ? t('connection.editConnection') : t('connection.newConnection') }}
          </h2>
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
              {{ t('connection.databaseType') }}
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
              <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.connectionName') }}</label>
              <input
                v-model="form.name"
                type="text"
                placeholder="My Database"
                class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
              />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.host') }}</label>
                <input
                  v-model="form.host"
                  type="text"
                  placeholder="localhost"
                  class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.port') }}</label>
                <input
                  v-model.number="form.port"
                  type="number"
                  class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.username') }}</label>
                <input
                  v-model="form.username"
                  type="text"
                  class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.password') }}</label>
                <input
                  v-model="form.password"
                  type="password"
                  class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                />
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.database') }} ({{ t('common.optional') }})</label>
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
              <label for="useSSL" class="text-xs text-gray-400">{{ t('connection.useSSL') }}</label>
            </div>

            <!-- SSH Tunnel (MySQL only in backend) -->
            <div v-if="activeDbType === 'mysql'" class="mt-6 pt-4 border-t border-[#333]">
              <button
                type="button"
                class="flex items-center gap-2 w-full text-left mb-3"
                @click="showSshSection = !showSshSection"
              >
                <Lock :size="16" class="text-[#1677ff] flex-shrink-0" />
                <span class="text-sm font-semibold text-gray-200">{{ t('connection.sshTunnel.title') }}</span>
                <ChevronDown v-if="showSshSection" :size="16" class="text-gray-400 ml-auto" />
                <ChevronRight v-else :size="16" class="text-gray-400 ml-auto" />
              </button>
              <div v-show="showSshSection" class="space-y-4 pl-6 border-l-2 border-[#333] ml-1">
                <div class="flex items-center gap-2">
                  <input
                    v-model="form.sshTunnel.enabled"
                    type="checkbox"
                    id="sshEnabled"
                    class="w-4 h-4 rounded border-[#444] bg-[#3c3c3c] text-[#1677ff] focus:ring-[#1677ff]"
                  />
                  <label for="sshEnabled" class="text-xs text-gray-400">{{ t('connection.sshTunnel.enable') }}</label>
                </div>
                <template v-if="form.sshTunnel.enabled">
                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.sshTunnel.host') }}</label>
                      <input
                        v-model="form.sshTunnel.host"
                        type="text"
                        placeholder="jump.example.com"
                        class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.sshTunnel.port') }}</label>
                      <input
                        v-model.number="form.sshTunnel.port"
                        type="number"
                        min="1"
                        max="65535"
                        class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                      />
                    </div>
                  </div>
                  <div>
                    <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.sshTunnel.username') }}</label>
                    <input
                      v-model="form.sshTunnel.username"
                      type="text"
                      placeholder="ssh_user"
                      class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                    />
                  </div>
                  <div>
                    <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.sshTunnel.password') }} ({{ t('common.optional') }})</label>
                    <input
                      v-model="form.sshTunnel.password"
                      type="password"
                      :placeholder="t('connection.sshTunnel.passwordPlaceholder')"
                      class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
                    />
                  </div>
                  <div>
                    <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('connection.sshTunnel.privateKey') }} ({{ t('common.optional') }})</label>
                    <textarea
                      v-model="form.sshTunnel.privateKey"
                      :placeholder="t('connection.sshTunnel.privateKeyPlaceholder')"
                      rows="4"
                      class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none font-mono resize-y"
                    />
                    <p class="text-xs text-gray-500 mt-1">{{ t('connection.sshTunnel.mysqlOnly') }}</p>
                  </div>
                </template>
              </div>
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
              {{ testStatus === 'loading' ? t('common.loading') : testStatus === 'success' ? t('common.success') : testStatus === 'error' ? t('common.error') : t('connection.testConnection') }}
            </button>
            <span v-if="errorMessage" class="text-xs text-red-400">{{ errorMessage }}</span>
          </div>

          <div class="flex items-center gap-3">
            <button
              @click="emit('close')"
              class="px-4 py-2 rounded text-xs font-semibold bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 transition-colors"
            >
              {{ t('common.cancel') }}
            </button>
            <button
              v-if="!isEditMode"
              @click="handleConnect"
              class="px-6 py-2 rounded text-xs font-semibold bg-[#1677ff] hover:bg-[#4096ff] text-white transition-colors"
            >
              {{ t('connection.connect') }}
            </button>
            <button
              v-else
              @click="handleUpdate"
              class="px-6 py-2 rounded text-xs font-semibold bg-[#1677ff] hover:bg-[#4096ff] text-white transition-colors"
            >
              {{ t('connection.update') }}
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
