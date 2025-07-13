export interface Project {
  id: number
  name: string
  path: string
  repository_url?: string
  default_branch: string
  setup_command?: string
  config: ProjectConfig
  created_at: string
  updated_at: string
}

export interface ProjectConfig {
  git_remote?: string
  agent_defaults?: Record<string, string>
  notifications?: boolean
  current_branch?: string     // Current active branch
  
  // SSH Configuration
  ssh_host?: string           // user@hostname
  ssh_port?: number           // default 22
  ssh_key_path?: string       // path to private key
  remote_project_path?: string // path on remote server
  environment_vars?: Record<string, string> // env vars to set
  remote_setup_cmd?: string   // setup command with variables
}

export interface Session {
  id: number
  project_id: number
  name: string
  branch_name: string
  original_branch?: string
  worktree_path: string
  status: 'active' | 'paused' | 'stopped'
  config: Record<string, any>
  created_at: string
  last_used_at: string
  last_activity_at?: string
  activity_status?: 'idle' | 'streaming' | 'new' | 'viewed'
  last_viewed_at?: string
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
  claude_session_id?: string
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
  setup_command?: string
  config?: ProjectConfig
}

export interface CreateSessionRequest {
  project_id: number
  name: string
  branch_name: string
  base_branch?: string
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