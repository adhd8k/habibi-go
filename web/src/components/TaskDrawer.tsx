import { RightDrawer } from './ui/RightDrawer'
import { TodoList } from './TodoList'

interface TaskDrawerProps {
  isOpen: boolean
  onClose: () => void
}

export function TaskDrawer({ isOpen, onClose }: TaskDrawerProps) {
  return (
    <RightDrawer
      isOpen={isOpen}
      onClose={onClose}
      title="Claude's Tasks"
    >
      <div className="h-full">
        <div className="text-sm text-gray-600 dark:text-gray-400 mb-4">
          Current task progress for the active session
        </div>
        <TodoList />
      </div>
    </RightDrawer>
  )
}