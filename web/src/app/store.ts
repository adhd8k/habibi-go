import { configureStore } from '@reduxjs/toolkit'
import { setupListeners } from '@reduxjs/toolkit/query'
import { websocketMiddleware } from './middleware/websocket'
import authReducer from '../features/auth/slice/authSlice'
import projectsReducer from '../features/projects/slice/projectsSlice'
import sessionsReducer from '../features/sessions/slice/sessionsSlice'
import agentsReducer from '../features/agents/slice/agentsSlice'
import { baseApi } from '../services/api/baseApi'

export const store = configureStore({
  reducer: {
    auth: authReducer,
    projects: projectsReducer,
    sessions: sessionsReducer,
    agents: agentsReducer,
    [baseApi.reducerPath]: baseApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types for websocket
        ignoredActions: ['websocket/connect', 'websocket/disconnect', 'websocket/send'],
        // Ignore these paths in the state
        ignoredPaths: ['websocket.socket'],
      },
    })
      .concat(baseApi.middleware)
      .concat(websocketMiddleware),
})

// Enable refetchOnFocus/refetchOnReconnect behaviors
setupListeners(store.dispatch)

export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch