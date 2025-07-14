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


export interface ApiError {
  error: string
  details?: string
}