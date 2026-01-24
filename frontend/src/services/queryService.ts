import type { QueryResult } from '../types'

import { ExecuteQuery, FormatSQL } from '../../wailsjs/go/main/App'

export const queryService = {
  async executeQuery(connectionId: string, sql: string): Promise<QueryResult> {
    try {
      const result = await ExecuteQuery(connectionId, sql)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to execute query:', error)
      return {
        columns: [],
        rows: [],
        rowCount: 0,
        error: error instanceof Error ? error.message : 'Unknown error',
      }
    }
  },

  async formatSQL(sql: string): Promise<string> {
    try {
      return await FormatSQL(sql)
    } catch (error) {
      console.error('Failed to format SQL:', error)
      return sql
    }
  },
}
