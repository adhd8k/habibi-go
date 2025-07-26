import api from '../../../api/client'
import type { SlashCommand, CommandResult, FileMention } from '../types'

// Slash command API methods
export const slashCommands = {
  // Get available slash commands for a session
  getCommands: async (sessionId: number): Promise<SlashCommand[]> => {
    const response = await api.get<SlashCommand[]>(`/sessions/${sessionId}/commands`)
    return response.data
  },

  // Execute a slash command
  executeCommand: async (sessionId: number, command: string, args?: string): Promise<CommandResult> => {
    const response = await api.post<CommandResult>(`/sessions/${sessionId}/commands`, {
      command,
      args: args || '',
    })
    return response.data
  },
}

// File API methods
export const fileApi = {
  // Search for files
  searchFiles: async (sessionId: number, query: string): Promise<FileMention[]> => {
    const response = await api.get<any[]>(`/sessions/${sessionId}/files/search`, {
      params: { q: query }
    })
    return response.data.map(file => ({
      path: file.relative_path,
      name: file.name,
      type: file.type as 'file' | 'directory',
      size: file.size,
      lastModified: file.last_modified
    }))
  },

  // List files in a directory
  listFiles: async (sessionId: number, path?: string): Promise<FileMention[]> => {
    const response = await api.get<any[]>(`/sessions/${sessionId}/files`, {
      params: path ? { path } : {}
    })
    return response.data.map(file => ({
      path: file.relative_path,
      name: file.name,
      type: file.type as 'file' | 'directory',
      size: file.size,
      lastModified: file.last_modified
    }))
  },
}