import { StartMonitor as StartMonitorGo, StopMonitor as StopMonitorGo } from '../../wailsjs/go/main/App'

/**
 * Start live monitoring for a MySQL connection. Backend will emit "live-stats" events every 5s.
 * Returns JSON string; if it contains "error" key, monitoring failed (e.g. not MySQL).
 */
export async function startMonitor(connectionId: string): Promise<{ error?: string }> {
  const raw = await StartMonitorGo(connectionId)
  try {
    return JSON.parse(raw || '{}') as { error?: string }
  } catch {
    return {}
  }
}

/**
 * Stop live monitoring for the given connection.
 */
export async function stopMonitor(connectionId: string): Promise<void> {
  await StopMonitorGo(connectionId)
}
