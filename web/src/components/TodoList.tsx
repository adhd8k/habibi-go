import { useState, useEffect } from 'react'
import { wsClient } from '../api/websocket'
import { useAppStore } from '../store'

interface Todo {
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  priority: 'high' | 'medium' | 'low'
  id: string
}

export function TodoList() {
  const [todos, setTodos] = useState<Todo[]>([])
  const { currentSession } = useAppStore()
  
  console.log('TodoList component rendered, currentSession:', currentSession?.id)

  // Load existing todos from chat history
  useEffect(() => {
    if (!currentSession) return

    const loadExistingTodos = async () => {
      try {
        const response = await fetch(`/api/v1/sessions/${currentSession.id}/chat`, {
          headers: {
            'Authorization': 'Basic ' + btoa('moe:jay')
          }
        })
        if (response.ok) {
          const data = await response.json()
          if (data.data && Array.isArray(data.data)) {
            // Find the most recent TodoWrite tool use
            const todoMessages = data.data
              .filter((msg: any) => msg.role === 'tool_use' && msg.tool_name === 'TodoWrite')
              .sort((a: any, b: any) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
            
            if (todoMessages.length > 0) {
              const latestTodos = todoMessages[0]
              try {
                const toolInput = typeof latestTodos.tool_input === 'string' 
                  ? JSON.parse(latestTodos.tool_input) 
                  : latestTodos.tool_input
                if (toolInput?.todos && Array.isArray(toolInput.todos)) {
                  console.log('Loaded existing todos from chat history:', toolInput.todos)
                  setTodos(toolInput.todos)
                }
              } catch (error) {
                console.error('Failed to parse existing todos:', error)
              }
            }
          }
        }
      } catch (error) {
        console.error('Failed to load chat history for todos:', error)
      }
    }

    loadExistingTodos()
  }, [currentSession?.id])

  // Listen for real-time TodoWrite tool calls - separate effect to avoid re-registration issues
  useEffect(() => {
    const handleClaudeOutput = (message: any) => {
      console.log('TodoList received claude_output:', message)
      
      // Check if this message is for any session (not just current)
      if (message.data) {
        const data = message.data
        console.log('TodoList checking message - session_id:', data.session_id, 'content_type:', data.content_type, 'tool_name:', data.tool_name)
        
        // Only process if it's for the current session
        if (data.session_id === currentSession?.id) {
          if (data.content_type === 'tool_use' && data.tool_name === 'TodoWrite') {
            console.log('TodoWrite detected! Tool input:', data.tool_input)
            try {
              let toolInput = data.tool_input
              
              // If tool_input is a string, try to parse it
              if (typeof toolInput === 'string') {
                console.log('Parsing tool_input string:', toolInput)
                toolInput = JSON.parse(toolInput)
              }
              
              // Check if toolInput.todos exists and is an array
              if (toolInput && toolInput.todos && Array.isArray(toolInput.todos)) {
                console.log('Updating todos from WebSocket:', toolInput.todos)
                setTodos(toolInput.todos)
              } else {
                console.warn('TodoWrite tool_input missing todos array:', toolInput)
              }
            } catch (error) {
              console.error('Failed to parse todo list from WebSocket:', error)
            }
          }
        }
      }
    }

    console.log('TodoList registering WebSocket handler')
    wsClient.on('claude_output', handleClaudeOutput)

    return () => {
      console.log('TodoList unregistering WebSocket handler')
      wsClient.off('claude_output', handleClaudeOutput)
    }
  }, [currentSession?.id])

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

  if (!currentSession) {
    return (
      <div className="text-sm text-gray-500">
        <p>No session selected</p>
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