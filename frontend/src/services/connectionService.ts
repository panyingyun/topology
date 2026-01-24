import { Connection, ConnectionStatus } from '../types';

// Mock connections data
const mockConnections: Connection[] = [
  {
    id: '1',
    name: 'Main Postgres',
    type: 'postgresql',
    host: 'localhost',
    port: 5432,
    username: 'postgres',
    database: 'ecommerce_prod',
    status: 'connected',
    group: 'Production',
  },
  {
    id: '2',
    name: 'Analytics Replica',
    type: 'postgresql',
    host: 'analytics.example.com',
    port: 5432,
    username: 'analytics',
    database: 'analytics_db',
    status: 'connected',
    group: 'Production',
  },
  {
    id: '3',
    name: 'Staging Cluster',
    type: 'mysql',
    host: 'staging.example.com',
    port: 3306,
    username: 'dev',
    database: 'staging',
    status: 'disconnected',
    group: 'Development',
  },
];

export const connectionService = {
  async getConnections(): Promise<Connection[]> {
    // Simulate API call
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve([...mockConnections]);
      }, 300);
    });
  },

  async createConnection(connection: Omit<Connection, 'id' | 'status'>): Promise<Connection> {
    return new Promise((resolve) => {
      setTimeout(() => {
        const newConnection: Connection = {
          ...connection,
          id: Date.now().toString(),
          status: 'disconnected',
        };
        mockConnections.push(newConnection);
        resolve(newConnection);
      }, 500);
    });
  },

  async testConnection(connection: Omit<Connection, 'id' | 'status'>): Promise<boolean> {
    return new Promise((resolve) => {
      setTimeout(() => {
        // Mock: always succeed for localhost, fail for others
        resolve(connection.host === 'localhost' || connection.host === '127.0.0.1');
      }, 1000);
    });
  },

  async deleteConnection(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        const index = mockConnections.findIndex(c => c.id === id);
        if (index > -1) {
          mockConnections.splice(index, 1);
        }
        resolve();
      }, 300);
    });
  },

  async updateConnectionStatus(id: string, status: ConnectionStatus): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        const connection = mockConnections.find(c => c.id === id);
        if (connection) {
          connection.status = status;
        }
        resolve();
      }, 300);
    });
  },
};
