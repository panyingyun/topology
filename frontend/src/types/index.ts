// Database connection types
export interface Connection {
  id: string;
  name: string;
  type: DatabaseType;
  host: string;
  port: number;
  username: string;
  password?: string;
  database?: string;
  status: ConnectionStatus;
  group?: string;
  savePassword?: boolean;
}

export type DatabaseType = 'postgresql' | 'mysql' | 'redis' | 'mongodb';
export type ConnectionStatus = 'connected' | 'disconnected' | 'connecting' | 'error';

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
  isPrimaryKey: boolean;
  isNotNull: boolean;
  isUnique: boolean;
  defaultValue?: string;
  length?: number;
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

export interface TableSchema {
  name: string;
  columns: Column[];
  indexes: Index[];
  foreignKeys: ForeignKey[];
}

// Query and result types
export interface QueryResult {
  columns: string[];
  rows: Record<string, any>[];
  rowCount: number;
  executionTime?: number;
  error?: string;
}

export interface OptimizedSQL {
  original: string;
  optimized: string;
  suggestions: string[];
  performanceGain?: string;
}

// Import/Export types
export type ImportFormat = 'csv' | 'json' | 'xlsx' | 'sql';
export type ExportFormat = 'csv' | 'excel' | 'sql' | 'parquet' | 'xml';

export interface ImportOptions {
  format: ImportFormat;
  encoding: string;
  delimiter?: string;
  hasHeader?: boolean;
  tableName: string;
}

export interface ExportOptions {
  format: ExportFormat;
  tableName: string;
  rowLimit?: number;
  dateFilter?: {
    column: string;
    startDate?: string;
    endDate?: string;
  };
}

// Table data types
export interface TableData {
  columns: string[];
  rows: Record<string, any>[];
  totalRows: number;
  page: number;
  pageSize: number;
}

export interface TableDataUpdate {
  rowIndex: number;
  column: string;
  value: any;
}
