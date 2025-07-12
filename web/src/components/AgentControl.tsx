import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { agentsApi } from '../api/client'
import { useAppStore } from '../store'
import { Agent } from '../types'
import { ClaudeChat } from './ClaudeChat'

export function AgentControl() {
  const queryClient = useQueryClient()
  const { currentSession } = useAppStore()

  const { data: agents, isLoading } = useQuery({
    queryKey: ['agents', currentSession?.id],
    queryFn: async () => {
      if (!currentSession) return []
      const response = await agentsApi.list(currentSession.id)
      // Handle the wrapped response format {data: [...], success: true}
      const data = response.data as any
      if (data && data.data && Array.isArray(data.data)) {
        return data.data
      }
      // Fallback to direct array if API format changes
      return Array.isArray(response.data) ? response.data : []
    },
    enabled: !!currentSession,
  })

  const stopMutation = useMutation({
    mutationFn: async (id: number) => {
      await agentsApi.stop(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
    },
  })

  const restartMutation = useMutation({
    mutationFn: async (id: number) => {
      const response = await agentsApi.restart(id)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
    },
  })

  if (!currentSession) {
    return (
      <div className="p-4 text-gray-500">
        Select a session to start chatting
      </div>
    )
  }

  // Find the Claude agent for this session
  const claudeAgent = agents?.find((agent: Agent) => 
    agent.agent_type === 'claude-code' && agent.status === 'running'
  )

  if (isLoading) {
    return (
      <div className="p-4">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b">
        <div className="flex justify-between items-center">
          <h2 className="text-lg font-semibold">Claude Assistant</h2>
          {claudeAgent && (
            <div className="flex items-center gap-2">
              <span className="text-xs px-2 py-1 rounded bg-green-200 text-green-800">
                Connected
              </span>
              <button
                onClick={() => stopMutation.mutate(claudeAgent.id)}
                className="text-xs px-2 py-1 bg-red-500 text-white rounded hover:bg-red-600"
              >
                Disconnect
              </button>
            </div>
          )}
        </div>
        
        {!claudeAgent && agents?.length === 0 && (
          <p className="text-sm text-gray-500 mt-2">
            Send a message to start chatting with Claude
          </p>
        )}
        
        {!claudeAgent && agents?.some((a: Agent) => a.agent_type === 'claude-code' && a.status !== 'running') && (
          <div className="mt-2">
            <p className="text-sm text-yellow-600 mb-2">
              Claude is not running. Would you like to restart?
            </p>
            <button
              onClick={() => {
                const stoppedAgent = agents.find((a: Agent) => 
                  a.agent_type === 'claude-code' && a.status !== 'running'
                )
                if (stoppedAgent) {
                  restartMutation.mutate(stoppedAgent.id)
                }
              }}
              className="text-sm px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600"
            >
              Restart Claude
            </button>
          </div>
        )}
      </div>

      <div className="flex-1 overflow-hidden">
        {claudeAgent ? (
          <ClaudeChat agent={claudeAgent} />
        ) : (
          <div className="h-full flex items-center justify-center text-gray-500">
            <div className="text-center">
              <p className="text-lg mb-2">Claude is not connected</p>
              <p className="text-sm">Create a session to start chatting</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}