// Connection types
export interface Connection {
  id: string;
  name: string;
  type: DatabaseType;
  host: string;
  port: number;
  username: string;
  password?: string;
  database?: string;
  useSSL?: boolean;
  sshTunnel?: SSHTunnel;
  status: ConnectionStatus;
  createdAt?: string;
  readOnly?: boolean;
}

export type DatabaseType = 'mysql' | 'postgresql' | 'sqlite';
export type ConnectionStatus = 'connected' | 'disconnected' | 'connecting' | 'error';

export interface SSHTunnel {
  enabled: boolean;
  host?: string;
  port?: number;
  username?: string;
  password?: string;
  privateKey?: string;
}

// Database and table types
export interface Database {
  name: string;
  connectionId: string;
  tables: Table[];
}

export interface Table {
  name: string;
  schema?: string;
  type: 'table' | 'view';
  rowCount?: number;
}

export interface Column {
  name: string;
  type: string;
  nullable: boolean;
  defaultValue?: string;
  isPrimaryKey: boolean;
  isUnique: boolean;
}

export interface TableSchema {
  name: string;
  columns: Column[];
  indexes: Index[];
  foreignKeys: ForeignKey[];
}

export interface Index {
  name: string;
  columns: string[];
  isUnique: boolean;
  type: string;
}

export interface ForeignKey {
  name: string;
  columns: string[];
  referencedTable: string;
  referencedColumns: string[];
  onDelete?: string;
  onUpdate?: string;
}

// Query and result types
export interface QueryResult {
  columns: string[];
  rows: Record<string, any>[];
  rowCount: number;
  executionTime?: number;
  affectedRows?: number;
  error?: string;
  cached?: boolean;
}

// Table data types
export interface TableData {
  columns: string[];
  rows: Record<string, any>[];
  totalRows: number;
  page: number;
  pageSize: number;
}

export interface UpdateRecord {
  rowIndex: number;
  column: string;
  oldValue: any;
  newValue: any;
}

// Tab types
export type TabType = 'query' | 'table';

export interface TabItem {
  id: string;
  type: TabType;
  title: string;
  connectionId?: string;
  database?: string;
  tableName?: string;
  sql?: string;
  /** One-shot: inject this SQL into the query editor when opening from table context menu */
  initialSql?: string;
  queryResult?: QueryResult;
}

// Export types
export type ExportFormat = 'csv' | 'json' | 'sql';

// Query history types
export interface QueryHistory {
  id: string
  connectionId: string
  sql: string
  executedAt: string
  success: boolean
  duration?: number
  rowCount?: number
}

// SQL snippet (saved fragment with alias)
export interface Snippet {
  id: string
  alias: string
  sql: string
  createdAt: string
}

// Import types
export type ImportFormat = 'csv' | 'json'

export interface ImportPreview {
  columns: string[]
  rows: Record<string, any>[]
  error?: string
}

export interface ImportResult {
  success: boolean
  inserted?: number
  totalRows?: number
  error?: string
}

// SQL Analysis types
export interface SQLAnalysis {
  queryType: string
  suggestions: string[]
  warnings: string[]
  performance: {
    estimatedComplexity?: string
    indexUsage?: string
  }
}

// Execution plan (EXPLAIN) for visualization
export interface ExecutionPlanNode {
  id: string
  parentId?: string | null
  type: string
  label: string
  detail?: string
  rows?: number
  cost?: string
  extra?: string
  fullTableScan: boolean
  indexUsed: boolean
}

export interface ExecutionPlanResult {
  nodes: ExecutionPlanNode[]
  summary: {
    totalDurationMs?: number
    warnings?: string[]
  }
  error?: string
}

export interface IndexSuggestion {
  table: string
  columns?: string[]
  createIndex: string
  reason: string
}

// Live monitor (MySQL): real-time stats pushed via "live-stats" event
export interface ProcessItem {
  id: string
  user: string
  host: string
  db: string
  command: string
  time: number
  state: string
  info: string
}

export interface LiveStatsPayload {
  connectionId: string
  threadsConnected: number
  processList: ProcessItem[]
  error?: string
}
