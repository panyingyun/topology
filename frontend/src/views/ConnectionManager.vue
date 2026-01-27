<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { Database, CheckCircle, XCircle, Loader, Lock, ChevronDown, ChevronRight, Eye, EyeOff, Plus, Trash2 } from 'lucide-vue-next'
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
  (e: 'deleted'): void
}>()

const savedConnections = ref<Connection[]>([])
const selectedId = ref<string | null>(null)
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
const passwordVisible = ref(false)
const sshPasswordVisible = ref(false)

const isEditMode = computed(() => !!selectedId.value)

function loadFormFrom(conn: Connection | null) {
  if (!conn) {
    activeDbType.value = 'mysql'
    form.name = ''
    form.host = '127.0.0.1'
    form.port = 3306
    form.username = 'root'
    form.password = ''
    form.database = ''
    form.useSSL = false
    form.sshTunnel = { enabled: false, host: '', port: 22, username: '', password: '', privateKey: '' }
    return
  }
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
}

async function loadConnections() {
  try {
    savedConnections.value = await connectionService.getConnections()
  } catch (e) {
    console.error('Failed to load connections:', e)
  }
}

watch(() => props.show, async (visible) => {
  if (visible) {
    await loadConnections()
    if (props.mode === 'edit' && props.editConnection) {
      selectedId.value = props.editConnection.id
      loadFormFrom(props.editConnection)
    } else {
      selectedId.value = null
      loadFormFrom(null)
    }
    testStatus.value = 'idle'
    errorMessage.value = ''
    passwordVisible.value = false
    sshPasswordVisible.value = false
  }
})

function selectNew() {
  selectedId.value = null
  loadFormFrom(null)
  errorMessage.value = ''
}

function selectConnection(conn: Connection) {
  selectedId.value = conn.id
  loadFormFrom(conn)
  errorMessage.value = ''
}

