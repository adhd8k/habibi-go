import { useAppStore } from '../../../store'
import { ClaudeChat } from './ClaudeChat'

export function AssistantView() {
  const { currentSession } = useAppStore()

  if (!currentSession) {
    return (
      <div className="p-4 text-gray-500 dark:text-gray-400">
        Select a session to start chatting
      </div>
    )
  }

  return (
    <div className="h-full w-full">
      <ClaudeChat />
    </div>
  )
}
