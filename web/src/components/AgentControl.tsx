import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { agentsApi } from '../api/client'
import { useAppStore } from '../store'
import { Agent } from '../types'
import { ClaudeChat } from './ClaudeChat'
import { AgentSelector } from './AgentSelector'

export function AgentControl() {
  const queryClient = useQueryClient()
  const { currentSession } = useAppStore()
  const [selectedAgentId, setSelectedAgentId] = useState<number | null>(null)

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

  // Filter Claude agents
  const claudeAgents = agents?.filter((agent: Agent) => 
    agent.agent_type === 'claude-code'
  ) || []
  
  // Find the selected or active agent
  const selectedAgent = selectedAgentId 
    ? claudeAgents.find((a: Agent) => a.id === selectedAgentId)
    : claudeAgents.find((a: Agent) => a.status === 'running') || claudeAgents[0]
    
  // Auto-select the first running agent or the most recent one
  if (!selectedAgentId && selectedAgent) {
    setSelectedAgentId(selectedAgent.id)
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

  const handleSelectAgent = async (agent: Agent) => {
    setSelectedAgentId(agent.id)
    
    // If agent is stopped, restart it
    if (agent.status !== 'running') {
      await restartMutation.mutateAsync(agent.id)
    }
  }
  
  const handleCreateNewAgent = () => {
    // This will trigger ClaudeChat to create a new agent when sending the first message
    setSelectedAgentId(null)
  }

  return (
    <div className="h-full flex flex-col">
      {claudeAgents.length > 0 && (
        <AgentSelector
          agents={claudeAgents}
          currentAgent={selectedAgent}
          onSelectAgent={handleSelectAgent}
          onCreateNewAgent={handleCreateNewAgent}
        />
      )}
      
      <div className="flex-1 overflow-hidden">
        {selectedAgent && selectedAgent.status === 'running' ? (
          <ClaudeChat agent={selectedAgent} />
        ) : selectedAgent ? (
          <div className="h-full flex items-center justify-center text-gray-500">
            <div className="text-center">
              <p className="text-lg mb-2">Restarting Claude...</p>
              <p className="text-sm">Please wait a moment</p>
            </div>
          </div>
        ) : (
          <ClaudeChat agent={null} />
        )}
      </div>
    </div>
  )
}