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
