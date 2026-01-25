import type { Snippet } from '../types'
import { GetSnippets, SaveSnippet, DeleteSnippet } from '../../wailsjs/go/main/App'

export const snippetService = {
  async getSnippets(): Promise<Snippet[]> {
    try {
      const json = await GetSnippets()
      return JSON.parse(json) as Snippet[]
    } catch (error) {
      console.error('Failed to get snippets:', error)
      return []
    }
  },

  async saveSnippet(alias: string, sql: string): Promise<void> {
    await SaveSnippet(alias, sql)
  },

  async deleteSnippet(id: string): Promise<void> {
    await DeleteSnippet(id)
  },
}
