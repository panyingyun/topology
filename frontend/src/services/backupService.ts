import {
  BackupNow,
  ListBackups,
  RestoreBackup,
  PickBackupFile,
  GetBackupSchedules,
  SetBackupSchedules,
  DeleteBackup,
  VerifyBackup,
} from '../../wailsjs/go/main/App'

export interface BackupRecord {
  connectionId: string
  path: string
  at: string
}

export interface BackupSchedule {
  connectionId: string
  enabled: boolean
  schedule: 'daily' | 'weekly'
  time: string
  day: number
  outputDir?: string
  lastRun?: string
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

export interface VerifyResult {
  exists: boolean
  size: number
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

  async getSchedules(): Promise<BackupSchedule[]> {
    try {
      const json = await GetBackupSchedules()
      return JSON.parse(json) as BackupSchedule[]
    } catch {
      return []
    }
  },

  async setSchedules(schedules: BackupSchedule[]): Promise<void> {
    await SetBackupSchedules(JSON.stringify(schedules))
  },

  async deleteBackup(path: string): Promise<{ success: boolean; error?: string }> {
    try {
      const json = await DeleteBackup(path)
      return JSON.parse(json) as { success: boolean; error?: string }
    } catch (e) {
      return { success: false, error: e instanceof Error ? e.message : 'Delete failed' }
    }
  },

  async verifyBackup(path: string): Promise<VerifyResult> {
    try {
      const json = await VerifyBackup(path)
      return JSON.parse(json) as VerifyResult
    } catch {
      return { exists: false, size: 0 }
    }
  },
}
