import { z } from 'zod'

// Project schemas
export const ProjectConfigSchema = z.object({
  git_remote: z.string().optional(),
  agent_defaults: z.record(z.string()).optional(),
  notifications: z.boolean().optional(),
  current_branch: z.string().optional(),
  ssh_host: z.string().optional(),
  ssh_port: z.number().optional(),
  ssh_key_path: z.string().optional(),
  remote_project_path: z.string().optional(),
  environment_vars: z.record(z.string()).optional(),
  remote_setup_cmd: z.string().optional(),
})

export const ProjectSchema = z.object({
  id: z.number(),
  name: z.string(),
  path: z.string(),
  repository_url: z.string().optional(),
  default_branch: z.string(),
  setup_command: z.string().optional(),
  config: ProjectConfigSchema,
  created_at: z.string(),
  updated_at: z.string(),
})

// Session schemas
export const SessionSchema = z.object({
  id: z.number(),
  project_id: z.number(),
  name: z.string(),
  branch_name: z.string(),
  original_branch: z.string().optional(),
  worktree_path: z.string(),
  status: z.enum(['active', 'paused', 'stopped']),
  config: z.record(z.any()),
  created_at: z.string(),
  last_used_at: z.string(),
  last_activity_at: z.string().optional(),
  activity_status: z.enum(['idle', 'streaming', 'new', 'viewed']).optional(),
  last_viewed_at: z.string().optional(),
})

// Agent schemas
export const AgentSchema = z.object({
  id: z.number(),
  session_id: z.number(),
  agent_type: z.string(),
  pid: z.number(),
  status: z.enum(['starting', 'running', 'stopped', 'failed']),
  config: z.record(z.any()),
  command: z.string(),
  working_directory: z.string(),
  communication_method: z.string(),
  input_pipe_path: z.string().optional(),
  output_pipe_path: z.string().optional(),
  last_heartbeat: z.string().optional(),
  resource_usage: z.record(z.any()),
  started_at: z.string(),
  stopped_at: z.string().optional(),
})

// Chat message schemas
export const ChatMessageSchema = z.object({
  id: z.number(),
  agent_id: z.number(),
  role: z.enum(['user', 'assistant', 'system', 'tool_use', 'tool_result']),
  content: z.string(),
  created_at: z.string(),
  tool_name: z.string().optional(),
  tool_input: z.string().optional(),
  tool_use_id: z.string().optional(),
  tool_content: z.string().optional(),
})

// Request schemas
export const CreateProjectRequestSchema = z.object({
  name: z.string().min(1),
  path: z.string().min(1),
  repository_url: z.string().optional(),
  default_branch: z.string().optional(),
  setup_command: z.string().optional(),
  config: ProjectConfigSchema.optional(),
})

export const CreateSessionRequestSchema = z.object({
  project_id: z.number(),
  name: z.string().min(1),
  branch_name: z.string().min(1),
  base_branch: z.string().optional(),
})

export const CreateAgentRequestSchema = z.object({
  session_id: z.number(),
  agent_type: z.string(),
  command: z.string(),
  config: z.record(z.any()).optional(),
})

// Type exports
export type Project = z.infer<typeof ProjectSchema>
export type Session = z.infer<typeof SessionSchema>
export type Agent = z.infer<typeof AgentSchema>
export type ChatMessage = z.infer<typeof ChatMessageSchema>
export type CreateProjectRequest = z.infer<typeof CreateProjectRequestSchema>
export type CreateSessionRequest = z.infer<typeof CreateSessionRequestSchema>
export type CreateAgentRequest = z.infer<typeof CreateAgentRequestSchema>