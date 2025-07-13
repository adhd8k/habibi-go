import { useQuery } from '@tanstack/react-query'
import { agentsApi } from '../api/client'
import { useAppStore } from '../store'
import { Agent } from '../types'
import { TodoList } from './TodoList'

export function SessionTasks() {
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
      <div className="h-full flex items-center justify-center text-gray-500">
        <div className="text-center">
          <p className="text-lg mb-2">No session selected</p>
          <p className="text-sm">Select a session to view tasks</p>
        </div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-32 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-24"></div>
        </div>
      </div>
    )
  }

  // Find the most recent Claude agent
  const claudeAgents = agents?.filter((agent: Agent) => 
    agent.agent_type === 'claude-code'
  ) || []
  
  // Get the running agent or the most recent one
  const currentAgent = claudeAgents.find((a: Agent) => a.status === 'running') || claudeAgents[0]

  return <TodoList agent={currentAgent} />
}