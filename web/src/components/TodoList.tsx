import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { agentsApi } from '../api/client'
import { wsClient } from '../api/websocket'
import { Agent } from '../types'

interface Todo {
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  priority: 'high' | 'medium' | 'low'
  id: string
}

interface TodoListProps {
  agent: Agent | null
}

export function TodoList({ agent }: TodoListProps) {
  const [todos, setTodos] = useState<Todo[]>([])

  // Load chat history to extract todos
  const { data: historyData } = useQuery({
    queryKey: ['chat-history-todos', agent?.id],
    queryFn: async () => {
      if (!agent) return { messages: [] }
      const response = await agentsApi.chatHistory(agent.id, 100)
      return response.data
    },
    enabled: !!agent && agent.status === 'running',
    refetchInterval: 5000 // Refresh every 5 seconds to catch new todos
  })

  // Extract todos from tool calls in chat history
  useEffect(() => {
    if (!historyData?.messages) return

    // Find the most recent TodoWrite tool call
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
        }
      } catch (error) {
        console.error('Failed to parse todo list:', error)
      }
    }
  }, [historyData])

  // Listen for real-time TodoWrite tool calls
  useEffect(() => {
    if (!agent) return

    const handleAgentOutput = (message: any) => {
      if (message.agent_id === agent.id && message.data) {
        const data = message.data
        if (data.content_type === 'tool_use' && data.tool_name === 'TodoWrite') {
          try {
            const toolInput = data.tool_input
            if (toolInput?.todos && Array.isArray(toolInput.todos)) {
              setTodos(toolInput.todos)
            }
          } catch (error) {
            console.error('Failed to parse todo list from WebSocket:', error)
          }
        }
      }
    }

    // Subscribe to agent output
    wsClient.subscribe(agent.id)
    wsClient.on('agent_output', handleAgentOutput)

    return () => {
      wsClient.unsubscribe(agent.id)
      wsClient.off('agent_output')
    }
  }, [agent?.id])


  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return '✅'
      case 'in_progress': return '🔄'
      case 'pending': return '⏳'
      default: return '❓'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'text-green-700'
      case 'in_progress': return 'text-blue-700'
      case 'pending': return 'text-gray-600'
      default: return 'text-gray-500'
    }
  }

  if (!agent) {
    return (
      <div className="text-sm text-gray-500">
        <p>No agent selected</p>
      </div>
    )
  }

  if (todos.length === 0) {
    return (
      <div className="text-sm text-gray-500">
        <p className="mb-1">No tasks yet</p>
        <p className="text-xs text-gray-400">
          Tasks appear when Claude uses TodoWrite
        </p>
      </div>
    )
  }

  // Group todos by status
  const todosByStatus = {
    in_progress: todos.filter(t => t.status === 'in_progress'),
    pending: todos.filter(t => t.status === 'pending'),
    completed: todos.filter(t => t.status === 'completed')
  }

  return (
    <div className="h-full">
      <div>
        
        {/* In Progress */}
        {todosByStatus.in_progress.length > 0 && (
          <div className="mb-3">
            <h4 className="text-xs font-medium text-gray-600 mb-1">In Progress</h4>
            <div className="space-y-1">
              {todosByStatus.in_progress.map((todo) => (
                <div
                  key={todo.id}
                  className="flex items-start gap-2 p-2 bg-blue-50 border border-blue-200 rounded text-xs"
                >
                  <span>{getStatusIcon(todo.status)}</span>
                  <div className="flex-1">
                    <p className={`${getStatusColor(todo.status)}`}>
                      {todo.content}
                    </p>
                    <span className={`text-xs ${
                      todo.priority === 'high' ? 'text-red-600' :
                      todo.priority === 'medium' ? 'text-yellow-600' :
                      'text-green-600'
                    }`}>
                      {todo.priority}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Pending */}
        {todosByStatus.pending.length > 0 && (
          <div className="mb-3">
            <h4 className="text-xs font-medium text-gray-600 mb-1">Pending</h4>
            <div className="space-y-1">
              {todosByStatus.pending.map((todo) => (
                <div
                  key={todo.id}
                  className="flex items-start gap-2 p-2 bg-gray-50 border border-gray-200 rounded text-xs"
                >
                  <span className="text-sm">{getStatusIcon(todo.status)}</span>
                  <div className="flex-1">
                    <p className={`${getStatusColor(todo.status)}`}>
                      {todo.content}
                    </p>
                    <span className={`text-xs ${
                      todo.priority === 'high' ? 'text-red-600' :
                      todo.priority === 'medium' ? 'text-yellow-600' :
                      'text-green-600'
                    }`}>
                      {todo.priority}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Completed */}
        {todosByStatus.completed.length > 0 && (
          <div>
            <h4 className="text-xs font-medium text-gray-600 mb-1">
              Completed ({todosByStatus.completed.length})
            </h4>
            <div className="space-y-1">
              {todosByStatus.completed.slice(0, 2).map((todo) => (
                <div
                  key={todo.id}
                  className="flex items-start gap-2 p-1 text-xs opacity-60"
                >
                  <span className="text-sm">{getStatusIcon(todo.status)}</span>
                  <p className={`flex-1 line-through ${getStatusColor(todo.status)}`}>
                    {todo.content}
                  </p>
                </div>
              ))}
              {todosByStatus.completed.length > 2 && (
                <p className="text-xs text-gray-400 pl-6">
                  +{todosByStatus.completed.length - 2} more
                </p>
              )}
            </div>
          </div>
        )}

        {/* Compact Summary */}
        <div className="mt-3 pt-2 border-t border-gray-200">
          <div className="flex gap-3 text-xs text-gray-600">
            <span>
              <span className="font-semibold text-blue-600">{todosByStatus.in_progress.length}</span> active
            </span>
            <span>
              <span className="font-semibold text-gray-600">{todosByStatus.pending.length}</span> pending
            </span>
            <span>
              <span className="font-semibold text-green-600">{todosByStatus.completed.length}</span> done
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}