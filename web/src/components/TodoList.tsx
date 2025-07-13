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

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'text-red-600 bg-red-50 border-red-200'
      case 'medium': return 'text-yellow-600 bg-yellow-50 border-yellow-200'
      case 'low': return 'text-green-600 bg-green-50 border-green-200'
      default: return 'text-gray-600 bg-gray-50 border-gray-200'
    }
  }

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
      <div className="p-4 text-gray-500 text-center">
        <p>No agent selected</p>
      </div>
    )
  }

  if (todos.length === 0) {
    return (
      <div className="p-4 text-gray-500 text-center">
        <p className="text-lg mb-2">No tasks yet</p>
        <p className="text-sm">Tasks will appear here when Claude creates a todo list</p>
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
    <div className="h-full overflow-y-auto">
      <div className="p-4">
        <h3 className="text-lg font-semibold mb-4">Claude's Task List</h3>
        
        {/* In Progress */}
        {todosByStatus.in_progress.length > 0 && (
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-700 mb-2">In Progress</h4>
            <div className="space-y-2">
              {todosByStatus.in_progress.map((todo) => (
                <div
                  key={todo.id}
                  className="flex items-start gap-3 p-3 bg-blue-50 border border-blue-200 rounded-lg"
                >
                  <span className="text-lg">{getStatusIcon(todo.status)}</span>
                  <div className="flex-1">
                    <p className={`font-medium ${getStatusColor(todo.status)}`}>
                      {todo.content}
                    </p>
                    <span className={`inline-block mt-1 text-xs px-2 py-1 rounded-full border ${getPriorityColor(todo.priority)}`}>
                      {todo.priority} priority
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Pending */}
        {todosByStatus.pending.length > 0 && (
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-700 mb-2">Pending</h4>
            <div className="space-y-2">
              {todosByStatus.pending.map((todo) => (
                <div
                  key={todo.id}
                  className="flex items-start gap-3 p-3 bg-gray-50 border border-gray-200 rounded-lg"
                >
                  <span className="text-lg">{getStatusIcon(todo.status)}</span>
                  <div className="flex-1">
                    <p className={`${getStatusColor(todo.status)}`}>
                      {todo.content}
                    </p>
                    <span className={`inline-block mt-1 text-xs px-2 py-1 rounded-full border ${getPriorityColor(todo.priority)}`}>
                      {todo.priority} priority
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Completed */}
        {todosByStatus.completed.length > 0 && (
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-700 mb-2">Completed</h4>
            <div className="space-y-2">
              {todosByStatus.completed.map((todo) => (
                <div
                  key={todo.id}
                  className="flex items-start gap-3 p-3 bg-green-50 border border-green-200 rounded-lg opacity-75"
                >
                  <span className="text-lg">{getStatusIcon(todo.status)}</span>
                  <div className="flex-1">
                    <p className={`line-through ${getStatusColor(todo.status)}`}>
                      {todo.content}
                    </p>
                    <span className={`inline-block mt-1 text-xs px-2 py-1 rounded-full border ${getPriorityColor(todo.priority)}`}>
                      {todo.priority} priority
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Summary */}
        <div className="mt-6 pt-4 border-t border-gray-200">
          <div className="flex justify-around text-sm text-gray-600">
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">{todosByStatus.in_progress.length}</div>
              <div>In Progress</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-gray-600">{todosByStatus.pending.length}</div>
              <div>Pending</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">{todosByStatus.completed.length}</div>
              <div>Completed</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}