async function deleteConnection(e: MouseEvent, id: string) {
  e.stopPropagation()
  try {
    await connectionService.deleteConnection(id)
    await loadConnections()
    if (selectedId.value === id) {
      selectedId.value = null
      loadFormFrom(null)
    }
    emit('deleted')
  } catch (err) {
    console.error('Delete connection failed:', err)
    errorMessage.value = err instanceof Error ? err.message : 'Delete failed'
  }
}

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
  const id = selectedId.value
  if (!id) return
  const existing = savedConnections.value.find((c) => c.id === id)
  if (!existing) return
  try {
    const payload = buildConnectionPayload()
    const updated: Connection = {
      ...existing,
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
      <div class="theme-bg-panel rounded-lg border theme-border w-full max-w-5xl max-h-[90vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b theme-border flex items-center justify-between">
          <h2 class="text-lg font-semibold theme-text">
            {{ isEditMode ? t('connection.editConnection') : t('connection.newConnection') }}
          </h2>
          <button
            @click="emit('close')"
            class="theme-text-muted-hover transition-colors"
          >
            <XCircle :size="20" />
          </button>
        </div>

        <div class="flex-1 flex overflow-hidden min-h-0">
          <!-- Left: saved connections list -->
          <div class="w-56 flex-shrink-0 border-r theme-border flex flex-col theme-bg-footer/50">
            <div class="p-2 border-b theme-border">
              <button
                type="button"
                @click="selectNew"
                :class="[
                  'w-full flex items-center justify-center gap-2 py-2 rounded text-xs font-semibold transition-colors',
                  !selectedId ? 'bg-[#1677ff] text-white' : 'theme-bg-input theme-bg-input-hover theme-text'
                ]"
              >
                <Plus :size="14" />
                {{ t('connection.newConnection') }}
              </button>
            </div>
            <div class="flex-1 overflow-y-auto custom-scrollbar p-2 space-y-1">
              <div
                v-for="c in savedConnections"
                :key="c.id"
                @click="selectConnection(c)"
                :class="[
                  'flex items-center gap-2 px-3 py-2 rounded text-left text-xs transition-colors cursor-pointer',
                  selectedId === c.id ? 'theme-bg-input ring-1 ring-[#1677ff]/50' : 'theme-bg-hover theme-text'
                ]"
              >
                <span class="flex-1 truncate theme-text">{{ c.name }}</span>
                <button
                  type="button"
                  @click="deleteConnection($event, c.id)"
                  class="theme-text-muted-hover p-0.5 rounded"
                  :title="t('connection.delete')"
                >
                  <Trash2 :size="12" />
                </button>
              </div>
              <p v-if="savedConnections.length === 0" class="px-3 py-4 text-xs theme-text-muted">
                {{ t('connection.noSavedConnections') }}
              </p>
            </div>
          </div>

          <!-- Right: form -->
          <div class="flex-1 overflow-y-auto custom-scrollbar p-6 min-w-0">
          <!-- Database Type Selection -->
          <div class="mb-6">
            <label class="block text-xs font-semibold theme-text-muted mb-3 uppercase tracking-wider">
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
                    : 'theme-border theme-bg-footer hover:border-[var(--border-strong)]'
                ]"
              >
                <div class="text-3xl mb-2">{{ db.icon }}</div>
                <div class="text-sm font-semibold theme-text">{{ db.label }}</div>
              </div>
            </div>
          </div>

          <!-- Connection Form -->
          <div class="space-y-4">
            <div>
              <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.connectionName') }}</label>
              <input
                v-model="form.name"
                type="text"
                placeholder="My Database"
                class="w-full theme-input rounded px-3 py-2 text-sm"
              />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.host') }}</label>
                <input
                  v-model="form.host"
                  type="text"
                  placeholder="localhost"
                  class="w-full theme-input rounded px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.port') }}</label>
                <input
                  v-model.number="form.port"
                  type="number"
                  class="w-full theme-input rounded px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.username') }}</label>
                <input
                  v-model="form.username"
                  type="text"
                  class="w-full theme-input rounded px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.password') }}</label>
                <div class="relative">
                  <input
                    v-model="form.password"
                    :type="passwordVisible ? 'text' : 'password'"
                    class="w-full theme-input rounded px-3 py-2 pr-9 text-sm"
                  />
                  <button
                    type="button"
                    class="absolute right-2 top-1/2 -translate-y-1/2 theme-text-muted-hover transition-colors"
                    :title="passwordVisible ? t('connection.hidePassword') : t('connection.showPassword')"
                    @click="passwordVisible = !passwordVisible"
                  >
                    <EyeOff v-if="passwordVisible" :size="14" />
                    <Eye v-else :size="14" />
                  </button>
                </div>
              </div>
            </div>

            <div>
              <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.database') }} ({{ t('common.optional') }})</label>
              <input
                v-model="form.database"
                type="text"
                class="w-full theme-input rounded px-3 py-2 text-sm"
              />
            </div>

            <div class="flex items-center gap-2">
              <input
                v-model="form.useSSL"
                type="checkbox"
                id="useSSL"
                class="w-4 h-4 rounded theme-border-strong theme-bg-input text-[#1677ff] focus:ring-[#1677ff]"
              />
              <label for="useSSL" class="text-xs theme-text-muted">{{ t('connection.useSSL') }}</label>
            </div>

            <!-- SSH Tunnel (MySQL only in backend) -->
            <div v-if="activeDbType === 'mysql'" class="mt-6 pt-4 border-t theme-border">
              <button
                type="button"
                class="flex items-center gap-2 w-full text-left mb-3"
                @click="showSshSection = !showSshSection"
              >
                <Lock :size="16" class="text-[#1677ff] flex-shrink-0" />
                <span class="text-sm font-semibold theme-text">{{ t('connection.sshTunnel.title') }}</span>
                <ChevronDown v-if="showSshSection" :size="16" class="theme-text-muted ml-auto" />
                <ChevronRight v-else :size="16" class="theme-text-muted ml-auto" />
              </button>
              <div v-show="showSshSection" class="space-y-4 pl-6 border-l-2 theme-border ml-1">
                <div class="flex items-center gap-2">
                  <input
                    v-model="form.sshTunnel.enabled"
                    type="checkbox"
                    id="sshEnabled"
                    class="w-4 h-4 rounded theme-border-strong theme-bg-input text-[#1677ff] focus:ring-[#1677ff]"
                  />
                  <label for="sshEnabled" class="text-xs theme-text-muted">{{ t('connection.sshTunnel.enable') }}</label>
                </div>
                <template v-if="form.sshTunnel.enabled">
                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.sshTunnel.host') }}</label>
                      <input
                        v-model="form.sshTunnel.host"
                        type="text"
                        placeholder="jump.example.com"
                        class="w-full theme-input rounded px-3 py-2 text-sm"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.sshTunnel.port') }}</label>
                      <input
                        v-model.number="form.sshTunnel.port"
                        type="number"
                        min="1"
                        max="65535"
                        class="w-full theme-input rounded px-3 py-2 text-sm"
                      />
                    </div>
                  </div>
                  <div>
                    <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.sshTunnel.username') }}</label>
                    <input
                      v-model="form.sshTunnel.username"
                      type="text"
                      placeholder="ssh_user"
                      class="w-full theme-input rounded px-3 py-2 text-sm"
                    />
                  </div>
                  <div>
                    <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.sshTunnel.password') }} ({{ t('common.optional') }})</label>
                    <div class="relative">
                      <input
                        v-model="form.sshTunnel.password"
                        :type="sshPasswordVisible ? 'text' : 'password'"
                        :placeholder="t('connection.sshTunnel.passwordPlaceholder')"
                        class="w-full theme-input rounded px-3 py-2 pr-9 text-sm"
                      />
                      <button
                        type="button"
                        class="absolute right-2 top-1/2 -translate-y-1/2 theme-text-muted-hover transition-colors"
                        :title="sshPasswordVisible ? t('connection.hidePassword') : t('connection.showPassword')"
                        @click="sshPasswordVisible = !sshPasswordVisible"
                      >
                        <EyeOff v-if="sshPasswordVisible" :size="14" />
                        <Eye v-else :size="14" />
                      </button>
                    </div>
                  </div>
                  <div>
                    <label class="block text-xs font-semibold theme-text-muted mb-2">{{ t('connection.sshTunnel.privateKey') }} ({{ t('common.optional') }})</label>
                    <textarea
                      v-model="form.sshTunnel.privateKey"
                      :placeholder="t('connection.sshTunnel.privateKeyPlaceholder')"
                      rows="4"
                      class="w-full theme-input rounded px-3 py-2 text-sm font-mono resize-y"
                    />
                    <p class="text-xs theme-text-muted opacity-80 mt-1">{{ t('connection.sshTunnel.mysqlOnly') }}</p>
                  </div>
                </template>
              </div>
            </div>
          </div>
        </div>
        </div>

        <div class="px-6 py-4 border-t theme-border flex items-center justify-between theme-bg-footer">
          <div class="flex items-center gap-3">
            <button
              @click="testConnection"
              :disabled="testStatus === 'loading'"
              :class="[
                'flex items-center gap-2 px-4 py-2 rounded text-xs font-semibold transition-all',
                testStatus === 'loading'
                  ? 'bg-gray-600 cursor-not-allowed theme-text-muted'
                  : 'theme-bg-input theme-bg-input-hover theme-text'
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
              class="px-4 py-2 rounded text-xs font-semibold theme-bg-input theme-bg-input-hover theme-text transition-colors"
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
