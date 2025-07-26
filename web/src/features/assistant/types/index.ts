export interface SlashCommand {
  name: string
  description: string
  path?: string
  is_builtin: boolean
  category?: string
  arguments?: string[]
}

export interface CommandResult {
  type: 'clear_chat' | 'show_modal' | 'claude_message' | 'error' | 'show_help' | 'status' | 'config' | 'info' | 'action' | 'vim_mode' | 'compact'
  data: any
}

export interface FileMention {
  path: string
  name: string
  type: 'file' | 'directory'
  size?: number
  lastModified?: string
}