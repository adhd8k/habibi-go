import { createSlice, PayloadAction } from '@reduxjs/toolkit'
import { Agent } from '../../../shared/types/schemas'
import { WebSocketMessage } from '../../../app/middleware/websocket'

interface AgentsState {
  currentAgent: Agent | null
  streamingMessages: Record<number, string> // agent_id -> accumulated message
  isStreaming: Record<number, boolean> // agent_id -> streaming status
}

const initialState: AgentsState = {
  currentAgent: null,
  streamingMessages: {},
  isStreaming: {},
}

export const agentsSlice = createSlice({
  name: 'agents',
  initialState,
  reducers: {
    setCurrentAgent: (state, action: PayloadAction<Agent | null>) => {
      state.currentAgent = action.payload
    },
    
    startStreaming: (state, action: PayloadAction<number>) => {
      const agentId = action.payload
      state.isStreaming[agentId] = true
      state.streamingMessages[agentId] = ''
    },
    
    appendStreamingMessage: (state, action: PayloadAction<{ agentId: number; content: string }>) => {
      const { agentId, content } = action.payload
      state.streamingMessages[agentId] = (state.streamingMessages[agentId] || '') + content
    },
    
    stopStreaming: (state, action: PayloadAction<number>) => {
      const agentId = action.payload
      state.isStreaming[agentId] = false
      delete state.streamingMessages[agentId]
    },
    
    clearStreamingMessage: (state, action: PayloadAction<number>) => {
      const agentId = action.payload
      delete state.streamingMessages[agentId]
    },
  },
  extraReducers: (builder) => {
    // Handle WebSocket messages
    builder.addCase('websocket/messageReceived' as any, (state, action: PayloadAction<WebSocketMessage>) => {
      const message = action.payload
      
      switch (message.type) {
        case 'agent_output':
          if (message.agent_id) {
            if (message.data?.type === 'start') {
              state.isStreaming[message.agent_id] = true
              state.streamingMessages[message.agent_id] = ''
            } else if (message.data?.type === 'data' && message.data?.content) {
              state.streamingMessages[message.agent_id] = 
                (state.streamingMessages[message.agent_id] || '') + message.data.content
            } else if (message.data?.type === 'end') {
              state.isStreaming[message.agent_id] = false
              delete state.streamingMessages[message.agent_id]
            }
          }
          break
          
        case 'agent_status':
          if (message.agent_id && state.currentAgent?.id === message.agent_id) {
            state.currentAgent = {
              ...state.currentAgent,
              status: message.data?.status || state.currentAgent.status,
            }
          }
          break
      }
    })
  },
})

export const { 
  setCurrentAgent,
  startStreaming,
  appendStreamingMessage,
  stopStreaming,
  clearStreamingMessage,
} = agentsSlice.actions

export default agentsSlice.reducer

// Selectors
export const selectCurrentAgent = (state: { agents: AgentsState }) => 
  state.agents.currentAgent

export const selectStreamingMessage = (state: { agents: AgentsState }, agentId: number) => 
  state.agents.streamingMessages[agentId] || ''

export const selectIsStreaming = (state: { agents: AgentsState }, agentId: number) => 
  state.agents.isStreaming[agentId] || false