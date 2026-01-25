import type { QueryHistory } from '../types'

import {
  GetQueryHistory,
  ClearQueryHistory,
} from '../../wailsjs/go/main/App'

export const historyService = {
  async getQueryHistory(
    connectionId: string = '',
    searchTerm: string = '',
    limit: number = 50
  ): Promise<QueryHistory[]> {
    try {
      const result = await GetQueryHistory(connectionId, searchTerm, limit)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to get query history:', error)
      return []
    }
  },

  async clearQueryHistory(): Promise<void> {
    try {
      await ClearQueryHistory()
    } catch (error) {
      console.error('Failed to clear query history:', error)
      throw error
    }
  },
}
