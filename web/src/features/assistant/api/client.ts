import api from '../../../api/client'
import type { SlashCommand, CommandResult } from '../types'

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