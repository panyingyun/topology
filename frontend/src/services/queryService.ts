import type { QueryResult, ExecutionPlanResult, IndexSuggestion } from '../types'

import { ExecuteQuery, FormatSQL, GetExecutionPlan, GetQueryCacheStats, GetIndexSuggestions } from '../../wailsjs/go/main/App'

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
  /** Execute SQL. sessionId isolates DB session per tab (e.g. tab id); pass '' when no tab context. */
  async executeQuery(connectionId: string, sessionId: string, sql: string): Promise<QueryResult> {
    try {
      const result = await withTimeout(
        ExecuteQuery(connectionId, sessionId, sql),
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

  /** Get structured execution plan (EXPLAIN) for visualization. MySQL and PostgreSQL supported. */
  async getExecutionPlan(connectionId: string, sessionId: string, sql: string): Promise<ExecutionPlanResult> {
    try {
      const result = await GetExecutionPlan(connectionId, sessionId, sql)
      return JSON.parse(result) as ExecutionPlanResult
    } catch (error) {
      console.error('Failed to get execution plan:', error)
      return {
        nodes: [],
        summary: {},
        error: error instanceof Error ? error.message : 'Unknown error',
      }
    }
  },

  async getQueryCacheStats(): Promise<{ hits: number; misses: number }> {
    try {
      const raw = await GetQueryCacheStats()
      const o = JSON.parse(raw) as { hits: number; misses: number }
      return { hits: o.hits ?? 0, misses: o.misses ?? 0 }
    } catch {
      return { hits: 0, misses: 0 }
    }
  },

  async getIndexSuggestions(
    connectionId: string,
    sessionId: string,
    sql: string
  ): Promise<{ suggestions: IndexSuggestion[]; error?: string }> {
    try {
      const raw = await GetIndexSuggestions(connectionId, sessionId, sql)
      const o = JSON.parse(raw) as { suggestions?: IndexSuggestion[]; error?: string }
      return { suggestions: o.suggestions ?? [], error: o.error }
    } catch (e) {
      return {
        suggestions: [],
        error: e instanceof Error ? e.message : 'Failed to get index suggestions',
      }
    }
  },
}
