import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { sessionsApi, projectsApi } from '../api/client'
import { useAppStore } from '../store'
import { Session, CreateSessionRequest } from '../types'
import { useSessionTodos } from '../hooks/useSessionTodos'
import { wsClient } from '../api/websocket'
import { DropdownMenu } from './ui/DropdownMenu'
import { useRunStartupScriptMutation } from '../features/sessions/api/sessionsApi'

// Component to show in-progress task for a session
function SessionInProgressTask({ sessionId }: { sessionId: number }) {
  const { inProgressTask } = useSessionTodos(sessionId)
  
  if (!inProgressTask) return null
  
  return (
    <div className="text-xs text-blue-600 mt-1 flex items-center gap-1">
      <span className="animate-pulse">🔄</span>
      <span className="truncate">{inProgressTask}</span>
    </div>
  )
}

export function SessionManager() {
  const queryClient = useQueryClient()
  const { currentProject, currentSession, setCurrentSession } = useAppStore()
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newSession, setNewSession] = useState({
    session_name: '',
    branch_name: '',
    base_branch: 'main',
  })
  const [showBranchSuggestions, setShowBranchSuggestions] = useState(false)
  const [, setUpdateTrigger] = useState(0)
  const [runStartupScript] = useRunStartupScriptMutation()

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

  const { data: branches } = useQuery({
    queryKey: ['branches', currentProject?.id],
    queryFn: async () => {
      if (!currentProject) return { local: [], remote: [] }
      const response = await projectsApi.getBranches(currentProject.id)
      const data = response.data as any
      if (data && data.data) {
        return data.data
      }
      return response.data
    },
    enabled: !!currentProject && showCreateForm,
  })

  const createMutation = useMutation({
    mutationFn: async (data: CreateSessionRequest) => {
      const response = await sessionsApi.create(data)
      return response.data
    },
    onSuccess: async (response: any) => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      setShowCreateForm(false)
      setNewSession({ session_name: '', branch_name: '', base_branch: 'main' })
      
      // Extract session data from response
      const session = (response as any).data || response
      
      // Set the new session as current
      setCurrentSession(session)
    },
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

  const handleCreateSession = () => {
    if (!currentProject || !newSession.session_name || !newSession.branch_name) return
    
    createMutation.mutate({
      project_id: currentProject.id,
      name: newSession.session_name,
      branch_name: newSession.branch_name,
      base_branch: newSession.base_branch,
    })
  }

  if (!currentProject) {
    return (
      <div className="p-4 text-gray-500">
        Select a project to view sessions
      </div>
    )
  }

  return (
    <div className="p-4">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-semibold">Sessions</h2>
        <button
          onClick={() => setShowCreateForm(!showCreateForm)}
          className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 text-sm"
        >
          New Session
        </button>
      </div>

      {showCreateForm && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <input
            type="text"
            placeholder="Session name"
            value={newSession.session_name}
            onChange={(e) => setNewSession({ ...newSession, session_name: e.target.value })}
            className="w-full p-2 border rounded mb-2"
          />
          <input
            type="text"
            placeholder="Branch name"
            value={newSession.branch_name}
            onChange={(e) => setNewSession({ ...newSession, branch_name: e.target.value })}
            className="w-full p-2 border rounded mb-2"
          />
          <div className="relative mb-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Base Branch
            </label>
            <input
              type="text"
              placeholder="Base branch (e.g., main)"
              value={newSession.base_branch}
              onChange={(e) => setNewSession({ ...newSession, base_branch: e.target.value })}
              onFocus={() => setShowBranchSuggestions(true)}
              onBlur={() => setTimeout(() => setShowBranchSuggestions(false), 200)}
              className="w-full p-2 border rounded"
            />
            {showBranchSuggestions && branches && (
              <div className="absolute z-10 w-full mt-1 bg-white border rounded shadow-lg max-h-48 overflow-y-auto">
                {branches.local.length > 0 && (
                  <>
                    <div className="px-3 py-1 text-xs font-semibold text-gray-500 bg-gray-50">
                      Local Branches
                    </div>
                    {branches.local.map((branch: string) => (
                      <button
                        key={`local-${branch}`}
                        type="button"
                        onClick={() => {
                          setNewSession({ ...newSession, base_branch: branch })
                          setShowBranchSuggestions(false)
                        }}
                        className="w-full text-left px-3 py-2 hover:bg-gray-100 flex items-center gap-2"
                      >
                        <span className="text-blue-600">●</span>
                        {branch}
                      </button>
                    ))}
                  </>
                )}
                {branches.remote.length > 0 && (
                  <>
                    <div className="px-3 py-1 text-xs font-semibold text-gray-500 bg-gray-50">
                      Remote Branches
                    </div>
                    {branches.remote.map((branch: string) => (
                      <button
                        key={`remote-${branch}`}
                        type="button"
                        onClick={() => {
                          setNewSession({ ...newSession, base_branch: branch })
                          setShowBranchSuggestions(false)
                        }}
                        className="w-full text-left px-3 py-2 hover:bg-gray-100 flex items-center gap-2"
                      >
                        <span className="text-green-600">●</span>
                        {branch}
                      </button>
                    ))}
                  </>
                )}
                {branches.local.length === 0 && branches.remote.length === 0 && (
                  <div className="px-3 py-2 text-sm text-gray-500">
                    No branches found
                  </div>
                )}
              </div>
            )}
            <p className="text-xs text-gray-500 mt-1">
              The branch to create your new branch from
            </p>
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleCreateSession}
              disabled={createMutation.isPending}
              className="px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600 text-sm disabled:opacity-50"
            >
              Create
            </button>
            <button
              onClick={() => {
                setShowCreateForm(false)
                setNewSession({ session_name: '', branch_name: '', base_branch: 'main' })
              }}
              className="px-3 py-1 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      ) : (
        <div className="space-y-2">
          {sessions?.map((session: Session) => (
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
                className={`w-full text-left p-3 rounded-lg transition-colors ${
                  currentSession?.id === session.id
                    ? 'bg-green-100 border-green-500 border'
                    : 'bg-gray-50 hover:bg-gray-100 border-gray-200 border'
                }`}
              >
                <div className="flex justify-between items-start pr-8">
                  <div className="flex items-center gap-2">
                    <div>
                      <div className="font-medium">{session.name}</div>
                      <div className="text-sm text-gray-600">{session.branch_name}</div>
                      <SessionInProgressTask sessionId={session.id} />
                    </div>
                    {/* Activity indicator */}
                    {session.activity_status && session.activity_status !== 'idle' && (
                      <div className={`w-2 h-2 rounded-full ${
                        session.activity_status === 'streaming' ? 'bg-yellow-500 animate-pulse' :
                        session.activity_status === 'new' ? 'bg-green-500' :
                        session.activity_status === 'viewed' ? 'bg-blue-500' :
                        'bg-gray-400'
                      }`} title={`Activity: ${session.activity_status}`} />
                    )}
                  </div>
                  <span className={`text-xs px-2 py-1 rounded ${
                    session.status === 'active' ? 'bg-green-200 text-green-800' :
                    session.status === 'paused' ? 'bg-yellow-200 text-yellow-800' :
                    'bg-gray-200 text-gray-800'
                  }`}>
                    {session.status}
                  </span>
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
    </div>
  )
}