import type { QueryResult } from '../types'

import { ExecuteQuery, FormatSQL } from '../../wailsjs/go/main/App'

const QUERY_TIMEOUT_MS = 120000 // 2 minutes

function withTimeout<T>(promise: Promise<T>, ms: number, message: string): Promise<T> {
  return Promise.race([
    promise,
    new Promise<T>((_, reject) =>
      setTimeout(() => reject(new Error(message)), ms)
    ),
  ])
}

export const queryService = {
  async executeQuery(connectionId: string, sql: string): Promise<QueryResult> {
    try {
      const result = await withTimeout(
        ExecuteQuery(connectionId, sql),
        QUERY_TIMEOUT_MS,
        'Query timeout (exceeded ' + QUERY_TIMEOUT_MS / 1000 + 's)'
      )
      return JSON.parse(result) as QueryResult
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
