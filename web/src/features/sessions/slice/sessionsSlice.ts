import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { Session } from '../../../shared/types/schemas'
import { WebSocketMessage } from '../../../app/middleware/websocket'

interface SessionsState {
  currentSession: Session | null
  recentSessions: Session[]
  activityStatus: Record<number, string> // session_id -> activity_status
}

const initialState: SessionsState = {
  currentSession: null,
  recentSessions: [],
  activityStatus: {},
}

export const sessionsSlice = createSlice({
  name: 'sessions',
  initialState,
  reducers: {
    setCurrentSession: (state, action: PayloadAction<Session | null>) => {
      state.currentSession = action.payload
      
      // Add to recent sessions if not already there
      if (action.payload && !state.recentSessions.find(s => s.id === action.payload!.id)) {
        state.recentSessions = [
          action.payload,
          ...state.recentSessions.filter(s => s.id !== action.payload!.id)
        ].slice(0, 5)
      }
    },
    
    updateSessionActivity: (state, action: PayloadAction<{ sessionId: number; status: string }>) => {
      const { sessionId, status } = action.payload
      state.activityStatus[sessionId] = status
      
      // Update current session if it matches
      if (state.currentSession?.id === sessionId) {
        state.currentSession = {
          ...state.currentSession,
          activity_status: status as any,
        }
      }
    },
    
    removeFromRecent: (state, action: PayloadAction<number>) => {
      state.recentSessions = state.recentSessions.filter(s => s.id !== action.payload)
      
      if (state.currentSession?.id === action.payload) {
        state.currentSession = null
      }
    },
    
    clearCurrentSession: (state) => {
      state.currentSession = null
    },
  },
  extraReducers: (builder) => {
    // Handle WebSocket messages
    builder.addCase('websocket/messageReceived', (state, action: PayloadAction<WebSocketMessage>) => {
      const message = action.payload
      
      switch (message.type) {
        case 'session_activity_update':
          if (message.data?.session_id && message.data?.activity_status) {
            state.activityStatus[message.data.session_id] = message.data.activity_status
          }
          break
          
        case 'session_update':
          if (message.data?.session && state.currentSession?.id === message.data.session.id) {
            state.currentSession = message.data.session
          }
          break
          
        case 'session_deleted':
          if (message.data?.session_id) {
            state.recentSessions = state.recentSessions.filter(s => s.id !== message.data.session_id)
            if (state.currentSession?.id === message.data.session_id) {
              state.currentSession = null
            }
          }
          break
      }
    })
  },
})

export const { 
  setCurrentSession, 
  updateSessionActivity, 
  removeFromRecent,
  clearCurrentSession,
} = sessionsSlice.actions

export default sessionsSlice.reducer

// Selectors
export const selectCurrentSession = (state: { sessions: SessionsState }) => 
  state.sessions.currentSession

export const selectRecentSessions = (state: { sessions: SessionsState }) => 
  state.sessions.recentSessions

export const selectSessionActivity = (state: { sessions: SessionsState }, sessionId: number) => 
  state.sessions.activityStatus[sessionId]