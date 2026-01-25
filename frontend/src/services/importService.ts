import type { ImportPreview, ImportResult, ImportFormat } from '../types'

import {
  ImportDataPreview,
  ImportData,
} from '../../wailsjs/go/main/App'

export const importService = {
  async previewImport(
    filePath: string,
    format: ImportFormat
  ): Promise<ImportPreview> {
    try {
      const result = await ImportDataPreview(filePath, format)
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to preview import:', error)
      return {
        columns: [],
        rows: [],
        error: error instanceof Error ? error.message : 'Unknown error',
      }
    }
  },

  async importData(
    connectionId: string,
    database: string,
    tableName: string,
    filePath: string,
    format: ImportFormat,
    columnMapping: Record<string, string>,
    sessionId: string = ''
  ): Promise<ImportResult> {
    try {
      const mappingJSON = JSON.stringify(columnMapping)
      const result = await ImportData(
        connectionId,
        database,
        tableName,
        filePath,
        format,
        mappingJSON,
        sessionId
      )
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to import data:', error)
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      }
    }
  },
}
