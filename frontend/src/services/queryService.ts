import { QueryResult, OptimizedSQL } from '../types';

export const queryService = {
  async executeQuery(connectionId: string, sql: string): Promise<QueryResult> {
    return new Promise((resolve) => {
      setTimeout(() => {
        // Mock query results
        const mockResults: QueryResult = {
          columns: ['id', 'user_id', 'amount', 'status', 'created_at'],
          rows: [
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
          ],
          rowCount: 1000,
          executionTime: 42,
        };

        // If SQL contains LIMIT, adjust row count
        const limitMatch = sql.match(/LIMIT\s+(\d+)/i);
        if (limitMatch) {
          const limit = parseInt(limitMatch[1]);
          mockResults.rows = mockResults.rows.slice(0, limit);
        }

        resolve(mockResults);
      }, 500);
    });
  },

  async optimizeSQL(sql: string): Promise<OptimizedSQL> {
    return new Promise((resolve) => {
      setTimeout(() => {
        // Mock AI optimization
        const optimized = sql
          .replace(/SELECT\s+\*/gi, 'SELECT id, user_id, amount, status, created_at')
          .replace(/FROM\s+(\w+)/gi, 'FROM $1 USE INDEX (idx_$1_created_at)')
          .replace(/LIMIT\s+(\d+)/gi, (match, limit) => {
            return `ORDER BY created_at DESC\nLIMIT ${limit}`;
          });

        const suggestions = [
          'Added index hint for primary key range',
          'Filtered specific columns instead of SELECT *',
          'Added ORDER BY clause for better performance',
        ];

        resolve({
          original: sql,
          optimized: optimized,
          suggestions,
          performanceGain: '~35% faster',
        });
      }, 1500);
    });
  },
};
