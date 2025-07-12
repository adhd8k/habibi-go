import { Session, Project } from '../../../shared/types/schemas'
import { LoadingSpinner } from '../../../shared/components/LoadingSpinner'
import { ErrorMessage } from '../../../shared/components/ErrorMessage'
import { SessionListItem } from './SessionListItem'

interface SessionManagerViewProps {
  sessions: Session[]
  currentProject: Project | null
  isLoading: boolean
  error: any
  onCreateSession: () => void
}

export function SessionManagerView({
  sessions,
  currentProject,
  isLoading,
  error,
  onCreateSession,
}: SessionManagerViewProps) {
  if (!currentProject) {
    return (
      <div className="p-4 text-center text-gray-500">
        <p>Select a project to view sessions</p>
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-gray-200">
        <div className="flex justify-between items-center mb-2">
          <h2 className="text-lg font-semibold">Sessions</h2>
          <button
            onClick={onCreateSession}
            className="bg-blue-600 text-white px-3 py-1.5 rounded-md hover:bg-blue-700 text-sm"
          >
            New Session
          </button>
        </div>
        <p className="text-sm text-gray-600">
          Project: {currentProject.name}
        </p>
      </div>

      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <LoadingSpinner className="mt-4" />
        ) : error ? (
          <ErrorMessage
            message={error.message || 'Failed to load sessions'}
            className="m-4"
          />
        ) : sessions.length === 0 ? (
          <div className="p-4 text-center text-gray-500">
            <p>No sessions yet</p>
            <p className="text-sm mt-1">Create a session to get started</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {sessions.map((session) => (
              <SessionListItem key={session.id} session={session} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}