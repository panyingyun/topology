import type { Connection } from '../types'

import {
  GetConnections,
  CreateConnection,
  TestConnection,
  DeleteConnection,
  UpdateConnection,
  ReconnectConnection,
} from '../../wailsjs/go/main/App'

export const connectionService = {
  async getConnections(): Promise<Connection[]> {
    try {
      const result = await GetConnections()
      return JSON.parse(result)
    } catch (error) {
      console.error('Failed to get connections:', error)
      return []
    }
  },

  async createConnection(connection: Omit<Connection, 'id' | 'status' | 'createdAt'>): Promise<Connection> {
    const connJSON = JSON.stringify(connection)
    await CreateConnection(connJSON)
    return {
      ...connection,
      id: Date.now().toString(),
      status: 'disconnected',
      createdAt: new Date().toISOString(),
    }
  },

  async updateConnection(connection: Connection): Promise<void> {
    const connJSON = JSON.stringify(connection)
    await UpdateConnection(connJSON)
  },

  async testConnection(connection: Omit<Connection, 'id' | 'status' | 'createdAt'>): Promise<boolean> {
    try {
      const connJSON = JSON.stringify(connection)
      return await TestConnection(connJSON)
    } catch (error) {
      console.error('Failed to test connection:', error)
      return false
    }
  },

  async reconnectConnection(id: string): Promise<void> {
    await ReconnectConnection(id)
  },

  async deleteConnection(id: string): Promise<void> {
    await DeleteConnection(id)
  },
}
