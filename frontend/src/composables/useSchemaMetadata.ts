import { ref, onMounted } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { schemaService, type SchemaMetadata } from '../services/schemaService'

const SCHEMA_READY_EVENT = 'schema-metadata-ready'

/** Per-connection schema metadata cache for SQL completion. */
const cache = ref<Record<string, SchemaMetadata>>({})
let unsubscribe: (() => void) | null = null

export function useSchemaMetadata() {
  const ensureListener = () => {
    if (unsubscribe) return
    unsubscribe = EventsOn(SCHEMA_READY_EVENT, async (connectionId: string) => {
      if (!connectionId) return
      try {
        const meta = await schemaService.getSchemaMetadata(connectionId)
        cache.value = { ...cache.value, [connectionId]: meta }
      } catch (e) {
        console.error('Failed to fetch schema metadata after ready event:', e)
      }
    })
  }

  /** Trigger backend to load metadata (async). When done, cache is updated via event. */
  const load = (connectionId: string) => {
    if (!connectionId) return
    ensureListener()
    schemaService.loadSchemaMetadata(connectionId)
    // Optionally fetch current cache in case it was already loaded
    schemaService.getSchemaMetadata(connectionId).then((meta) => {
      if (meta.databases && meta.databases.length > 0) {
        cache.value = { ...cache.value, [connectionId]: meta }
      }
    }).catch(() => {})
  }

  /** Get cached metadata for connection (may be empty if not loaded yet). */
  const get = (connectionId: string): SchemaMetadata | undefined => {
    return cache.value[connectionId]
  }

  /** Flatten all table names across databases (e.g. ["users", "orders"]). */
  const getAllTableNames = (connectionId: string): string[] => {
    const meta = cache.value[connectionId]
    if (!meta?.databases?.length) return []
    const set = new Set<string>()
    for (const db of meta.databases) {
      for (const t of db.tables || []) {
        if (t.name) set.add(t.name)
      }
    }
    return Array.from(set)
  }

  /** Flatten all columns with optional table prefix (e.g. "users" -> ["id","name"] or "users.id"). */
  const getColumnsForTable = (connectionId: string, tableName: string): Array<{ label: string; detail?: string }> => {
    const meta = cache.value[connectionId]
    if (!meta?.databases?.length) return []
    const tableNameLower = tableName.toLowerCase()
    const items: Array<{ label: string; detail?: string }> = []
    for (const db of meta.databases) {
      for (const t of db.tables || []) {
        if (t.name.toLowerCase() !== tableNameLower) continue
        for (const c of t.columns || []) {
          items.push({ label: c.name, detail: c.type })
        }
        return items
      }
    }
    return items
  }

  /** All columns from all tables, with detail "table (type)". */
  const getAllColumns = (connectionId: string): Array<{ label: string; detail?: string }> => {
    const meta = cache.value[connectionId]
    if (!meta?.databases?.length) return []
    const items: Array<{ label: string; detail?: string }> = []
    for (const db of meta.databases) {
      for (const t of db.tables || []) {
        for (const c of t.columns || []) {
          items.push({ label: c.name, detail: `${t.name} (${c.type || ''})` })
        }
      }
    }
    return items
  }

  onMounted(() => ensureListener())

  return {
    schemaCache: cache,
    load,
    get,
    getAllTableNames,
    getColumnsForTable,
    getAllColumns,
  }
}
