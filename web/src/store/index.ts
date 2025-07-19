import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import type { Project, Session } from '../types'

interface AppState {
  // Current selections
  currentProject: Project | null
  currentSession: Session | null
  
  // UI state
  sidebarOpen: boolean
  isSessionLoading: boolean
  sessionError: string | null
  darkMode: boolean
  
  // Session management
  recentSessions: Session[]
  
  // Actions
  setCurrentProject: (project: Project | null) => void
  setCurrentSession: (session: Session | null) => void
  setSidebarOpen: (open: boolean) => void
  setSessionLoading: (loading: boolean) => void
  setSessionError: (error: string | null) => void
  setDarkMode: (enabled: boolean) => void
  addRecentSession: (session: Session) => void
  updateSession: (id: number, updates: Partial<Session>) => void
  removeSession: (id: number) => void
}

// Initialize dark mode from localStorage
const initializeDarkMode = () => {
  const saved = localStorage.getItem('darkMode')
  const enabled = saved ? JSON.parse(saved) : false
  if (enabled) {
    document.documentElement.classList.add('dark')
  }
  return enabled
}

export const useAppStore = create<AppState>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    currentProject: null,
    currentSession: null,
    sidebarOpen: true,
    isSessionLoading: false,
    sessionError: null,
    darkMode: initializeDarkMode(),
    recentSessions: [],
    
    // Actions
    setCurrentProject: (project) => set({ currentProject: project }),
    
    setCurrentSession: (session) => {
      set({ currentSession: session, sessionError: null })
      // Add to recent sessions if not already there
      if (session && !get().recentSessions.find(s => s.id === session.id)) {
        get().addRecentSession(session)
      }
    },
    
    setSidebarOpen: (open) => set({ sidebarOpen: open }),
    setSessionLoading: (loading) => set({ isSessionLoading: loading }),
    setSessionError: (error) => set({ sessionError: error }),
    setDarkMode: (enabled) => {
      set({ darkMode: enabled })
      localStorage.setItem('darkMode', JSON.stringify(enabled))
      if (enabled) {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
    },
    
    addRecentSession: (session) => set((state) => ({
      recentSessions: [session, ...state.recentSessions.filter(s => s.id !== session.id)].slice(0, 5)
    })),
    
    updateSession: (id, updates) => set((state) => ({
      currentSession: state.currentSession?.id === id 
        ? { ...state.currentSession, ...updates } 
        : state.currentSession,
      recentSessions: state.recentSessions.map(s => 
        s.id === id ? { ...s, ...updates } : s
      )
    })),
    
    removeSession: (id) => set((state) => ({
      currentSession: state.currentSession?.id === id ? null : state.currentSession,
      recentSessions: state.recentSessions.filter(s => s.id !== id)
    })),
  }))
)