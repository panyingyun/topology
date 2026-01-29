import {
  BackupNow,
  ListBackups,
  RestoreBackup,
  PickBackupFile,
} from '../../wailsjs/go/main/App'

export interface BackupRecord {
  connectionId: string
  path: string
  at: string
}

export interface BackupResult {
  success: boolean
  path?: string
  error?: string
}

export interface RestoreResult {
  success: boolean
  error?: string
}

export const backupService = {
  async backupNow(connectionId: string): Promise<BackupResult> {
    try {
      const json = await BackupNow(connectionId)
      return JSON.parse(json) as BackupResult
    } catch (e) {
      return {
        success: false,
        error: e instanceof Error ? e.message : 'Backup failed',
      }
    }
  },

  async listBackups(connectionId: string): Promise<BackupRecord[]> {
    try {
      const json = await ListBackups(connectionId)
      return JSON.parse(json) as BackupRecord[]
    } catch {
      return []
    }
  },

  async restoreBackup(connectionId: string, backupPath: string): Promise<RestoreResult> {
    try {
      const json = await RestoreBackup(connectionId, backupPath)
      return JSON.parse(json) as RestoreResult
    } catch (e) {
      return {
        success: false,
        error: e instanceof Error ? e.message : 'Restore failed',
      }
    }
  },

  async pickBackupFile(): Promise<string> {
    try {
      return await PickBackupFile()
    } catch {
      return ''
    }
  },
}
