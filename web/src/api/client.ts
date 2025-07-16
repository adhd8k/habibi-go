import axios from 'axios'
import type {
  Project,
  Session,
  CreateProjectRequest,
  CreateSessionRequest,
} from '../types'

// Get auth credentials from localStorage or environment
const getAuthHeader = () => {
  const auth = localStorage.getItem('habibi_auth')
  if (auth) {
    const { username, password } = JSON.parse(auth)
    return 'Basic ' + btoa(`${username}:${password}`)
  }
  return null
}

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth header to all requests
api.interceptors.request.use(
  (config) => {
    const authHeader = getAuthHeader()
    if (authHeader) {
      config.headers.Authorization = authHeader
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Add response interceptor for debugging and auth handling
api.interceptors.response.use(
  (response) => {
    console.log('API Response:', response.config.url, response.data)
    return response
  },
  (error) => {
    console.error('API Error:', error.response?.data || error.message)
    
    // Handle 401 Unauthorized
    if (error.response?.status === 401) {
      // Clear stored auth
      localStorage.removeItem('habibi_auth')
      
      // Prompt for credentials
      const username = prompt('Username:')
      const password = prompt('Password:')
      
      if (username && password) {
        // Store credentials
        localStorage.setItem('habibi_auth', JSON.stringify({ username, password }))
        
        // Retry the request
        const config = error.config
        config.headers.Authorization = 'Basic ' + btoa(`${username}:${password}`)
        return api(config)
      }
    }
    
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
  merge: (id: number, targetBranch?: string) => 
    api.post(`/sessions/${id}/merge`, { target_branch: targetBranch }),
  mergeToOriginal: (id: number) => api.post(`/sessions/${id}/merge-to-original`),
  close: (id: number) => api.post(`/sessions/${id}/close`),
  openWithEditor: (id: number) => api.post(`/sessions/${id}/open-editor`),
}

export interface DiffFile {
  path: string
  status: 'added' | 'modified' | 'deleted'
  additions: number
  deletions: number
  diff: string
}

export interface ChatMessage {
  id: number
  session_id: number
  role: 'user' | 'assistant' | 'system' | 'tool_use' | 'tool_result'
  content: string
  created_at: string
  tool_name?: string
  tool_input?: string
  tool_use_id?: string
  tool_content?: string
}

export default api