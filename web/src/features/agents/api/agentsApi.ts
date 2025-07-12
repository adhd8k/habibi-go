import { baseApi, ApiResponse } from '../../../services/api/baseApi'
import { 
  Agent, 
  ChatMessage, 
  CreateAgentRequest,
  AgentSchema,
  ChatMessageSchema,
  CreateAgentRequestSchema 
} from '../../../shared/types/schemas'

export const agentsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getAgents: builder.query<Agent[], number | void>({
      query: (sessionId) => sessionId
        ? `/agents?session_id=${sessionId}`
        : '/agents',
      transformResponse: (response: ApiResponse<Agent[]>) => {
        if (response.data) {
          return response.data.map(agent => AgentSchema.parse(agent))
        }
        return []
      },
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Agent' as const, id })),
              { type: 'Agent', id: 'LIST' },
            ]
          : [{ type: 'Agent', id: 'LIST' }],
    }),

    getAgent: builder.query<Agent, number>({
      query: (id) => `/agents/${id}`,
      transformResponse: (response: ApiResponse<Agent>) => {
        if (response.data) {
          return AgentSchema.parse(response.data)
        }
        throw new Error('Agent not found')
      },
      providesTags: (_result, _error, id) => [{ type: 'Agent', id }],
    }),

    createAgent: builder.mutation<Agent, CreateAgentRequest>({
      query: (agent) => ({
        url: '/agents',
        method: 'POST',
        body: CreateAgentRequestSchema.parse(agent),
      }),
      transformResponse: (response: ApiResponse<Agent>) => {
        if (response.data) {
          return AgentSchema.parse(response.data)
        }
        throw new Error('Failed to create agent')
      },
      invalidatesTags: [
        { type: 'Agent', id: 'LIST' },
        { type: 'Session', id: 'LIST' },
      ],
    }),

    stopAgent: builder.mutation<void, number>({
      query: (id) => ({
        url: `/agents/${id}/stop`,
        method: 'POST',
      }),
      invalidatesTags: (_result, _error, id) => [
        { type: 'Agent', id },
        { type: 'Agent', id: 'LIST' },
      ],
    }),

    restartAgent: builder.mutation<Agent, number>({
      query: (id) => ({
        url: `/agents/${id}/restart`,
        method: 'POST',
      }),
      transformResponse: (response: ApiResponse<Agent>) => {
        if (response.data) {
          return AgentSchema.parse(response.data)
        }
        throw new Error('Failed to restart agent')
      },
      invalidatesTags: (_result, _error, id) => [
        { type: 'Agent', id },
        { type: 'Agent', id: 'LIST' },
      ],
    }),

    sendAgentMessage: builder.mutation<void, { agentId: number; message: string }>({
      query: ({ agentId, message }) => ({
        url: `/agents/${agentId}/send`,
        method: 'POST',
        body: { message },
      }),
      invalidatesTags: (_result, _error, { agentId }) => [
        { type: 'Chat', id: `agent-${agentId}` },
      ],
    }),

    getAgentMessages: builder.query<ChatMessage[], number>({
      query: (agentId) => `/agents/${agentId}/messages`,
      transformResponse: (response: ApiResponse<ChatMessage[]>) => {
        if (response.data) {
          return response.data.map(msg => ChatMessageSchema.parse(msg))
        }
        return []
      },
      providesTags: (_result, _error, agentId) => [
        { type: 'Chat', id: `agent-${agentId}` },
      ],
    }),

    clearAgentMessages: builder.mutation<void, number>({
      query: (agentId) => ({
        url: `/agents/${agentId}/messages`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, agentId) => [
        { type: 'Chat', id: `agent-${agentId}` },
      ],
    }),
  }),
})

export const {
  useGetAgentsQuery,
  useGetAgentQuery,
  useCreateAgentMutation,
  useStopAgentMutation,
  useRestartAgentMutation,
  useSendAgentMessageMutation,
  useGetAgentMessagesQuery,
  useClearAgentMessagesMutation,
} = agentsApi