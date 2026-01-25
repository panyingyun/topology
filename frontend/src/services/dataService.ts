import type { Table, TableData, TableSchema, UpdateRecord } from '../types'

import {
  GetDatabases,
  GetTables,
  GetTableData,
  UpdateTableData,
  GetTableSchema,
  ExportData,
} from '../../wailsjs/go/main/App'

export const dataService = {
  async getDatabases(connectionId: string): Promise<string[]> {
    try {
      const result = await GetDatabases(connectionId)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to get databases:', error)
      return []
    }
  },

  async getTables(connectionId: string, database: string): Promise<Table[]> {
    try {
      const result = await GetTables(connectionId, database)
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
    offset: number = 0
  ): Promise<TableData> {
    try {
      const result = await GetTableData(connectionId, database, tableName, limit, offset)
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
    updates: UpdateRecord[]
  ): Promise<void> {
    try {
      const updatesJSON = JSON.stringify(updates)
      await UpdateTableData(connectionId, database, tableName, updatesJSON)
    } catch (error) {
      console.error('Failed to update table data:', error)
      throw error
    }
  },

  async getTableSchema(connectionId: string, database: string, tableName: string): Promise<TableSchema> {
    try {
      const result = await GetTableSchema(connectionId, database, tableName)
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
    format: string
  ): Promise<{ success: boolean; filename?: string; error?: string }> {
    try {
      const result = await ExportData(connectionId, database, tableName, format)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to export data:', error)
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      }
    }
  },
}
