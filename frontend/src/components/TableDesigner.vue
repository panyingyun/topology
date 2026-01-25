<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, Trash2, X, Database, Key, Link, FileCode } from 'lucide-vue-next'
import { schemaService } from '../services/schemaService'
import type { TableSchema, Column, Index, ForeignKey, DatabaseType } from '../types'

const { t } = useI18n()

const props = defineProps<{
  show: boolean
  connectionId?: string
  database?: string
  driver?: DatabaseType
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'create', sql: string): void
}>()

const tableName = ref('')
const columns = ref<Column[]>([
  {
    name: 'id',
    type: 'INT',
    nullable: false,
    isPrimaryKey: true,
    isUnique: false,
  },
])
const indexes = ref<Index[]>([])
const foreignKeys = ref<ForeignKey[]>([])

const showSQL = ref(false)
const generatedSQL = ref('')

const commonTypes = [
  'INT', 'BIGINT', 'VARCHAR(255)', 'TEXT', 'DATE', 'DATETIME', 'TIMESTAMP',
  'DECIMAL(10,2)', 'FLOAT', 'DOUBLE', 'BOOLEAN', 'BLOB'
]

const addColumn = () => {
  columns.value.push({
    name: '',
    type: 'VARCHAR(255)',
    nullable: true,
    isPrimaryKey: false,
    isUnique: false,
  })
}

const removeColumn = (index: number) => {
  columns.value.splice(index, 1)
}

const addIndex = () => {
  indexes.value.push({
    name: '',
    columns: [],
    isUnique: false,
    type: 'BTREE',
  })
}

const removeIndex = (index: number) => {
  indexes.value.splice(index, 1)
}

const addForeignKey = () => {
  foreignKeys.value.push({
    name: '',
    columns: [],
    referencedTable: '',
    referencedColumns: [],
    onDelete: 'RESTRICT',
    onUpdate: 'RESTRICT',
  })
}

const removeForeignKey = (index: number) => {
  foreignKeys.value.splice(index, 1)
}

const generateSQL = async () => {
  if (!tableName.value.trim()) {
    alert(t('designer.enterTableName'))
    return
  }
  if (columns.value.length === 0) {
    alert(t('designer.addAtLeastOneColumn'))
    return
  }

  const schema: TableSchema = {
    name: tableName.value,
    columns: columns.value,
    indexes: indexes.value,
    foreignKeys: foreignKeys.value,
  }

  try {
    const sql = await schemaService.generateCreateTableSQL(
      schema,
      props.driver || 'mysql'
    )
    generatedSQL.value = sql
    showSQL.value = true
  } catch (error) {
    console.error('Failed to generate SQL:', error)
    alert(t('designer.generateFailed') + ': ' + (error instanceof Error ? error.message : 'Unknown error'))
  }
}

const handleCreate = () => {
  if (generatedSQL.value) {
    emit('create', generatedSQL.value)
    handleClose()
  }
}

const handleClose = () => {
  tableName.value = ''
  columns.value = [{
    name: 'id',
    type: 'INT',
    nullable: false,
    isPrimaryKey: true,
    isUnique: false,
  }]
  indexes.value = []
  foreignKeys.value = []
  showSQL.value = false
  generatedSQL.value = ''
  emit('close')
}
</script>

