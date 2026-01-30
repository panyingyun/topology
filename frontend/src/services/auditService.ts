import { QueryAuditLog, ExportAuditLog } from '../../wailsjs/go/main/App'

export interface AuditEntry {
  at: string
  op: string
  detail: string
  connectionId?: string
  database?: string
  table?: string
}

export const auditService = {
  async query(limit: number = 100, since: string = '', opFilter: string = ''): Promise<AuditEntry[]> {
    try {
      const json = await QueryAuditLog(limit, since, opFilter)
      const arr = JSON.parse(json) as AuditEntry[]
      return Array.isArray(arr) ? arr : []
    } catch {
      return []
    }
  },

  async exportFormat(format: 'json' | 'csv'): Promise<string> {
    try {
      return await ExportAuditLog(format)
    } catch {
      return ''
    }
  },
}
