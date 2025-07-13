import { useQuery } from '@tanstack/react-query'
import { agentsApi } from '../api/client'
import { useAppStore } from '../store'
import { Agent } from '../types'
import { ClaudeChat } from './ClaudeChat'
import { TodoList } from './TodoList'

export function AgentControl() {
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

  if (!currentSession) {
    return (
      <div className="p-4 text-gray-500">
        Select a session to start chatting
      </div>
    )
  }

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

  // Filter Claude agents
  const claudeAgents = agents?.filter((agent: Agent) => 
    agent.agent_type === 'claude-code'
  ) || []
  
  // Always use the first available Claude agent
  const currentAgent = claudeAgents.find((a: Agent) => a.status === 'running') || claudeAgents[0]

  return (
    <div className="h-full flex flex-col">
      {/* Top section with Todo List */}
      <div className="p-4 border-b">
        <div className="max-h-64 overflow-y-auto">
          <h3 className="text-sm font-semibold text-gray-700 mb-2">Claude's Tasks</h3>
          <TodoList agent={currentAgent} />
        </div>
      </div>
      
      {/* Chat area below */}
      <div className="flex-1 overflow-hidden">
        <ClaudeChat agent={currentAgent} />
      </div>
    </div>
  )
}