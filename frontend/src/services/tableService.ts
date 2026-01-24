import { Table, TableData, TableSchema, Column, Index, ForeignKey } from '../types';

// Mock tables data
const mockTables: Record<string, Table[]> = {
  '1': [
    { name: 'users', type: 'table', rowCount: 1250 },
    { name: 'orders', type: 'table', rowCount: 3420 },
    { name: 'transactions', type: 'table', rowCount: 8900 },
    { name: 'products', type: 'table', rowCount: 560 },
  ],
  '2': [
    { name: 'analytics_events', type: 'table', rowCount: 50000 },
    { name: 'user_sessions', type: 'table', rowCount: 12000 },
  ],
  '3': [
    { name: 'test_users', type: 'table', rowCount: 100 },
  ],
};

// Mock table schemas
const mockSchemas: Record<string, TableSchema> = {
  'users': {
    name: 'users',
    columns: [
      {
        name: 'id',
        type: 'BIGINT (AUTO_INC)',
        isPrimaryKey: true,
        isNotNull: true,
        isUnique: true,
      },
      {
        name: 'email',
        type: 'VARCHAR(255)',
        isPrimaryKey: false,
        isNotNull: true,
        isUnique: true,
      },
      {
        name: 'name',
        type: 'VARCHAR(255)',
        isPrimaryKey: false,
        isNotNull: false,
        isUnique: false,
      },
      {
        name: 'created_at',
        type: 'TIMESTAMP',
        isPrimaryKey: false,
        isNotNull: true,
        isUnique: false,
        defaultValue: 'CURRENT_TIMESTAMP',
      },
    ],
    indexes: [
      {
        name: 'PRIMARY_KEY',
        columns: ['id'],
        isUnique: true,
        type: 'PRIMARY',
      },
      {
        name: 'idx_users_email',
        columns: ['email'],
        isUnique: true,
        type: 'UNIQUE',
      },
    ],
    foreignKeys: [],
  },
  'transactions': {
    name: 'transactions',
    columns: [
      {
        name: 'id',
        type: 'VARCHAR(50)',
        isPrimaryKey: true,
        isNotNull: true,
        isUnique: true,
      },
      {
        name: 'user_id',
        type: 'VARCHAR(50)',
        isPrimaryKey: false,
        isNotNull: true,
        isUnique: false,
      },
      {
        name: 'amount',
        type: 'DECIMAL(10,2)',
        isPrimaryKey: false,
        isNotNull: true,
        isUnique: false,
      },
      {
        name: 'status',
        type: 'VARCHAR(20)',
        isPrimaryKey: false,
        isNotNull: true,
        isUnique: false,
      },
      {
        name: 'created_at',
        type: 'TIMESTAMP',
        isPrimaryKey: false,
        isNotNull: true,
        isUnique: false,
      },
    ],
    indexes: [
      {
        name: 'PRIMARY_KEY',
        columns: ['id'],
        isUnique: true,
        type: 'PRIMARY',
      },
      {
        name: 'idx_transactions_created_at',
        columns: ['created_at'],
        isUnique: false,
        type: 'INDEX',
      },
    ],
    foreignKeys: [],
  },
};

export const tableService = {
  async getTables(connectionId: string): Promise<Table[]> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(mockTables[connectionId] || []);
      }, 300);
    });
  },

  async getTableData(
    connectionId: string,
    tableName: string,
    limit: number = 100,
    offset: number = 0
  ): Promise<TableData> {
    return new Promise((resolve) => {
      setTimeout(() => {
        const table = mockTables[connectionId]?.find(t => t.name === tableName);
        const totalRows = table?.rowCount || 0;

        // Mock data based on table name
        let mockRows: Record<string, any>[] = [];
        if (tableName === 'transactions') {
          mockRows = [
            {
              id: 'TXN_8841',
              user_id: 'USR_2291',
              amount: '$1,250.00',
              status: 'COMPLETED',
              created_at: '2023-10-24 14:22:10',
            },
            {
              id: 'TXN_8842',
              user_id: 'USR_4102',
              amount: '$745.50',
              status: 'COMPLETED',
              created_at: '2023-10-24 14:15:05',
            },
            {
              id: 'TXN_8843',
              user_id: 'USR_8812',
              amount: '$920.00',
              status: 'PENDING',
              created_at: '2023-10-24 14:12:44',
            },
            {
              id: 'TXN_8844',
              user_id: 'USR_3301',
              amount: '$2,100.00',
              status: 'COMPLETED',
              created_at: '2023-10-24 14:10:02',
            },
          ];
        } else if (tableName === 'users') {
          mockRows = [
            {
              id: 1,
              email: 'user1@example.com',
              name: 'John Doe',
              created_at: '2023-01-15 10:30:00',
            },
            {
              id: 2,
              email: 'user2@example.com',
              name: 'Jane Smith',
              created_at: '2023-02-20 14:45:00',
            },
          ];
        }

        resolve({
          columns: mockRows.length > 0 ? Object.keys(mockRows[0]) : [],
          rows: mockRows.slice(offset, offset + limit),
          totalRows,
          page: Math.floor(offset / limit) + 1,
          pageSize: limit,
        });
      }, 400);
    });
  },

  async updateTableData(
    connectionId: string,
    tableName: string,
    updates: Array<{ rowIndex: number; column: string; value: any }>
  ): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        // Mock update - in real implementation, this would update the database
        console.log('Updating table data:', { connectionId, tableName, updates });
        resolve();
      }, 500);
    });
  },

  async getTableSchema(connectionId: string, tableName: string): Promise<TableSchema> {
    return new Promise((resolve) => {
      setTimeout(() => {
        const schema = mockSchemas[tableName] || {
          name: tableName,
          columns: [],
          indexes: [],
          foreignKeys: [],
        };
        resolve(schema);
      }, 300);
    });
  },

  async updateTableSchema(
    connectionId: string,
    tableName: string,
    schema: TableSchema
  ): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockSchemas[tableName] = schema;
        resolve();
      }, 500);
    });
  },
};
