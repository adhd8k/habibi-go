import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { agentsApi } from '../api/client'
import { wsClient } from '../api/websocket'
import { useAppStore } from '../store'
import { Agent, CreateAgentRequest } from '../types'

export function AgentControl() {
  const queryClient = useQueryClient()
  const { currentSession, currentAgent, setCurrentAgent } = useAppStore()
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [command, setCommand] = useState('')
  const [output, setOutput] = useState<string[]>([])
  const [newAgent, setNewAgent] = useState({
    agent_type: '',
    command: '',
  })

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

  const createMutation = useMutation({
    mutationFn: async (data: CreateAgentRequest) => {
      const response = await agentsApi.create(data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
      setShowCreateForm(false)
      setNewAgent({ agent_type: '', command: '' })
    },
  })

  const stopMutation = useMutation({
    mutationFn: async (id: number) => {
      await agentsApi.stop(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] })
    },
  })

  const executeMutation = useMutation({
    mutationFn: async ({ agent_id, command }: { agent_id: number; command: string }) => {
      const response = await agentsApi.execute({ agent_id, command })
      return response.data
    },
  })

  useEffect(() => {
    if (currentAgent) {
      wsClient.subscribe(currentAgent.id)
      
      wsClient.on('agent_output', (data) => {
        if (data.agent_id === currentAgent.id) {
          setOutput((prev) => [...prev, data.output])
        }
      })

      wsClient.on('agent_status', (data) => {
        if (data.agent_id === currentAgent.id) {
          queryClient.invalidateQueries({ queryKey: ['agents'] })
        }
      })

      return () => {
        wsClient.unsubscribe(currentAgent.id)
        wsClient.off('agent_output')
        wsClient.off('agent_status')
      }
    }
  }, [currentAgent, queryClient])

  const handleCreateAgent = () => {
    if (!currentSession || !newAgent.agent_type || !newAgent.command) return
    
    createMutation.mutate({
      session_id: currentSession.id,
      agent_type: newAgent.agent_type,
      command: newAgent.command,
    })
  }

  const handleExecuteCommand = () => {
    if (!currentAgent || !command) return
    
    executeMutation.mutate(
      { agent_id: currentAgent.id, command },
      {
        onSuccess: () => {
          setCommand('')
        },
      }
    )
  }

  if (!currentSession) {
    return (
      <div className="p-4 text-gray-500">
        Select a session to view agents
      </div>
    )
  }

  return (
    <div className="p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-semibold">Agents</h2>
        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 text-sm"
        >
          Start Agent
        </button>
      </div>

      {showCreateForm && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <input
            type="text"
            placeholder="Agent type (e.g., claude, gpt-4)"
            value={newAgent.agent_type}
            onChange={(e) => setNewAgent({ ...newAgent, agent_type: e.target.value })}
            className="w-full p-2 border rounded mb-2"
          />
          <input
            type="text"
            placeholder="Command to run"
            value={newAgent.command}
            onChange={(e) => setNewAgent({ ...newAgent, command: e.target.value })}
            className="w-full p-2 border rounded mb-2"
          />
          <div className="flex gap-2">
            <button
              onClick={handleCreateAgent}
              disabled={createMutation.isPending}
              className="px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600 text-sm disabled:opacity-50"
            >
              Start
            </button>
            <button
              onClick={() => {
                setShowCreateForm(false)
                setNewAgent({ agent_type: '', command: '' })
              }}
              className="px-3 py-1 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      ) : (
        <div className="space-y-2">
          {agents?.map((agent: Agent) => (
            <div
              key={agent.id}
              onClick={() => setCurrentAgent(agent)}
              className={`p-3 rounded-lg cursor-pointer transition-colors ${
                currentAgent?.id === agent.id
                  ? 'bg-purple-100 border-purple-500 border'
                  : 'bg-gray-50 hover:bg-gray-100 border-gray-200 border'
              }`}
            >
              <div className="flex justify-between items-start">
                <div>
                  <div className="font-medium">{agent.agent_type}</div>
                  <div className="text-sm text-gray-600">PID: {agent.pid}</div>
                  <div className="text-xs text-gray-500">{agent.command}</div>
                </div>
                <div className="flex items-center gap-2">
                  <span className={`text-xs px-2 py-1 rounded ${
                    agent.status === 'running' ? 'bg-green-200 text-green-800' :
                    agent.status === 'starting' ? 'bg-yellow-200 text-yellow-800' :
                    agent.status === 'failed' ? 'bg-red-200 text-red-800' :
                    'bg-gray-200 text-gray-800'
                  }`}>
                    {agent.status}
                  </span>
                  {agent.status === 'running' && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        stopMutation.mutate(agent.id)
                      }}
                      className="text-xs px-2 py-1 bg-red-500 text-white rounded hover:bg-red-600"
                    >
                      Stop
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {currentAgent && currentAgent.status === 'running' && (
        <div className="mt-6">
          <h3 className="font-semibold mb-2">Execute Command</h3>
          <div className="flex gap-2">
            <input
              type="text"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleExecuteCommand()}
              placeholder="Enter command..."
              className="flex-1 p-2 border rounded"
            />
            <button
              onClick={handleExecuteCommand}
              disabled={executeMutation.isPending}
              className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
            >
              Execute
            </button>
          </div>
          
          {output.length > 0 && (
            <div className="mt-4">
              <h3 className="font-semibold mb-2">Output</h3>
              <div className="bg-gray-900 text-gray-100 p-3 rounded font-mono text-sm max-h-96 overflow-y-auto">
                {output.map((line, i) => (
                  <div key={i}>{line}</div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}