<template>
  <Transition name="fade">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      @click.self="handleClose"
    >
      <div class="bg-[#252526] rounded-lg border border-[#333] w-full max-w-5xl max-h-[90vh] overflow-hidden flex flex-col">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-[#333] flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-200 flex items-center gap-2">
            <Database :size="20" class="text-[#1677ff]" />
            {{ t('designer.title') }}
          </h2>
          <button
            @click="handleClose"
            class="text-gray-400 hover:text-gray-200 transition-colors"
          >
            <X :size="20" />
          </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <!-- Table Name -->
          <div class="mb-6">
            <label class="block text-xs font-semibold text-gray-400 mb-2">{{ t('designer.tableName') }}</label>
            <input
              v-model="tableName"
              type="text"
              placeholder="users"
              class="w-full bg-[#3c3c3c] border border-[#444] rounded px-3 py-2 text-sm text-gray-200 focus:border-[#1677ff] focus:outline-none"
            />
          </div>

          <!-- Columns -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-sm font-semibold text-gray-300">{{ t('designer.columns') }}</h3>
              <button
                @click="addColumn"
                class="flex items-center gap-1 px-3 py-1 bg-[#1677ff] hover:bg-[#4096ff] text-white text-xs rounded transition-colors"
              >
                <Plus :size="12" />
                {{ t('designer.addColumn') }}
              </button>
            </div>
            <div class="space-y-2">
              <div
                v-for="(col, idx) in columns"
                :key="idx"
                class="grid grid-cols-12 gap-2 p-3 bg-[#1e1e1e] rounded border border-[#333]"
              >
                  <input
                    v-model="col.name"
                    type="text"
                    :placeholder="t('designer.columnName')"
                    class="col-span-3 bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                  />
                  <select
                    v-model="col.type"
                    class="col-span-3 bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                  >
                    <option v-for="type in commonTypes" :key="type" :value="type">
                      {{ type }}
                    </option>
                  </select>
                  <label class="col-span-2 flex items-center gap-1 text-xs text-gray-400">
                    <input
                      v-model="col.nullable"
                      type="checkbox"
                      class="w-4 h-4 rounded border-[#444] bg-[#3c3c3c] text-[#1677ff]"
                    />
                    {{ t('designer.nullable') }}
                  </label>
                  <label class="col-span-2 flex items-center gap-1 text-xs text-gray-400">
                    <input
                      v-model="col.isPrimaryKey"
                      type="checkbox"
                      class="w-4 h-4 rounded border-[#444] bg-[#3c3c3c] text-[#1677ff]"
                    />
                    <Key :size="12" />
                    {{ t('designer.primaryKey') }}
                  </label>
                  <label class="col-span-1 flex items-center gap-1 text-xs text-gray-400">
                    <input
                      v-model="col.isUnique"
                      type="checkbox"
                      :disabled="col.isPrimaryKey"
                      class="w-4 h-4 rounded border-[#444] bg-[#3c3c3c] text-[#1677ff]"
                    />
                    {{ t('designer.unique') }}
                  </label>
                <button
                  @click="removeColumn(idx)"
                  class="col-span-1 flex items-center justify-center text-red-400 hover:text-red-300"
                >
                  <Trash2 :size="14" />
                </button>
              </div>
            </div>
          </div>

          <!-- Indexes -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-sm font-semibold text-gray-300">{{ t('designer.indexes') }}</h3>
              <button
                @click="addIndex"
                class="flex items-center gap-1 px-3 py-1 bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 text-xs rounded transition-colors"
              >
                <Plus :size="12" />
                {{ t('designer.addIndex') }}
              </button>
            </div>
            <div v-if="indexes.length === 0" class="text-xs text-gray-500 text-center py-4">
              {{ t('designer.noIndexes') }}
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="(idx, idxIndex) in indexes"
                :key="idxIndex"
                class="p-3 bg-[#1e1e1e] rounded border border-[#333]"
              >
                <div class="flex items-center gap-2 mb-2">
                  <input
                    v-model="idx.name"
                    type="text"
                    :placeholder="t('designer.indexName')"
                    class="flex-1 bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                  />
                  <label class="flex items-center gap-1 text-xs text-gray-400">
                    <input
                      v-model="idx.isUnique"
                      type="checkbox"
                      class="w-4 h-4 rounded border-[#444] bg-[#3c3c3c] text-[#1677ff]"
                    />
                    {{ t('designer.unique') }}
                  </label>
                  <button
                    @click="removeIndex(idxIndex)"
                    class="text-red-400 hover:text-red-300"
                  >
                    <Trash2 :size="14" />
                  </button>
                </div>
                <input
                  v-model="idx.columns"
                  type="text"
                  :placeholder="t('designer.indexColumns')"
                  class="w-full bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                  @input="(e: any) => idx.columns = e.target.value.split(',').map((s: string) => s.trim())"
                />
              </div>
            </div>
          </div>

          <!-- Foreign Keys -->
          <div class="mb-6">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-sm font-semibold text-gray-300">{{ t('designer.foreignKeys') }}</h3>
              <button
                @click="addForeignKey"
                class="flex items-center gap-1 px-3 py-1 bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 text-xs rounded transition-colors"
              >
                <Plus :size="12" />
                {{ t('designer.addForeignKey') }}
              </button>
            </div>
            <div v-if="foreignKeys.length === 0" class="text-xs text-gray-500 text-center py-4">
              {{ t('designer.noForeignKeys') }}
            </div>
            <div v-else class="space-y-2">
              <div
                v-for="(fk, fkIndex) in foreignKeys"
                :key="fkIndex"
                class="p-3 bg-[#1e1e1e] rounded border border-[#333]"
              >
                <div class="grid grid-cols-2 gap-2 mb-2">
                  <input
                    v-model="fk.name"
                    type="text"
                    :placeholder="t('designer.fkName')"
                    class="bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                  />
                  <input
                    v-model="fk.referencedTable"
                    type="text"
                    :placeholder="t('designer.referencedTable')"
                    class="bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                  />
                </div>
                <div class="grid grid-cols-2 gap-2 mb-2">
                  <input
                    v-model="fk.columns"
                    type="text"
                    :placeholder="t('designer.fkColumns')"
                    class="bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                    @input="(e: any) => fk.columns = e.target.value.split(',').map((s: string) => s.trim())"
                  />
                  <input
                    v-model="fk.referencedColumns"
                    type="text"
                    :placeholder="t('designer.refColumns')"
                    class="bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                    @input="(e: any) => fk.referencedColumns = e.target.value.split(',').map((s: string) => s.trim())"
                  />
                </div>
                <div class="flex items-center gap-4">
                  <select
                    v-model="fk.onDelete"
                    class="flex-1 bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                  >
                    <option value="RESTRICT">RESTRICT</option>
                    <option value="CASCADE">CASCADE</option>
                    <option value="SET NULL">SET NULL</option>
                    <option value="NO ACTION">NO ACTION</option>
                  </select>
                  <select
                    v-model="fk.onUpdate"
                    class="flex-1 bg-[#3c3c3c] border border-[#444] rounded px-2 py-1.5 text-xs text-gray-200 focus:border-[#1677ff] focus:outline-none"
                  >
                    <option value="RESTRICT">RESTRICT</option>
                    <option value="CASCADE">CASCADE</option>
                    <option value="SET NULL">SET NULL</option>
                    <option value="NO ACTION">NO ACTION</option>
                  </select>
                  <button
                    @click="removeForeignKey(fkIndex)"
                    class="text-red-400 hover:text-red-300"
                  >
                    <Trash2 :size="14" />
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Generated SQL -->
          <Transition name="slide">
            <div v-if="showSQL" class="mb-6 p-4 bg-[#1e1e1e] rounded border border-[#333]">
              <div class="flex items-center justify-between mb-2">
                <h3 class="text-sm font-semibold text-gray-300 flex items-center gap-2">
                  <FileCode :size="16" />
                  {{ t('designer.generatedSQL') }}
                </h3>
                <button
                  @click="showSQL = false"
                  class="text-gray-400 hover:text-gray-200"
                >
                  <X :size="16" />
                </button>
              </div>
              <pre class="text-xs text-gray-300 font-mono overflow-x-auto p-3 bg-[#252526] rounded border border-[#333]">{{ generatedSQL }}</pre>
            </div>
          </Transition>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-[#333] flex items-center justify-between bg-[#2d2d30]">
          <button
            @click="handleClose"
            class="px-4 py-2 rounded text-xs font-semibold bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 transition-colors"
          >
            {{ t('common.cancel') }}
          </button>
          <div class="flex items-center gap-3">
            <button
              @click="generateSQL"
              class="px-4 py-2 rounded text-xs font-semibold bg-[#3c3c3c] hover:bg-[#4c4c4c] text-gray-300 transition-colors"
            >
              {{ t('designer.generateSQL') }}
            </button>
            <button
              v-if="generatedSQL"
              @click="handleCreate"
              class="px-6 py-2 rounded text-xs font-semibold bg-[#1677ff] hover:bg-[#4096ff] text-white transition-colors"
            >
              {{ t('designer.createTable') }}
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

.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
