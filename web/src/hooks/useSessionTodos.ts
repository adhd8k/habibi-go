import { useState, useEffect } from 'react'
import { wsClient } from '../api/websocket'

interface Todo {
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  priority: 'high' | 'medium' | 'low'
  id: string
}

// Store todos for all sessions
const sessionTodos = new Map<number, Todo[]>()

export function useSessionTodos(sessionId?: number) {
  const [todos, setTodos] = useState<Todo[]>([])
  const [inProgressTask, setInProgressTask] = useState<string | null>(null)

  useEffect(() => {
    if (!sessionId) return

    // Load from cache if available
    const cachedTodos = sessionTodos.get(sessionId) || []
    setTodos(cachedTodos)
    updateInProgressTask(cachedTodos)

    // Load from API
    const loadTodos = async () => {
      try {
        const response = await fetch(`/api/v1/sessions/${sessionId}/chat`, {
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
                  sessionTodos.set(sessionId, toolInput.todos)
                  setTodos(toolInput.todos)
                  updateInProgressTask(toolInput.todos)
                }
              } catch (error) {
                console.error('Failed to parse todos:', error)
              }
            }
          }
        }
      } catch (error) {
        console.error('Failed to load todos:', error)
      }
    }

    loadTodos()
  }, [sessionId])

  // Listen for real-time updates
  useEffect(() => {
    const handleClaudeOutput = (message: any) => {
      if (message.data && message.data.content_type === 'tool_use' && message.data.tool_name === 'TodoWrite') {
        const msgSessionId = message.data.session_id
        
        try {
          let toolInput = message.data.tool_input
          if (typeof toolInput === 'string') {
            toolInput = JSON.parse(toolInput)
          }
          
          if (toolInput?.todos && Array.isArray(toolInput.todos)) {
            // Update cache for this session
            sessionTodos.set(msgSessionId, toolInput.todos)
            
            // Update state if it's the current session
            if (msgSessionId === sessionId) {
              setTodos(toolInput.todos)
              updateInProgressTask(toolInput.todos)
            }
          }
        } catch (error) {
          console.error('Failed to parse todo update:', error)
        }
      }
    }

    wsClient.on('claude_output', handleClaudeOutput)
    return () => {
      wsClient.off('claude_output', handleClaudeOutput)
    }
  }, [sessionId])

  const updateInProgressTask = (todoList: Todo[]) => {
    const inProgress = todoList.find(t => t.status === 'in_progress')
    setInProgressTask(inProgress?.content || null)
  }

  return { todos, inProgressTask }
}

// Export function to get cached in-progress task for any session
export function getSessionInProgressTask(sessionId: number): string | null {
  const todos = sessionTodos.get(sessionId)
  if (!todos) return null
  
  const inProgress = todos.find(t => t.status === 'in_progress')
  return inProgress?.content || null
}