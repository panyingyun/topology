import type { TableSchema } from '../types'

import {
  GenerateCreateTableSQL,
  AnalyzeSQL,
} from '../../wailsjs/go/main/App'

export const schemaService = {
  async generateCreateTableSQL(
    schema: TableSchema,
    driver: string
  ): Promise<string> {
    try {
      const schemaJSON = JSON.stringify(schema)
      return await GenerateCreateTableSQL(schemaJSON, driver)
    } catch (error) {
      console.error('Failed to generate CREATE TABLE SQL:', error)
      throw error
    }
  },

  async analyzeSQL(sql: string, driver: string): Promise<any> {
    try {
      const result = await AnalyzeSQL(sql, driver)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to analyze SQL:', error)
      throw error
    }
  },
}
