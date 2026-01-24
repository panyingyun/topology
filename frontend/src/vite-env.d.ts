/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare global {
  interface Window {
    runtime?: {
      WindowMinimise: () => void
      Quit: () => void
    }
    go: {
      main: {
        App: {
          GetConnections: () => Promise<string>
          CreateConnection: (conn: string) => Promise<void>
          TestConnection: (conn: string) => Promise<boolean>
          DeleteConnection: (id: string) => Promise<void>
          ExecuteQuery: (connectionId: string, sql: string) => Promise<string>
          FormatSQL: (sql: string) => Promise<string>
          GetTables: (connectionId: string) => Promise<string>
          GetTableData: (connectionId: string, tableName: string, limit: number, offset: number) => Promise<string>
          UpdateTableData: (connectionId: string, tableName: string, updates: string) => Promise<void>
          GetTableSchema: (connectionId: string, tableName: string) => Promise<string>
          ExportData: (connectionId: string, tableName: string, format: string) => Promise<string>
        }
      }
    }
  }
}
