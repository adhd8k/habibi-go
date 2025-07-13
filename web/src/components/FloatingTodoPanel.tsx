import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { agentsApi } from '../api/client'
import { wsClient } from '../api/websocket'
import { useAppStore } from '../store'
import { Agent } from '../types'

interface Todo {
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  priority: 'high' | 'medium' | 'low'
  id: string
}

export function FloatingTodoPanel() {
  const [isOpen, setIsOpen] = useState(false)
  const [isMinimized, setIsMinimized] = useState(false)
  const [todos, setTodos] = useState<Todo[]>([])
  const [currentAgent, setCurrentAgent] = useState<Agent | null>(null)
  const { currentSession } = useAppStore()

  // Get agents for current session
  const { data: agents } = useQuery({
    queryKey: ['agents', currentSession?.id],
    queryFn: async () => {
      if (!currentSession) return []
      const response = await agentsApi.list(currentSession.id)
      const data = response.data as any
      if (data && data.data && Array.isArray(data.data)) {
        return data.data
      }
      return Array.isArray(response.data) ? response.data : []
    },
    enabled: !!currentSession,
  })

  // Find current Claude agent
  useEffect(() => {
    if (!agents) return
    
    const claudeAgents = agents.filter((agent: Agent) => 
      agent.agent_type === 'claude-code'
    )
    
    const activeAgent = claudeAgents.find((a: Agent) => a.status === 'running') || claudeAgents[0]
    setCurrentAgent(activeAgent || null)
  }, [agents])

  // Load chat history to extract todos
  const { data: historyData } = useQuery({
    queryKey: ['chat-history-todos-floating', currentAgent?.id],
    queryFn: async () => {
      if (!currentAgent) return { messages: [] }
      const response = await agentsApi.chatHistory(currentAgent.id, 100)
      return response.data
    },
    enabled: !!currentAgent && currentAgent.status === 'running',
    refetchInterval: 5000
  })

  // Extract todos from chat history
  useEffect(() => {
    if (!historyData?.messages) return

    const todoWriteCalls = historyData.messages
      .filter((msg: any) => msg.role === 'tool_use' && msg.tool_name === 'TodoWrite')
      .sort((a: any, b: any) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())

    if (todoWriteCalls.length > 0) {
      const latestCall = todoWriteCalls[0]
      try {
        const toolInput = typeof latestCall.tool_input === 'string' 
          ? JSON.parse(latestCall.tool_input) 
          : latestCall.tool_input
        
        if (toolInput?.todos && Array.isArray(toolInput.todos)) {
          setTodos(toolInput.todos)
          // Auto-open panel when todos are first detected
          if (toolInput.todos.length > 0 && !isOpen && todos.length === 0) {
            setIsOpen(true)
          }
        }
      } catch (error) {
        console.error('Failed to parse todo list:', error)
      }
    }
  }, [historyData, isOpen, todos.length])

  // Listen for real-time updates
  useEffect(() => {
    if (!currentAgent) return

    const handleAgentOutput = (message: any) => {
      if (message.agent_id === currentAgent.id && message.data) {
        const data = message.data
        if (data.content_type === 'tool_use' && data.tool_name === 'TodoWrite') {
          try {
            const toolInput = data.tool_input
            if (toolInput?.todos && Array.isArray(toolInput.todos)) {
              setTodos(toolInput.todos)
              // Auto-open panel when todos are updated
              if (toolInput.todos.length > 0 && !isOpen) {
                setIsOpen(true)
              }
            }
          } catch (error) {
            console.error('Failed to parse todo list from WebSocket:', error)
          }
        }
      }
    }

    wsClient.subscribe(currentAgent.id)
    wsClient.on('agent_output', handleAgentOutput)

    return () => {
      wsClient.unsubscribe(currentAgent.id)
      wsClient.off('agent_output')
    }
  }, [currentAgent?.id, isOpen])

  // Always show the button if there's a session, even with no todos
  if (!currentSession) {
    return null
  }

  const inProgressCount = todos.filter(t => t.status === 'in_progress').length
  const pendingCount = todos.filter(t => t.status === 'pending').length
  const completedCount = todos.filter(t => t.status === 'completed').length

  return (
    <>
      {/* Floating button */}
      {!isOpen && (
        <button
          onClick={() => setIsOpen(true)}
          className="fixed bottom-4 right-4 bg-blue-500 hover:bg-blue-600 text-white rounded-full p-3 shadow-lg transition-all hover:scale-110 z-50"
          title="Show Claude's Tasks"
        >
          <div className="flex items-center gap-2">
            <span className="text-xl">📋</span>
            {inProgressCount > 0 && (
              <span className="bg-yellow-400 text-black text-xs rounded-full w-5 h-5 flex items-center justify-center">
                {inProgressCount}
              </span>
            )}
          </div>
        </button>
      )}

      {/* Floating panel */}
      {isOpen && (
        <div 
          className={`fixed bottom-4 right-4 bg-white rounded-lg shadow-2xl border border-gray-200 z-50 transition-all ${
            isMinimized ? 'w-64' : 'w-96 max-h-[600px]'
          }`}
        >
          {/* Header */}
          <div className="flex items-center justify-between p-3 border-b bg-gray-50 rounded-t-lg">
            <div className="flex items-center gap-2">
              <span className="text-lg">📋</span>
              <h3 className="font-semibold">Claude's Tasks</h3>
              <div className="flex gap-1">
                {inProgressCount > 0 && (
                  <span className="bg-blue-100 text-blue-700 text-xs px-2 py-0.5 rounded-full">
                    {inProgressCount} active
                  </span>
                )}
                {pendingCount > 0 && (
                  <span className="bg-gray-100 text-gray-700 text-xs px-2 py-0.5 rounded-full">
                    {pendingCount} pending
                  </span>
                )}
              </div>
            </div>
            <div className="flex gap-1">
              <button
                onClick={() => setIsMinimized(!isMinimized)}
                className="text-gray-500 hover:text-gray-700 p-1"
                title={isMinimized ? "Expand" : "Minimize"}
              >
                {isMinimized ? '▲' : '▼'}
              </button>
              <button
                onClick={() => setIsOpen(false)}
                className="text-gray-500 hover:text-gray-700 p-1"
                title="Close"
              >
                ✕
              </button>
            </div>
          </div>

          {/* Content */}
          {!isMinimized && (
            <div className="p-3 overflow-y-auto max-h-[500px]">
              {todos.length === 0 ? (
                <div className="text-center py-4">
                  <p className="text-gray-500 mb-2">No tasks yet</p>
                  <p className="text-xs text-gray-400">
                    Tasks will appear here when Claude tracks them
                  </p>
                </div>
              ) : (
                <>
              {/* In Progress */}
              {todos.filter(t => t.status === 'in_progress').length > 0 && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-gray-700 mb-2">In Progress</h4>
                  <div className="space-y-2">
                    {todos.filter(t => t.status === 'in_progress').map((todo) => (
                      <div key={todo.id} className="flex items-start gap-2 p-2 bg-blue-50 rounded text-sm">
                        <span>🔄</span>
                        <div className="flex-1">
                          <p className="text-blue-900">{todo.content}</p>
                          <span className={`text-xs ${
                            todo.priority === 'high' ? 'text-red-600' :
                            todo.priority === 'medium' ? 'text-yellow-600' :
                            'text-green-600'
                          }`}>
                            {todo.priority} priority
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Pending */}
              {todos.filter(t => t.status === 'pending').length > 0 && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Pending</h4>
                  <div className="space-y-2">
                    {todos.filter(t => t.status === 'pending').map((todo) => (
                      <div key={todo.id} className="flex items-start gap-2 p-2 bg-gray-50 rounded text-sm">
                        <span>⏳</span>
                        <div className="flex-1">
                          <p className="text-gray-700">{todo.content}</p>
                          <span className={`text-xs ${
                            todo.priority === 'high' ? 'text-red-600' :
                            todo.priority === 'medium' ? 'text-yellow-600' :
                            'text-green-600'
                          }`}>
                            {todo.priority} priority
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Completed */}
              {todos.filter(t => t.status === 'completed').length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-2">
                    Completed ({completedCount})
                  </h4>
                  <div className="space-y-1">
                    {todos.filter(t => t.status === 'completed').slice(0, 3).map((todo) => (
                      <div key={todo.id} className="flex items-start gap-2 p-1 text-sm opacity-60">
                        <span>✅</span>
                        <p className="flex-1 line-through text-gray-600">{todo.content}</p>
                      </div>
                    ))}
                    {completedCount > 3 && (
                      <p className="text-xs text-gray-500 pl-6">
                        +{completedCount - 3} more completed
                      </p>
                    )}
                  </div>
                </div>
              )}
                </>
              )}
            </div>
          )}
        </div>
      )}
    </>
  )
}