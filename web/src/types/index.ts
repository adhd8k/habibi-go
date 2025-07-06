export interface Project {
  id: number
  name: string
  path: string
  repository_url?: string
  default_branch: string
  config: Record<string, any>
  created_at: string
  updated_at: string
}

export interface Session {
  id: number
  project_id: number
  name: string
  branch_name: string
  worktree_path: string
  status: 'active' | 'paused' | 'stopped'
  config: Record<string, any>
  created_at: string
  last_used_at: string
}

export interface Agent {
  id: number
  session_id: number
  agent_type: string
  pid: number
  status: 'starting' | 'running' | 'stopped' | 'failed'
  config: Record<string, any>
  command: string
  working_directory: string
  communication_method: string
  input_pipe_path?: string
  output_pipe_path?: string
  last_heartbeat?: string
  resource_usage: Record<string, any>
  started_at: string
  stopped_at?: string
}

export interface AgentCommand {
  id: number
  agent_id: number
  command_text: string
  response_text?: string
  status: 'pending' | 'completed' | 'failed'
  execution_time_ms?: number
  created_at: string
  completed_at?: string
}

export interface AgentStatus {
  agent: Agent
  is_active: boolean
  process_exists: boolean
  is_healthy: boolean
  last_seen?: string
  process_info?: {
    cpu_percent: number
    memory_mb: number
    status: string
  }
}

export interface CreateProjectRequest {
  name: string
  path: string
  repository_url?: string
  default_branch?: string
}

export interface CreateSessionRequest {
  project_name: string
  session_name: string
  branch_name: string
}

export interface CreateAgentRequest {
  session_id: number
  agent_type: string
  command: string
  config?: Record<string, any>
}

export interface ExecuteCommandRequest {
  agent_id: number
  command: string
}

export interface ApiError {
  error: string
  details?: string
}