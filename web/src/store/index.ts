import { create } from 'zustand'
import type { Project, Session, Agent } from '../types'

interface AppState {
  // Current selections
  currentProject: Project | null
  currentSession: Session | null
  currentAgent: Agent | null
  
  // UI state
  sidebarOpen: boolean
  
  // Actions
  setCurrentProject: (project: Project | null) => void
  setCurrentSession: (session: Session | null) => void
  setCurrentAgent: (agent: Agent | null) => void
  setSidebarOpen: (open: boolean) => void
}

export const useAppStore = create<AppState>((set) => ({
  // Initial state
  currentProject: null,
  currentSession: null,
  currentAgent: null,
  sidebarOpen: true,
  
  // Actions
  setCurrentProject: (project) => set({ currentProject: project }),
  setCurrentSession: (session) => set({ currentSession: session }),
  setCurrentAgent: (agent) => set({ currentAgent: agent }),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
}))