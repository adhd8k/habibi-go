import axios from 'axios'
import type {
  Project,
  Session,
  Agent,
  AgentCommand,
  AgentStatus,
  CreateProjectRequest,
  CreateSessionRequest,
  CreateAgentRequest,
  ExecuteCommandRequest,
} from '../types'

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add response interceptor for debugging
api.interceptors.response.use(
  (response) => {
    console.log('API Response:', response.config.url, response.data)
    return response
  },
  (error) => {
    console.error('API Error:', error.response?.data || error.message)
    return Promise.reject(error)
  }
)

// Projects API
export const projectsApi = {
  list: () => api.get<Project[]>('/projects'),
  get: (id: number) => api.get<Project>(`/projects/${id}`),
  create: (data: CreateProjectRequest) => api.post<Project>('/projects', data),
  update: (id: number, data: Partial<Project>) => api.put<Project>(`/projects/${id}`, data),
  delete: (id: number) => api.delete(`/projects/${id}`),
}

// Sessions API
export const sessionsApi = {
  list: (projectId?: number) => 
    api.get<Session[]>('/sessions', { params: { project_id: projectId } }),
  get: (id: number) => api.get<Session>(`/sessions/${id}`),
  create: (data: CreateSessionRequest) => api.post<Session>('/sessions', data),
  update: (id: number, data: Partial<Session>) => api.put<Session>(`/sessions/${id}`, data),
  delete: (id: number) => api.delete(`/sessions/${id}`),
  getDiffs: (id: number) => api.get<any>(`/sessions/${id}/diffs`),
  rebase: (id: number) => api.post(`/sessions/${id}/rebase`),
  push: (id: number, remoteBranch?: string) => 
    api.post(`/sessions/${id}/push`, { remote_branch: remoteBranch }),
  close: (id: number) => api.post(`/sessions/${id}/close`),
}

export interface DiffFile {
  path: string
  status: 'added' | 'modified' | 'deleted'
  additions: number
  deletions: number
  diff: string
}

// Agents API
export const agentsApi = {
  list: (sessionId?: number) => 
    api.get<Agent[]>('/agents', { params: { session_id: sessionId } }),
  get: (id: number) => api.get<Agent>(`/agents/${id}`),
  create: (data: CreateAgentRequest) => api.post<Agent>('/agents', data),
  status: (id: number) => api.get<AgentStatus>(`/agents/${id}/status`),
  stop: (id: number) => api.post(`/agents/${id}/stop`),
  restart: (id: number) => api.post<Agent>(`/agents/${id}/restart`),
  execute: (data: ExecuteCommandRequest) => 
    api.post<AgentCommand>(`/agents/${data.agent_id}/execute`, { command: data.command }),
  logs: (id: number, since?: string) => 
    api.get<string[]>(`/agents/${id}/logs`, { params: { since } }),
  chatHistory: (id: number, limit?: number) =>
    api.get<{ messages: ChatMessage[] }>(`/agents/${id}/chat`, { params: { limit } }),
  deleteChatHistory: (id: number) =>
    api.delete(`/agents/${id}/chat`),
}

export interface ChatMessage {
  id: number
  agent_id: number
  role: 'user' | 'assistant' | 'system'
  content: string
  created_at: string
}

export default api