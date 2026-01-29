import type { Table, TableData, TableSchema, UpdateRecord } from '../types'

import {
  GetDatabases,
  GetTables,
  GetTableData,
  UpdateTableData,
  GetTableSchema,
  ExportData,
  DeleteTableRows,
  InsertTableRows,
} from '../../wailsjs/go/main/App'

/** Optional sessionId for per-tab DB session isolation; pass '' for shared connection. */
const defaultSession = ''

export const dataService = {
  async getDatabases(connectionId: string, sessionId: string = defaultSession): Promise<string[]> {
    try {
      const result = await GetDatabases(connectionId, sessionId)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to get databases:', error)
      return []
    }
  },

  async getTables(connectionId: string, database: string, sessionId: string = defaultSession): Promise<Table[]> {
    try {
      const result = await GetTables(connectionId, database, sessionId)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to get tables:', error)
      return []
    }
  },

  async getTableData(
    connectionId: string,
    database: string,
    tableName: string,
    limit: number = 100,
    offset: number = 0,
    sessionId: string = defaultSession
  ): Promise<TableData> {
    try {
      const result = await GetTableData(connectionId, database, tableName, limit, offset, sessionId)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to get table data:', error)
      return {
        columns: [],
        rows: [],
        totalRows: 0,
        page: 1,
        pageSize: limit,
      }
    }
  },

  async updateTableData(
    connectionId: string,
    database: string,
    tableName: string,
    updates: UpdateRecord[],
    sessionId: string = defaultSession
  ): Promise<void> {
    try {
      const updatesJSON = JSON.stringify(updates)
      await UpdateTableData(connectionId, database, tableName, updatesJSON, sessionId)
    } catch (error) {
      console.error('Failed to update table data:', error)
      throw error
    }
  },

  async getTableSchema(
    connectionId: string,
    database: string,
    tableName: string,
    sessionId: string = defaultSession
  ): Promise<TableSchema> {
    try {
      const result = await GetTableSchema(connectionId, database, tableName, sessionId)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to get table schema:', error)
      return {
        name: tableName,
        columns: [],
        indexes: [],
        foreignKeys: [],
      }
    }
  },

  async exportData(
    connectionId: string,
    database: string,
    tableName: string,
    format: string,
    sessionId: string = defaultSession
  ): Promise<{ success: boolean; filename?: string; error?: string }> {
    try {
      const result = await ExportData(connectionId, database, tableName, format, sessionId)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to export data:', error)
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      }
    }
  },

  async deleteTableRows(
    connectionId: string,
    database: string,
    tableName: string,
    rows: Record<string, unknown>[],
    sessionId: string = defaultSession
  ): Promise<void> {
    const rowsJSON = JSON.stringify(rows)
    await DeleteTableRows(connectionId, database, tableName, rowsJSON, sessionId)
  },

  async insertTableRows(
    connectionId: string,
    database: string,
    tableName: string,
    rows: Record<string, unknown>[],
    sessionId: string = defaultSession
  ): Promise<void> {
    const rowsJSON = JSON.stringify(rows)
    await InsertTableRows(connectionId, database, tableName, rowsJSON, sessionId)
  },
}
