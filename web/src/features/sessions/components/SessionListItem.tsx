import { useAppDispatch, useAppSelector } from '../../../app/hooks'
import { Session } from '../../../shared/types/schemas'
import { setCurrentSession, selectCurrentSession } from '../slice/sessionsSlice'
import { useDeleteSessionMutation } from '../api/sessionsApi'

interface SessionListItemProps {
  session: Session
}

export function SessionListItem({ session }: SessionListItemProps) {
  const dispatch = useAppDispatch()
  const currentSession = useAppSelector(selectCurrentSession)
  const [deleteSession] = useDeleteSessionMutation()

  const isActive = currentSession?.id === session.id
  const activityClass = session.activity_status === 'new' 
    ? 'bg-blue-500' 
    : session.activity_status === 'streaming'
    ? 'bg-green-500 animate-pulse'
    : ''

  const handleSelect = () => {
    dispatch(setCurrentSession(session))
  }

  const handleDelete = async (e: React.MouseEvent) => {
    e.stopPropagation()
    
    if (confirm(`Delete session "${session.name}"? This action cannot be undone.`)) {
      try {
        await deleteSession(session.id).unwrap()
      } catch (error) {
        console.error('Failed to delete session:', error)
      }
    }
  }

  return (
    <div
      className={`p-3 hover:bg-gray-50 cursor-pointer transition-colors ${
        isActive ? 'bg-blue-50 border-l-4 border-blue-600' : ''
      }`}
      onClick={handleSelect}
    >
      <div className="flex justify-between items-start">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-medium text-sm truncate">{session.name}</h3>
            {activityClass && (
              <div className={`w-2 h-2 rounded-full ${activityClass}`} />
            )}
          </div>
          <p className="text-xs text-gray-600 truncate">{session.branch_name}</p>
          <p className="text-xs text-gray-500 mt-1">
            {session.status} • {new Date(session.last_used_at).toLocaleDateString()}
          </p>
        </div>
        <button
          onClick={handleDelete}
          className="ml-2 text-gray-400 hover:text-red-600 transition-colors"
          title="Delete session"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </div>
  )
}