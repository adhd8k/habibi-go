import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { sessionsApi } from '../api/client'
import { useAppStore } from '../store'
import { Session } from '../types'
import { useSessionTodos } from '../hooks/useSessionTodos'
import { wsClient } from '../api/websocket'
import { DropdownMenu } from './ui/DropdownMenu'
import { useRunStartupScriptMutation } from '../features/sessions/api/sessionsApi'
import { CreateSessionModal } from './CreateSessionModal'

// Component to show in-progress task for a session
function SessionInProgressTask({ sessionId }: { sessionId: number }) {
  const { inProgressTask } = useSessionTodos(sessionId)

  if (!inProgressTask) return null

  return (
    <p className="text-xs text-blue-600 dark:text-blue-400 mt-1 flex items-center gap-1 overflow-hidden truncate">
      <span className="animate-pulse">🔄</span>
      {inProgressTask.length > 25 ? inProgressTask.slice(0, 25) + '...' : inProgressTask}
    </p>
  )
}

export function SessionManager() {
  const queryClient = useQueryClient()
  const { currentProject, currentSession, setCurrentSession } = useAppStore()
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [, setUpdateTrigger] = useState(0)
  const [runStartupScript] = useRunStartupScriptMutation()
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 5

  // Listen for todo updates to trigger re-renders
  useEffect(() => {
    const handleTodoUpdate = () => {
      setUpdateTrigger(prev => prev + 1)
    }

    wsClient.on('claude_output', handleTodoUpdate)
    return () => {
      wsClient.off('claude_output', handleTodoUpdate)
    }
  }, [])

  const { data: sessions, isLoading } = useQuery({
    queryKey: ['sessions', currentProject?.id],
    queryFn: async () => {
      if (!currentProject) return []
      const response = await sessionsApi.list(currentProject.id)
      // Handle the wrapped response format {data: [...], success: true}
      const data = response.data as any
      if (data && data.data && Array.isArray(data.data)) {
        return data.data
      }
      // Fallback to direct array if API format changes
      return Array.isArray(response.data) ? response.data : []
    },
    enabled: !!currentProject,
  })


  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await sessionsApi.delete(id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      // Clear current session if it was deleted
      if (currentSession?.id === deleteMutation.variables) {
        setCurrentSession(null)
      }
    },
  })

  const openEditorMutation = useMutation({
    mutationFn: async (id: number) => {
      await sessionsApi.openWithEditor(id)
    },
    onSuccess: () => {
      // No need to invalidate queries, just show success feedback
    },
  })


  // Calculate pagination
  const totalPages = Math.ceil((sessions?.length || 0) / itemsPerPage)
  const startIndex = (currentPage - 1) * itemsPerPage
  const endIndex = startIndex + itemsPerPage
  const currentSessions = sessions?.slice(startIndex, endIndex) || []

  if (!currentProject) {
    return (
      <div className="p-4 text-center text-gray-500 dark:text-gray-400">
        <div className="text-sm">
          Select a project above to view sessions
        </div>
      </div>
    )
  }

  return (
    <div className="p-4">
      <div className="flex justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Sessions</h2>
        <button
          onClick={() => setShowCreateModal(true)}
          className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 text-sm"
        >
          New Session
        </button>
      </div>


      {isLoading ? (
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-2"></div>
          <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2"></div>
        </div>
      ) : (
        <div className="space-y-2">
          {currentSessions.map((session: Session) => (
            <div key={session.id} className="relative group">
              <button
                onClick={() => {
                  setCurrentSession(session)
                  // Mark session as viewed when selected
                  if (session.activity_status === 'new') {
                    // Update the session activity status to viewed
                    // This would typically be done via an API call
                    queryClient.invalidateQueries({ queryKey: ['sessions'] })
                  }
                }}
                className={`w-full text-left p-3 rounded-lg transition-colors ${currentSession?.id === session.id
                  ? 'bg-green-100 dark:bg-green-900 border-green-500 border'
                  : 'bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 border-gray-200 dark:border-gray-600 border'
                  }`}
              >
                <div className="flex justify-between items-start pr-8">
                  <div className="flex items-center gap-2">
                    {/* Activity indicator */}
                    {session.activity_status && session.activity_status !== 'idle' && (
                      <div className={`w-2 h-2 rounded-full ${session.activity_status === 'streaming' ? 'bg-yellow-500 animate-pulse' :
                        session.activity_status === 'new' ? 'bg-green-500' :
                          session.activity_status === 'viewed' ? 'bg-blue-500' :
                            'bg-gray-400'
                        }`} title={`Activity: ${session.activity_status}`} />
                    )}
                    <div>
                      <div className="font-medium text-gray-900 dark:text-gray-100">{session.name}</div>
                      <div className="text-sm text-gray-600 dark:text-gray-300">{session.branch_name}</div>
                      <SessionInProgressTask sessionId={session.id} />
                    </div>
                  </div>
                </div>
              </button>
              <div className="absolute top-3 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
                <DropdownMenu
                  items={[
                    {
                      label: 'Open with Editor',
                      onClick: () => openEditorMutation.mutate(session.id),
                      icon: (
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                        </svg>
                      )
                    },
                    {
                      label: 'Run Startup Script',
                      onClick: async () => {
                        try {
                          const result = await runStartupScript(session.id).unwrap()
                          if (result.output) {
                            alert(`Startup script output:\n\n${result.output}`)
                          } else {
                            alert('Startup script executed successfully')
                          }
                        } catch (error: any) {
                          alert(`Failed to run startup script: ${error.message || error}`)
                        }
                      },
                      icon: (
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                        </svg>
                      )
                    },
                    {
                      label: 'Delete Session',
                      onClick: () => {
                        if (confirm(`Are you sure you want to delete session "${session.name}"?`)) {
                          deleteMutation.mutate(session.id)
                        }
                      },
                      danger: true,
                      icon: (
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      )
                    }
                  ]}
                />
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Pagination Controls */}
      {totalPages > 1 && (
        <div className="flex justify-center items-center gap-2 mt-4">
          <button
            onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
            disabled={currentPage === 1}
            className="px-3 py-1 text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-300 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Previous
          </button>
          <span className="text-sm text-gray-600 dark:text-gray-400">
            Page {currentPage} of {totalPages}
          </span>
          <button
            onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
            disabled={currentPage === totalPages}
            className="px-3 py-1 text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-300 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Next
          </button>
        </div>
      )}

      {/* Create Session Modal */}
      {currentProject && (
        <CreateSessionModal
          isOpen={showCreateModal}
          onClose={() => setShowCreateModal(false)}
          projectId={currentProject.id}
          onSuccess={(session) => {
            setCurrentSession(session)
          }}
        />
      )}
    </div>
  )
}
