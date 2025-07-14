import { useAppStore } from '../store'
import { ClaudeChat } from './ClaudeChat'
import { TodoList } from './TodoList'

export function AssistantView() {
  const { currentSession } = useAppStore()

  if (!currentSession) {
    return (
      <div className="p-4 text-gray-500">
        Select a session to start chatting
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* Top section with Todo List */}
      <div className="p-4 border-b">
        <div className="max-h-64 overflow-y-auto">
          <h3 className="text-sm font-semibold text-gray-700 mb-2">Claude's Tasks</h3>
          <TodoList />
        </div>
      </div>
      
      {/* Chat area below */}
      <div className="flex-1 overflow-hidden">
        <ClaudeChat />
      </div>
    </div>
  )
}