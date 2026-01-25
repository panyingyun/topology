import {
  LoadSchemaMetadata,
  GetSchemaMetadata,
  AnalyzeSQL,
  GenerateCreateTableSQL,
} from '../../wailsjs/go/main/App'
import type { SQLAnalysis } from '../types'

export interface SchemaColumnMeta {
  name: string
  type?: string
}

export interface SchemaTableMeta {
  name: string
  columns: SchemaColumnMeta[]
}

export interface SchemaDBMeta {
  name: string
  tables: SchemaTableMeta[]
}

export interface SchemaMetadata {
  connectionId: string
  databases: SchemaDBMeta[]
}

export const schemaService = {
  /** Trigger async metadata fetch for the connection. Listen for 'schema-metadata-ready' then call getSchemaMetadata. */
  loadSchemaMetadata(connectionId: string): void {
    LoadSchemaMetadata(connectionId)
  },

  async getSchemaMetadata(connectionId: string): Promise<SchemaMetadata> {
    const json = await GetSchemaMetadata(connectionId)
    try {
      return JSON.parse(json) as SchemaMetadata
    } catch {
      return { connectionId, databases: [] }
    }
  },

  async analyzeSQL(sql: string, driver: string): Promise<SQLAnalysis> {
    const json = await AnalyzeSQL(sql, driver)
    return JSON.parse(json) as SQLAnalysis
  },

  async generateCreateTableSQL(schema: object, driver: string): Promise<string> {
    return GenerateCreateTableSQL(JSON.stringify(schema), driver)
  },
}
